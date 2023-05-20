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
	conferenceRequestID := evt.ID.String()
	meetEvent := &calendar.Event{
		Summary:     fmt.Sprintf("Meeting for %s", roomName.Name),
		Description: "This is a meeting created by Meetbot since there's no direct way to create Google Meet meetings.",
		Start: &calendar.EventDateTime{
			DateTime: now.Format("2006-01-02T15:04:05-07:00"),
		},
		End: &calendar.EventDateTime{
			DateTime: now.Add(30 * time.Minute).Format("2006-01-02T15:04:05-07:00"),
		},
		ConferenceData: &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId:             conferenceRequestID,
				ConferenceSolutionKey: &calendar.ConferenceSolutionKey{Type: "hangoutsMeet"},
			},
		},
		Reminders: &calendar.EventReminders{
			UseDefault:      false,
			ForceSendFields: []string{"Overrides"},
		},
	}

	meetEvent, err = srv.Events.
		Insert("primary", meetEvent).
		Context(ctx).
		ConferenceDataVersion(1). // Make sure the conference actually gets created.
		Do()
	if err != nil {
		log.Error().Err(err).Msg("Error creating event")
		m.replyTo(evt.RoomID, evt.ID, "Error creating event")
		return
	}
	log.Info().Interface("event", meetEvent).Msg("Created event")

	for meetEvent.ConferenceData.CreateRequest.Status.StatusCode != "success" {
		time.Sleep(1 * time.Second)
		meetEvent, err = srv.Events.Get("primary", meetEvent.Id).Do()
		if err != nil {
			log.Error().Err(err).Msg("Error getting event")
			m.replyTo(evt.RoomID, evt.ID, "Error getting event")
			return
		}
	}
	log.Info().Interface("event", meetEvent).Msg("Conference request succeeded")

	m.replyTo(evt.RoomID, evt.ID, meetEvent.HangoutLink)
}
