package meetbot

import (
	"context"

	"maunium.net/go/mautrix/event"
)

func (m *Meetbot) handlePing(ctx context.Context, evt *event.Event) {
	srv, err := m.getCalendarService(ctx, evt.Sender)
	if err != nil {
		m.replyTo(ctx, evt.RoomID, evt.ID, "You are not logged in.")
		return
	}

	_, err = srv.Calendars.Get("primary").Context(ctx).Do()
	if err != nil {
		m.replyTo(ctx, evt.RoomID, evt.ID, "Could not find primary calendar")
		return
	}

	m.replyTo(ctx, evt.RoomID, evt.ID, "You are logged in.")
}
