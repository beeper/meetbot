package meetbot

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/util/dbutil"

	"github.com/beeper/meetbot/internal/config"
	"github.com/beeper/meetbot/internal/database"
)

type loginState struct {
	userID  id.UserID
	roomID  id.RoomID
	eventID id.EventID
}

type Meetbot struct {
	log    *zerolog.Logger
	client *mautrix.Client
	db     *database.Database

	oauthCfg *oauth2.Config

	loginStates map[string]loginState
	services    map[id.UserID]*calendar.Service
}

func NewMeetbot(client *mautrix.Client, log *zerolog.Logger, db *dbutil.Database, config config.Config) *Meetbot {
	credsJSON, err := config.GetCredentialsJSON()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading Google credentials")
	}
	cfg, err := google.ConfigFromJSON(credsJSON, calendar.CalendarScope)
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading Google credentials")
	}

	wrapped := database.NewDatabase(db)
	if err := wrapped.DB.Upgrade(); err != nil {
		log.Fatal().Err(err).Msg("Error upgrading database")
	}

	return &Meetbot{
		log:         log,
		client:      client,
		db:          wrapped,
		oauthCfg:    cfg,
		loginStates: map[string]loginState{},
		services:    map[id.UserID]*calendar.Service{},
	}
}

func (m *Meetbot) HandleMessage(_ mautrix.EventSource, evt *event.Event) {
	log := log.With().
		Str("event_type", evt.Type.String()).
		Str("sender", evt.Sender.String()).
		Str("room_id", string(evt.RoomID)).
		Str("event_id", string(evt.ID)).
		Logger()
	ctx := log.WithContext(context.TODO())

	if evt.Sender == m.client.UserID {
		log.Debug().Msg("Ignoring own message")
		return
	}

	isDM := false
	members, err := m.client.StateStore.GetRoomJoinedOrInvitedMembers(evt.RoomID)
	if err != nil {
		log.Error().Err(err).Msg("Error getting room members")
	} else if len(members) == 2 {
		isDM = true
	}

	msg := evt.Content.AsMessage()
	if isDM || m.isMention(msg) {
		m.handleCommand(ctx, evt, msg)
	}
}

func (m *Meetbot) isMention(msg *event.MessageEventContent) bool {
	return strings.HasPrefix(msg.Body, "!meet") ||
		strings.HasPrefix(msg.Body, "!meetbot") ||
		strings.HasPrefix(msg.Body, "@meetbot")
}

func (m *Meetbot) handleCommand(ctx context.Context, evt *event.Event, msg *event.MessageEventContent) {
	commandText := strings.TrimPrefix(msg.Body, "!meet")
	commandText = strings.TrimPrefix(commandText, "!meetbot")
	commandText = strings.TrimPrefix(commandText, "@meetbot")
	commandText = strings.TrimSpace(commandText)

	switch commandText {
	case "help":
		m.sendHelp(evt)
	case "login":
		m.handleLogin(ctx, evt)
	case "new", "":
		m.handleNew(ctx, evt)
	}
}
