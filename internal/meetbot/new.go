package meetbot

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/calendar/v3"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
)

func (m *Meetbot) handleNew(ctx context.Context, evt *event.Event) {
	srv, err := m.getCalendarService(ctx, evt.Sender)
	if err != nil {
		m.replyTo(ctx, evt.RoomID, evt.ID, `You are not logged in. Use "!meet login" to log in to Google Calendar`)
		return
	}

	var roomName event.RoomNameEventContent
	m.client.StateEvent(ctx, evt.RoomID, event.StateRoomName, "", &roomName)
	if roomName.Name == "" {
		roomName.Name = evt.RoomID.String()
	}

	msg := evt.Content.AsMessage()
	var attendees []*calendar.EventAttendee
	if msg.FormattedBody != "" {
		mentionParser := format.HTMLParser{
			PillConverter: func(displayname, mxid, eventID string, ctx format.Context) string {
				if len(mxid) > 0 && mxid[0] == '@' {
					if _, ok := ctx.ReturnData["mention"]; !ok {
						ctx.ReturnData["mention"] = []string{}
					}
					ctx.ReturnData["mention"] = append(ctx.ReturnData["mention"].([]string), mxid)
					return mxid
				}
				return displayname
			},
		}
		formatContext := format.NewContext(ctx)
		mentionParser.Parse(msg.FormattedBody, formatContext)
		for _, mention := range formatContext.ReturnData["mention"].([]string) {
			if mention == m.config.Username.String() {
				continue
			}
			userID := id.UserID(mention)
			if email, found := m.config.UserIDToEmail[userID]; found {
				attendees = append(attendees, &calendar.EventAttendee{
					Email: email,
				})
			}
		}
	}

	now := time.UnixMilli(evt.Timestamp)
	conferenceRequestID := evt.ID.String()
	meetEvent := &calendar.Event{
		Summary:     fmt.Sprintf("Meeting for %s", roomName.Name),
		Description: "This is a meeting created by Meetbot since there's no direct way to create Google Meet meetings.",
		Attendees:   attendees,
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
		m.replyTo(ctx, evt.RoomID, evt.ID, "Error creating event")
		return
	}
	log.Info().Interface("event", meetEvent).Msg("Created event")

	for meetEvent.ConferenceData.CreateRequest.Status.StatusCode != "success" {
		time.Sleep(1 * time.Second)
		meetEvent, err = srv.Events.Get("primary", meetEvent.Id).Do()
		if err != nil {
			log.Error().Err(err).Msg("Error getting event")
			m.replyTo(ctx, evt.RoomID, evt.ID, "Error getting event")
			return
		}
	}
	log.Info().Interface("event", meetEvent).Msg("Conference request succeeded")

	m.replyTo(ctx, evt.RoomID, evt.ID, meetEvent.HangoutLink)
}
