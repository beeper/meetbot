package meetbot

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"maunium.net/go/mautrix/event"
)

func (m *Meetbot) handleLogin(ctx context.Context, evt *event.Event) {
	if !strings.Contains(strings.TrimPrefix(evt.Content.AsMessage().Body, "login"), "--force") {
		userTok, err := m.db.GetUserRefreshToken(ctx, evt.Sender)
		if err == nil && userTok != "" {
			m.client.SendText(evt.RoomID, "You are already logged in to Google Calendar")
			return
		}
	}

	stateToken := uuid.New().String()
	m.loginStates[stateToken] = loginState{
		userID:  evt.Sender,
		roomID:  evt.RoomID,
		eventID: evt.ID,
	}
	url := m.oauthCfg.AuthCodeURL(stateToken, oauth2.AccessTypeOffline)
	m.client.SendMessageEvent(evt.RoomID, event.EventMessage, &event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          fmt.Sprintf("Login to Google Calendar: %s", url),
		Format:        event.FormatHTML,
		FormattedBody: fmt.Sprintf(`Click <a href="%s">here</a> to login to Google Calendar`, url),
	})
}

func (m *Meetbot) HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	state := r.FormValue("state")
	defer delete(m.loginStates, state)

	loginState, ok := m.loginStates[state]
	if !ok {
		m.log.Error().Msg("invalid state token")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid state token"))
		return
	}

	tok, err := m.oauthCfg.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		m.log.Error().Msg("failed to exchange code")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("failed to exchange code"))
		return
	}

	err = m.db.SetUserRefreshToken(ctx, loginState.userID, tok.RefreshToken)
	if err != nil {
		m.log.Error().Msg("failed to save user token")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to save user token"))
		return
	}

	m.replyTo(loginState.roomID, loginState.eventID, "Successfully logged in to Google Calendar")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<html>
			<head>
				<title>Meetbot</title>
				<style>
					body {
						text-align: center;
						padding: 2em;
					}
				</style>
			</head>
			<body>
				<h1>Successfully logged in to Google Calendar</h1>
				<p>You can now <a href="javascript:window.close()">close</a> this window.</p>
			</body>
		</html>
	`))
}
