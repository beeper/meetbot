package meetbot

import (
	"context"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (m *Meetbot) replyTo(ctx context.Context, roomID id.RoomID, eventID id.EventID, text string) {
	m.client.SendMessageEvent(ctx, roomID, event.EventMessage, &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    text,
		RelatesTo: &event.RelatesTo{
			InReplyTo: &event.InReplyTo{EventID: eventID},
		},
	})
}
