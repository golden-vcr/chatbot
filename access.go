package chatbot

// RequiredScopes is the list of all Twitch API scopes that a user must authorize in
// order to serve as the Golden VCR chat bot
var RequiredScopes = []string{
	"chat:read",
	"chat:edit",
}
