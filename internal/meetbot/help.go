package meetbot

import "maunium.net/go/mautrix/event"

var (
	helpText = `Usage: !meet [command]

The default command is "new".

Commands:

* help - Show this help text
* login [--force] - Log in to Google Meet
* new - Create a new meeting`
	helpHTML = `<b>Usage:</b> <code>!meet [command]</code><br><br>
The default command is "new".<br><br>
<b>Commands:</b>
<ul>
  <li><code>help</code> - Show this help text</li>
  <li><code>login [--force]</code> - Log in to Google Meet</li>
  <li><code>new</code> - Create a new meeting</li>
</ul>`
)

func (m *Meetbot) sendHelp(evt *event.Event) {
	m.client.SendMessageEvent(evt.RoomID, event.EventMessage, &event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          helpText,
		Format:        event.FormatHTML,
		FormattedBody: helpHTML,
	})
}
