package irc

// Message represents a single message received from the Twitch IRC server with both the
// 'twitch.tv/commands' and 'twitch.tv/tags' capabilities enabled
type Message struct {
	// Raw is the original message in its plain-text format, stripped of the trailing
	// newline
	Raw string

	// Extra is a parsed representation of the '@'-prefixed, ';'-delimited string of
	// 'key=value' pairs that may appear before the message, separated from it by a
	// space: e.g. '@foo=bar;x=42;type= :tmi.twitch.tv CLEARCHAT #somechannel' will have
	// an Extra map of {"foo": "bar", "x": "42", "type": ""}
	Extra map[string]string

	// Prefix is the user and/or host information that precedes the data of the message,
	// stripped of the leading colon, e.g. 'tmi.twitch.tv',
	// 'somebody!somebody@somebody.tmi.twitch.tv'
	Prefix string

	// Type is the message type identifier that follows the prefix, e.g. 'USERSTATE',
	// 'CLEARCHAT', 'JOIN', 'PART', 'PING', etc.
	Type string

	// Params is a list of all space-delimited tokens that follow the type, up to (and
	// not including) the first value that's prefixed with a ':'
	Params []string

	// Body is the string that follows the first ':' appearing after the message type
	Body string
}
