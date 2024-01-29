package main

import (
	"os"

	"github.com/codingconcepts/env"
	"github.com/golden-vcr/chatbot/internal/connection"
	"github.com/golden-vcr/chatbot/internal/state"
	"github.com/golden-vcr/chatbot/internal/tokens"
	"github.com/golden-vcr/server-common/entry"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Config struct {
	BindAddr   string `env:"BIND_ADDR"`
	ListenPort uint16 `env:"LISTEN_PORT" default:"5004"`
	PublicUrl  string `env:"PUBLIC_URL" default:"http://localhost:5004"`

	TwitchChannelName  string `env:"TWITCH_CHANNEL_NAME" required:"true"`
	TwitchBotUsername  string `env:"TWITCH_BOT_USERNAME" required:"true"`
	TwitchClientId     string `env:"TWITCH_BOT_CLIENT_ID" required:"true"`
	TwitchClientSecret string `env:"TWITCH_BOT_CLIENT_SECRET" required:"true"`

	TokenStoragePath string `env:"TOKEN_STORAGE_PATH" default:"twitch-tokens"`
}

func main() {
	app, ctx := entry.NewApplication("chatbot")
	defer app.Stop()

	// Parse config from environment variables
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		app.Fail("Failed to load .env file", err)
	}
	config := Config{}
	if err := env.Set(&config); err != nil {
		app.Fail("Failed to load config", err)
	}

	// Initialize an "agent", which is essentially a wrapper for the IRC bot that
	// maintains exactly one connection at a time, and which can respond to successful
	// logins by tearing down any existing connection and then initializing a new one
	// and reconnecting the bot
	agent := state.NewAgent(ctx, config.TwitchChannelName, config.TwitchBotUsername)

	// Start setting up our HTTP handlers, using gorilla/mux for routing
	r := mux.NewRouter()

	// The connection server exposes HTTP endpoints related to login and connection
	// management: we can use GET /status to see whether the chat bot is successfully
	// authenticated and connected to IRC, we can use GET /login to redirect a user to
	// Twitch in order to issue a User Access Token, and we can use GET /auth to handle
	// the redirect at the end of that flow, providing the server with the access token
	// so that it can connect to IRC as our chat bot user
	{
		redirectUri := config.PublicUrl + "/auth"
		client, err := connection.NewTwitchClient(ctx, config.TwitchClientId, config.TwitchClientSecret, redirectUri)
		if err != nil {
			app.Fail("Failed to initialize Twitch client for connection server", err)
		}
		tokenStore := tokens.NewStore(config.TokenStoragePath, config.TwitchBotUsername)
		connectionServer := connection.NewServer(ctx, app.Log(), agent, client, config.TwitchClientId, redirectUri, config.TwitchBotUsername, tokenStore)
		connectionServer.RegisterRoutes(r)
	}

	// Handle incoming HTTP connections until our top-level context is canceled, at
	// which point shut down cleanly
	entry.RunServer(ctx, app.Log(), r, config.BindAddr, config.ListenPort)
}
