package meetbot

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"maunium.net/go/mautrix/event"
)

func (m *Meetbot) handleNew(ctx context.Context, evt *event.Event) {
	refreshToken, err := m.db.GetUserRefreshToken(ctx, evt.Sender)
	if err != nil {
		log.Error().Err(err).Msg("Error getting user refresh token")
		m.replyTo(evt.RoomID, evt.ID, `You are not logged in. Use "!meet login" to log in to Google Calendar`)
		return
	}

	if _, ok := m.services[evt.Sender]; !ok {
		token, err := m.oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
		if err != nil {
			log.Error().Err(err).Msg("Error getting token")
			m.replyTo(evt.RoomID, evt.ID, "Error getting token")
			return
		}
		client := m.oauthCfg.Client(ctx, token)
		srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Error().Err(err).Msg("Error creating calendar service")
			m.replyTo(evt.RoomID, evt.ID, "Error creating calendar service")
			return
		}
		m.services[evt.Sender] = srv
	}

	var roomName event.RoomNameEventContent
	m.client.StateEvent(evt.RoomID, event.StateRoomName, "", &roomName)
	if roomName.Name == "" {
		roomName.Name = evt.RoomID.String()
	}

	srv := m.services[evt.Sender]

	now := time.UnixMilli(evt.Timestamp)
	meetEvent := &calendar.Event{
		Summary: fmt.Sprintf("Meeting for %s", roomName.Name),
		Start: &calendar.EventDateTime{
			DateTime: now.Format("2006-01-02T15:04:05-07:00"),
		},
		End: &calendar.EventDateTime{
			DateTime: now.Add(30 * time.Minute).Format("2006-01-02T15:04:05-07:00"),
		},
		ConferenceData: &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId:             evt.ID.String(),
				ConferenceSolutionKey: &calendar.ConferenceSolutionKey{Type: "hangoutsMeet"},
			},
		},
	}

	meetEvent, err = srv.Events.Insert("primary", meetEvent).Do()
	if err != nil {
		log.Error().Err(err).Msg("Error creating event")
		m.replyTo(evt.RoomID, evt.ID, "Error creating event")
		return
	}
	log.Info().Interface("event", meetEvent).Msg("Created event")
	m.replyTo(evt.RoomID, evt.ID, meetEvent.HangoutLink)
}
