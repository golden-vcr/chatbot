# chatbot

The **chatbot** service is responsible for maintaining an IRC client connection in the
Golden VCR Twitch channel, serving as a bot that can respond to events and process user
commands.

- **OpenAPI specification:** https://golden-vcr.github.io/chatbot/

## Development Guide

On a Linux or WSL system:

1. Install [Go 1.21](https://go.dev/doc/install)
2. Clone the [**terraform**](https://github.com/golden-vcr/terraform) repo alongside
   this one, and from the root of that repo:
    - Ensure that the module is initialized (via `terraform init`)
    - Ensure that valid terraform state is present
    - Run `terraform output -raw env_chatbot_local > ../chatbot/.env` to populate an
      `.env` file.
3. Ensure that the [**auth**](https://github.com/golden-vcr/auth?tab=readme-ov-file#development-guide)
   server is running locally.
4. From the root of this repository:
    - Run [`go run ./cmd/server`](./cmd/server/main.go) to start up the server.

Once done, the tapes server will be running at http://localhost:5006.

## Connecting the bot to IRC

In order to connect to IRC using our chat bot's Twitch account (`TapeBoy`), we must
connect our Twitch account to our Chat Bot Twitch Application, granting us a user access
token that we can use to authenticate against the Twitch IRC server.

The access token and refresh token are stored persistently so that login only needs to
occur once. To authenticate the chat bot:

1. Open a browser session and log in to [twitch.tv](https://twitch.tv) using the bot
   account
2. Visit the `/login` route to be redirected to an OAuth consent screen:
    - When running locally: http://localhost:5006/login
    - In a live deployment: https://goldenvcr.com/api/chatbot/login
3. Grant the chat bot application access to the bot account

If `GET /status` shows `connected`, the chat bot has successfully connected the IRC
server and joined the channel.
