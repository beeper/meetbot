package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mau.fi/util/dbutil"
	_ "go.mau.fi/util/dbutil/litestream"
	"gopkg.in/yaml.v3"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"

	"github.com/beeper/meetbot/internal/config"
	"github.com/beeper/meetbot/internal/meetbot"
)

func main() {
	// Arg parsing
	configPath := flag.String("config", "./config.yaml", "config file location")
	flag.Parse()

	// Load configuration
	log.Info().Str("config_path", *configPath).Msg("Reading config")
	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatal().Err(err).Str("config_path", *configPath).Msg("Failed opening the config")
	}

	var config config.Config
	if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatal().Err(err).Msg("Failed reading the config")
	}

	// Setup logging
	logger, err := config.Logging.Compile()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to compile logging configuration")
	}
	log.Logger = *logger
	zerolog.DefaultContextLogger = logger
	ctx := logger.WithContext(context.TODO())

	log.Info().Msg("Meetbot service starting...")

	// Open the meetbot database
	db, err := dbutil.NewFromConfig("meetbot", config.Database, dbutil.ZeroLogger(log.Logger))
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't open database")
	}

	client, err := mautrix.NewClient(config.Homeserver, "", "")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create matrix client")
	}
	client.Log = *logger

	cryptoHelper, err := cryptohelper.NewCryptoHelper(client, []byte("meetbot_cryptostore_key"), db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create crypto helper")
	}
	password, err := config.GetPassword()
	if err != nil {
		log.Fatal().Err(err).Str("password_file", config.PasswordFile).Msg("Could not read password from file")
	}
	cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: config.Username.String(),
		},
		Password: password,
	}
	cryptoHelper.DBAccountID = config.Username.String()

	err = cryptoHelper.Init(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize crypto helper")
	}
	client.Crypto = cryptoHelper

	// Set the bot's display name and avatar
	if config.Displayname != "" {
		client.SetDisplayName(context.Background(), config.Displayname)
	}
	if config.AvatarURL != "" {
		client.SetAvatarURL(context.Background(), config.AvatarURL.ParseOrIgnore())
	}

	meetbot := meetbot.NewMeetbot(ctx, client, logger, db, config)

	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.StateMember, func(ctx context.Context, evt *event.Event) {
		if evt.StateKey == nil || *evt.StateKey != config.Username.String() {
			return
		}
		if evt.Content.AsMember().Membership == event.MembershipInvite {
			log.Info().Stringer("room_id", evt.RoomID).Msg("Invited to room")
			_, err := client.JoinRoom(ctx, evt.RoomID.String(), "", nil)
			if err != nil {
				log.Err(err).Stringer("room_id", evt.RoomID).Msg("Failed to join room")
			}
		}
	})
	syncer.OnEventType(event.EventMessage, meetbot.HandleMessage)

	syncCtx, cancelSync := context.WithCancel(context.Background())
	var syncStopWait sync.WaitGroup
	syncStopWait.Add(1)

	// Start the sync loop
	go func() {
		log.Debug().Msg("starting sync loop")
		err = client.SyncWithContext(syncCtx)
		defer syncStopWait.Done()
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Fatal().Err(err).Msg("Sync error")
		}
	}()

	// Start the OAuth2 listener
	go func() {
		http.HandleFunc("/oauth2/callback", meetbot.HandleOAuth2Callback)
		err := http.ListenAndServe(config.Listen, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("HTTP listener error")
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Info().Msg("Interrupt received, stopping...")
	cancelSync()
	log.Info().Msg("meetbot stopped")
	os.Exit(0)
}
