# Google Meet Bot

This bot allows you to create Google Meet links using the `!meet` command.

You have to log in to Google Calendar via OAuth2. This is because there is no
direct Google Meet API and instead this bot uses the Google Calendar API to
create events with Google Meet links and then sends the link to the chat.
