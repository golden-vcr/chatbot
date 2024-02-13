package main

import (
	"os"

	"github.com/codingconcepts/env"
	"github.com/golden-vcr/auth"
	"github.com/golden-vcr/chatbot/internal/chatlog"
	"github.com/golden-vcr/chatbot/internal/connection"
	"github.com/golden-vcr/chatbot/internal/irc"
	"github.com/golden-vcr/chatbot/internal/state"
	"github.com/golden-vcr/chatbot/internal/tokens"
	"github.com/golden-vcr/server-common/entry"
	"github.com/golden-vcr/server-common/rmq"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	BindAddr   string `env:"BIND_ADDR"`
	ListenPort uint16 `env:"LISTEN_PORT" default:"5006"`
	PublicUrl  string `env:"PUBLIC_URL" default:"https://goldenvcr.com/api/chatbot"`

	TwitchChannelName  string `env:"TWITCH_CHANNEL_NAME" required:"true"`
	TwitchBotUsername  string `env:"TWITCH_BOT_USERNAME" required:"true"`
	TwitchClientId     string `env:"TWITCH_BOT_CLIENT_ID" required:"true"`
	TwitchClientSecret string `env:"TWITCH_BOT_CLIENT_SECRET" required:"true"`

	TokenStoragePath string `env:"TOKEN_STORAGE_PATH" default:"twitch-tokens"`

	AuthURL          string `env:"AUTH_URL" default:"http://localhost:5002"`
	AuthSharedSecret string `env:"AUTH_SHARED_SECRET" required:"true"`

	RmqHost     string `env:"RMQ_HOST" required:"true"`
	RmqPort     int    `env:"RMQ_PORT" required:"true"`
	RmqVhost    string `env:"RMQ_VHOST" required:"true"`
	RmqUser     string `env:"RMQ_USER" required:"true"`
	RmqPassword string `env:"RMQ_PASSWORD" required:"true"`
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

	// Initialize an AMQP client
	amqpConn, err := amqp.Dial(rmq.FormatConnectionString(config.RmqHost, config.RmqPort, config.RmqVhost, config.RmqUser, config.RmqPassword))
	if err != nil {
		app.Fail("Failed to connect to AMQP server", err)
	}
	defer amqpConn.Close()

	// Prepare a producer that we can use to send messages to the twitch-events queue in
	// response to incoming IRC messages that we need to respond to elsewhere in the
	// platform
	twitchEventsProducer, err := rmq.NewProducer(amqpConn, "twitch-events")
	if err != nil {
		app.Fail("Failed to initialize AMQP producer for twitch-events", err)
	}

	// We need an auth service client so that when a user sends a command that requires
	// accessing their backend state (e.g. '!balance'), we can request a JWT that will
	// authorize those requests
	authServiceClient := auth.NewServiceClient(config.AuthURL, config.AuthSharedSecret)

	// We need an auth client in order to authorize HTTP requests that require
	// admin-level access
	authClient, err := auth.NewClient(ctx, config.AuthURL)
	if err != nil {
		app.Fail("Failed to initialize auth client", err)
	}

	// Start setting up our HTTP handlers, using gorilla/mux for routing
	r := mux.NewRouter()

	// Establish a channel into which new IRC messages will be written as they're
	// received by the current bot
	messagesChan := make(chan *irc.Message)

	// The chatlog server buffers a subset of messages that have appeared recently in
	// the channel, and it serves that stream of messages to clients for rendering
	chatlogServer := chatlog.NewServer(ctx, app.Log(), messagesChan)
	chatlogServer.RegisterRoutes(ctx, r)

	// Initialize an "agent", which is essentially a wrapper for the IRC bot that
	// maintains exactly one connection at a time, and which can respond to successful
	// logins by tearing down any existing connection and then initializing a new one
	// and reconnecting the bot
	agent := state.NewAgent(ctx, app.Log(), config.TwitchChannelName, config.TwitchBotUsername, messagesChan, chatlogServer.EmitBotMessage, authServiceClient, twitchEventsProducer)

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
		connectionServer.RegisterRoutes(authClient, r)
	}

	// Handle incoming HTTP connections until our top-level context is canceled, at
	// which point shut down cleanly
	entry.RunServer(ctx, app.Log(), r, config.BindAddr, config.ListenPort)
}
