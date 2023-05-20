package meetbot

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/calendar/v3"
	"maunium.net/go/mautrix/event"
)

func (m *Meetbot) handleNew(ctx context.Context, evt *event.Event) {
	srv, err := m.getCalendarService(ctx, evt.Sender)
	if err != nil {
		m.replyTo(evt.RoomID, evt.ID, `You are not logged in. Use "!meet login" to log in to Google Calendar`)
		return
	}

	var roomName event.RoomNameEventContent
	m.client.StateEvent(evt.RoomID, event.StateRoomName, "", &roomName)
	if roomName.Name == "" {
		roomName.Name = evt.RoomID.String()
	}

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
