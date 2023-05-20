package meetbot

import (
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (m *Meetbot) replyTo(roomID id.RoomID, eventID id.EventID, text string) {
	m.client.SendMessageEvent(roomID, event.EventMessage, &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    text,
		RelatesTo: &event.RelatesTo{
			InReplyTo: &event.InReplyTo{EventID: eventID},
		},
	})
}
