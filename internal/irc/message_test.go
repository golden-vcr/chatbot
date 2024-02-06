package irc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseMessage(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantErr string
		want    *Message
	}{
		{
			"empty string",
			"",
			"no message type token found",
			nil,
		},
		{
			"invalid message type",
			":tmi.twitch.tv FOO-bad123",
			"message type is neither all-caps nor a three-digit numeric string",
			nil,
		},
		{
			"basic contrived message",
			"@foo=bar;baz=42 :tmi.twitch.tv 989 foo bar :something else\r",
			"",
			&Message{
				Raw: "@foo=bar;baz=42 :tmi.twitch.tv 989 foo bar :something else",
				Extra: map[string]string{
					"foo": "bar",
					"baz": "42",
				},
				Prefix: "tmi.twitch.tv",
				Type:   "989",
				Params: []string{
					"foo",
					"bar",
				},
				Body: "something else",
			},
		},
		{
			"CAP REQ",
			"CAP REQ :twitch.tv/commands twitch.tv/tags\r",
			"",
			&Message{
				Raw:    "CAP REQ :twitch.tv/commands twitch.tv/tags",
				Extra:  map[string]string{},
				Prefix: "",
				Type:   "CAP",
				Params: []string{
					"REQ",
				},
				Body: "twitch.tv/commands twitch.tv/tags",
			},
		},
		{
			"CAP ACK",
			":tmi.twitch.tv CAP * ACK :twitch.tv/commands twitch.tv/tags\r",
			"",
			&Message{
				Raw:    ":tmi.twitch.tv CAP * ACK :twitch.tv/commands twitch.tv/tags",
				Extra:  map[string]string{},
				Prefix: "tmi.twitch.tv",
				Type:   "CAP",
				Params: []string{
					"*",
					"ACK",
				},
				Body: "twitch.tv/commands twitch.tv/tags",
			},
		},
		{
			"Welcome message",
			":tmi.twitch.tv 372 tapeboy :You are in a maze of twisty passages, all alike.\r",
			"",
			&Message{
				Raw:    ":tmi.twitch.tv 372 tapeboy :You are in a maze of twisty passages, all alike.",
				Extra:  map[string]string{},
				Prefix: "tmi.twitch.tv",
				Type:   "372",
				Params: []string{
					"tapeboy",
				},
				Body: "You are in a maze of twisty passages, all alike.",
			},
		},
		{
			"GLOBALUSERSTATE",
			"@badge-info=;badges=;color=#DAA520;display-name=TapeBoy;emote-sets=0;user-id=1001686376;user-type= :tmi.twitch.tv GLOBALUSERSTATE\r",
			"",
			&Message{
				Raw: "@badge-info=;badges=;color=#DAA520;display-name=TapeBoy;emote-sets=0;user-id=1001686376;user-type= :tmi.twitch.tv GLOBALUSERSTATE",
				Extra: map[string]string{
					"badge-info":   "",
					"badges":       "",
					"color":        "#DAA520",
					"display-name": "TapeBoy",
					"emote-sets":   "0",
					"user-id":      "1001686376",
					"user-type":    "",
				},
				Prefix: "tmi.twitch.tv",
				Type:   "GLOBALUSERSTATE",
				Body:   "",
			},
		},
		{
			"USERSTATE",
			"@badge-info=;badges=;color=#DAA520;display-name=TapeBoy;emote-sets=0;mod=0;subscriber=0;user-type= :tmi.twitch.tv USERSTATE #goldenvcr\r",
			"",
			&Message{
				Raw: "@badge-info=;badges=;color=#DAA520;display-name=TapeBoy;emote-sets=0;mod=0;subscriber=0;user-type= :tmi.twitch.tv USERSTATE #goldenvcr",
				Extra: map[string]string{
					"badge-info":   "",
					"badges":       "",
					"color":        "#DAA520",
					"display-name": "TapeBoy",
					"emote-sets":   "0",
					"mod":          "0",
					"subscriber":   "0",
					"user-type":    "",
				},
				Prefix: "tmi.twitch.tv",
				Type:   "USERSTATE",
				Params: []string{
					"#goldenvcr",
				},
				Body: "",
			},
		},
		{
			"ROOMSTATE",
			"@emote-only=0;followers-only=-1;r9k=0;room-id=953753877;slow=0;subs-only=0 :tmi.twitch.tv ROOMSTATE #goldenvcr\r",
			"",
			&Message{
				Raw: "@emote-only=0;followers-only=-1;r9k=0;room-id=953753877;slow=0;subs-only=0 :tmi.twitch.tv ROOMSTATE #goldenvcr",
				Extra: map[string]string{
					"emote-only":     "0",
					"followers-only": "-1",
					"r9k":            "0",
					"room-id":        "953753877",
					"slow":           "0",
					"subs-only":      "0",
				},
				Prefix: "tmi.twitch.tv",
				Type:   "ROOMSTATE",
				Params: []string{
					"#goldenvcr",
				},
				Body: "",
			},
		},
		{
			"PRIVMSG",
			"@badge-info=;badges=;client-nonce=6c15773a6838cdc3bf14000a29259a33;color=#DAA520;display-name=TapeBoy;emotes=305954156:12-19;first-msg=0;flags=;id=a0ce80f0-17eb-4bd9-9115-dbde62d9ff51;mod=0;returning-chatter=0;room-id=953753877;subscriber=0;tmi-sent-ts=1706925305423;turbo=0;user-id=1001686376;user-type= :tapeboy!tapeboy@tapeboy.tmi.twitch.tv PRIVMSG #goldenvcr :Hello world PogChamp\r",
			"",
			&Message{
				Raw: "@badge-info=;badges=;client-nonce=6c15773a6838cdc3bf14000a29259a33;color=#DAA520;display-name=TapeBoy;emotes=305954156:12-19;first-msg=0;flags=;id=a0ce80f0-17eb-4bd9-9115-dbde62d9ff51;mod=0;returning-chatter=0;room-id=953753877;subscriber=0;tmi-sent-ts=1706925305423;turbo=0;user-id=1001686376;user-type= :tapeboy!tapeboy@tapeboy.tmi.twitch.tv PRIVMSG #goldenvcr :Hello world PogChamp",
				Extra: map[string]string{
					"badge-info":        "",
					"badges":            "",
					"client-nonce":      "6c15773a6838cdc3bf14000a29259a33",
					"color":             "#DAA520",
					"display-name":      "TapeBoy",
					"emotes":            "305954156:12-19",
					"first-msg":         "0",
					"flags":             "",
					"id":                "a0ce80f0-17eb-4bd9-9115-dbde62d9ff51",
					"mod":               "0",
					"returning-chatter": "0",
					"room-id":           "953753877",
					"subscriber":        "0",
					"tmi-sent-ts":       "1706925305423",
					"turbo":             "0",
					"user-id":           "1001686376",
					"user-type":         "",
				},
				Prefix: "tapeboy!tapeboy@tapeboy.tmi.twitch.tv",
				Type:   "PRIVMSG",
				Params: []string{
					"#goldenvcr",
				},
				Body: "Hello world PogChamp",
			},
		},
		{
			"PRIVMSG with cheer",
			"@badge-info=subscriber/2;badges=moderator/1,subscriber/0,bits/100;bits=200;color=#D2691E;display-name=TheBellaBunny;emotes=;first-msg=0;flags=;id=8be9e88b-4bdc-4deb-916b-2d40a4299e3f;mod=1;returning-chatter=0;room-id=953753877;subscriber=1;tmi-sent-ts=1707152232192;turbo=0;user-id=230460108;user-type=mod :thebellabunny!thebellabunny@thebellabunny.tmi.twitch.tv PRIVMSG #goldenvcr :Cheer100 Cheer100 ghost of a tiny man wearing a large bowler hat drinking from a penguin shaped glass\r",
			"",
			&Message{
				Raw: "@badge-info=subscriber/2;badges=moderator/1,subscriber/0,bits/100;bits=200;color=#D2691E;display-name=TheBellaBunny;emotes=;first-msg=0;flags=;id=8be9e88b-4bdc-4deb-916b-2d40a4299e3f;mod=1;returning-chatter=0;room-id=953753877;subscriber=1;tmi-sent-ts=1707152232192;turbo=0;user-id=230460108;user-type=mod :thebellabunny!thebellabunny@thebellabunny.tmi.twitch.tv PRIVMSG #goldenvcr :Cheer100 Cheer100 ghost of a tiny man wearing a large bowler hat drinking from a penguin shaped glass",
				Extra: map[string]string{
					"badge-info":        "subscriber/2",
					"badges":            "moderator/1,subscriber/0,bits/100",
					"bits":              "200",
					"color":             "#D2691E",
					"display-name":      "TheBellaBunny",
					"emotes":            "",
					"first-msg":         "0",
					"flags":             "",
					"id":                "8be9e88b-4bdc-4deb-916b-2d40a4299e3f",
					"mod":               "1",
					"returning-chatter": "0",
					"room-id":           "953753877",
					"subscriber":        "1",
					"tmi-sent-ts":       "1707152232192",
					"turbo":             "0",
					"user-id":           "230460108",
					"user-type":         "mod",
				},
				Prefix: "thebellabunny!thebellabunny@thebellabunny.tmi.twitch.tv",
				Type:   "PRIVMSG",
				Params: []string{
					"#goldenvcr",
				},
				Body: "Cheer100 Cheer100 ghost of a tiny man wearing a large bowler hat drinking from a penguin shaped glass",
			},
		},
		{
			"CLEARMSG",
			"@login=wasabimilkshake;room-id=953753877;target-msg-id=4e569024-f783-4218-9fe2-531d8c4d3556;tmi-sent-ts=1707185570996 :tmi.twitch.tv CLEARMSG #goldenvcr :wasabi test 2\r",
			"",
			&Message{
				Raw: "@login=wasabimilkshake;room-id=953753877;target-msg-id=4e569024-f783-4218-9fe2-531d8c4d3556;tmi-sent-ts=1707185570996 :tmi.twitch.tv CLEARMSG #goldenvcr :wasabi test 2",
				Extra: map[string]string{
					"login":         "wasabimilkshake",
					"room-id":       "953753877",
					"target-msg-id": "4e569024-f783-4218-9fe2-531d8c4d3556",
					"tmi-sent-ts":   "1707185570996",
				},
				Prefix: "tmi.twitch.tv",
				Type:   "CLEARMSG",
				Params: []string{
					"#goldenvcr",
				},
				Body: "wasabi test 2",
			},
		},
		{
			"CLEARCHAT with target user",
			"@room-id=953753877;target-user-id=90790024;tmi-sent-ts=1707185640947 :tmi.twitch.tv CLEARCHAT #goldenvcr :wasabimilkshake\r",
			"",
			&Message{
				Raw: "@room-id=953753877;target-user-id=90790024;tmi-sent-ts=1707185640947 :tmi.twitch.tv CLEARCHAT #goldenvcr :wasabimilkshake",
				Extra: map[string]string{
					"room-id":        "953753877",
					"target-user-id": "90790024",
					"tmi-sent-ts":    "1707185640947",
				},
				Prefix: "tmi.twitch.tv",
				Type:   "CLEARCHAT",
				Params: []string{
					"#goldenvcr",
				},
				Body: "wasabimilkshake",
			},
		},
		{
			"CLEARCHAT for entire log",
			"@room-id=953753877;tmi-sent-ts=1707185704103 :tmi.twitch.tv CLEARCHAT #goldenvcr\r",
			"",
			&Message{
				Raw: "@room-id=953753877;tmi-sent-ts=1707185704103 :tmi.twitch.tv CLEARCHAT #goldenvcr",
				Extra: map[string]string{
					"room-id":     "953753877",
					"tmi-sent-ts": "1707185704103",
				},
				Prefix: "tmi.twitch.tv",
				Type:   "CLEARCHAT",
				Params: []string{
					"#goldenvcr",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMessage(tt.s)
			if tt.wantErr != "" {
				assert.Equal(t, tt.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
