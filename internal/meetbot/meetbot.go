package meetbot

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mau.fi/util/dbutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

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
	config config.Config
	client *mautrix.Client
	db     *database.Database

	oauthCfg *oauth2.Config

	loginStates map[string]loginState
	services    map[id.UserID]*calendar.Service
}

func NewMeetbot(client *mautrix.Client, log *zerolog.Logger, db *dbutil.Database, config config.Config) *Meetbot {
	cfg, err := google.ConfigFromJSON(config.GetCredentialsJSON(), calendar.CalendarScope)
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading Google credentials")
	}
	cfg.RedirectURL = config.GetRedirectURL()

	wrapped := database.NewDatabase(db)
	if err := wrapped.DB.Upgrade(); err != nil {
		log.Fatal().Err(err).Msg("Error upgrading database")
	}

	return &Meetbot{
		log:         log,
		client:      client,
		config:      config,
		db:          wrapped,
		oauthCfg:    cfg,
		loginStates: map[string]loginState{},
		services:    map[id.UserID]*calendar.Service{},
	}
}

func (m *Meetbot) getCalendarService(ctx context.Context, userID id.UserID) (*calendar.Service, error) {
	if service, ok := m.services[userID]; ok {
		return service, nil
	}

	refreshToken, err := m.db.GetUserRefreshToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	token, err := m.oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		return nil, err
	}

	client := m.oauthCfg.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	m.services[userID] = srv
	return srv, nil
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
	mentionPrefix := fmt.Sprintf(`<a href="https://matrix.to/#/%s">`, m.client.UserID)
	m.log.Info().Str("mention_prefix", mentionPrefix).Str("formatted_body", msg.FormattedBody).Msg("Checking for mention")
	return strings.HasPrefix(msg.Body, "!meet") ||
		strings.HasPrefix(msg.Body, "!meetbot") ||
		strings.HasPrefix(msg.Body, "@meetbot") ||
		strings.HasPrefix(msg.Body, fmt.Sprintf("%s:", m.config.Displayname)) ||
		strings.HasPrefix(msg.FormattedBody, fmt.Sprintf(`<a href="https://matrix.to/#/%s">`, m.client.UserID))
}

func (m *Meetbot) handleCommand(ctx context.Context, evt *event.Event, msg *event.MessageEventContent) {
	commandText := strings.TrimPrefix(msg.Body, "!meet")
	commandText = strings.TrimPrefix(commandText, "!meetbot")
	commandText = strings.TrimPrefix(commandText, "@meetbot")
	commandText = strings.TrimPrefix(commandText, m.config.Displayname)
	commandText = strings.TrimPrefix(commandText, ":")
	commandText = strings.TrimSpace(commandText)

	switch commandText {
	case "help":
		m.sendHelp(evt)
	case "login":
		m.handleLogin(ctx, evt)
	case "ping":
		m.handlePing(ctx, evt)
	case "new", "":
		m.handleNew(ctx, evt)
	}
}
