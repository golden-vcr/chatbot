package chatlog

import (
	"testing"

	"github.com/golden-vcr/chatbot/internal/irc"
	"github.com/stretchr/testify/assert"
)

func Test_EventFromMessage(t *testing.T) {
	tests := []struct {
		name    string
		message *irc.Message
		wantErr error
		want    *Event
	}{
		{
			"message types not relevant to the chat log are ignored",
			&irc.Message{
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
			ErrIgnored,
			nil,
		},
		{
			"basic PRIVMSG",
			&irc.Message{
				Extra: map[string]string{
					"badge-info":        "",
					"badges":            "bits/100",
					"color":             "#00FF7F",
					"display-name":      "wasabimilkshake",
					"emotes":            "",
					"first-msg":         "0",
					"flags":             "",
					"id":                "ad6d1481-1471-4538-900a-493704fc60c5",
					"mod":               "0",
					"returning-chatter": "0",
					"room-id":           "953753877",
					"subscriber":        "0",
					"tmi-sent-ts":       "1707193714879",
					"turbo":             "0",
					"user-id":           "90790024",
					"user-type":         "",
				},
				Prefix: "wasabimilkshake!wasabimilkshake@wasabimilkshake.tmi.twitch.tv",
				Type:   "PRIVMSG",
				Params: []string{
					"#goldenvcr",
				},
				Body: "hello world, this is a test",
			},
			nil,
			&Event{
				Type: EventTypeAppend,
				Payload: &Payload{
					Append: &PayloadAppend{
						MessageId: "ad6d1481-1471-4538-900a-493704fc60c5",
						UserId:    "90790024",
						Username:  "wasabimilkshake",
						Color:     "#00FF7F",
						Text:      "hello world, this is a test",
						Emotes:    []EmoteDetails{},
					},
				},
			},
		},
		{
			"PRIVMSG with emotes",
			&irc.Message{
				Extra: map[string]string{
					"badge-info":        "",
					"badges":            "bits/100",
					"color":             "#00FF7F",
					"display-name":      "wasabimilkshake",
					"emotes":            "emotesv2_9d94d65bbef64763b7c09401156ea0bc:0-13,52-65/emotesv2_9fa2491b63344c15a7e1e2fea713a6e2:34-50,79-95,97-113",
					"first-msg":         "0",
					"flags":             "",
					"id":                "8921f142-cdab-4d16-ba1e-90a4f5e65f7f",
					"mod":               "0",
					"returning-chatter": "0",
					"room-id":           "953753877",
					"subscriber":        "0",
					"tmi-sent-ts":       "1707194149727",
					"turbo":             "0",
					"user-id":           "90790024",
					"user-type":         "",
				},
				Prefix: "wasabimilkshake!wasabimilkshake@wasabimilkshake.tmi.twitch.tv",
				Type:   "PRIVMSG",
				Params: []string{
					"#goldenvcr",
				},
				Body: "wasabi22Denton I have $1 and this golden1029Sadtape wasabi22Denton is an emote golden1029Sadtape golden1029Sadtape",
			},
			nil,
			&Event{
				Type: EventTypeAppend,
				Payload: &Payload{
					Append: &PayloadAppend{
						MessageId: "8921f142-cdab-4d16-ba1e-90a4f5e65f7f",
						UserId:    "90790024",
						Username:  "wasabimilkshake",
						Color:     "#00FF7F",
						Text:      "$0 I have $$1 and this $1 $0 is an emote $1 $1",
						Emotes: []EmoteDetails{
							{
								Name: "wasabi22Denton",
								Url:  "https://static-cdn.jtvnw.net/emoticons/v2/emotesv2_9d94d65bbef64763b7c09401156ea0bc/default/dark/1.0",
							},
							{
								Name: "golden1029Sadtape",
								Url:  "https://static-cdn.jtvnw.net/emoticons/v2/emotesv2_9fa2491b63344c15a7e1e2fea713a6e2/default/dark/1.0",
							},
						},
					},
				},
			},
		},
		{
			"PRIVMSG with cheer", // cheermote handling NYI
			&irc.Message{
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
			nil,
			&Event{
				Type: EventTypeAppend,
				Payload: &Payload{
					Append: &PayloadAppend{
						MessageId: "8be9e88b-4bdc-4deb-916b-2d40a4299e3f",
						UserId:    "230460108",
						Username:  "TheBellaBunny",
						Color:     "#D2691E",
						Text:      "Cheer100 Cheer100 ghost of a tiny man wearing a large bowler hat drinking from a penguin shaped glass",
						Emotes:    []EmoteDetails{},
					},
				},
			},
		},
		{
			"CLEARMSG",
			&irc.Message{
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
			nil,
			&Event{
				Type: EventTypeDelete,
				Payload: &Payload{
					Delete: &PayloadDelete{
						MessageId: "4e569024-f783-4218-9fe2-531d8c4d3556",
					},
				},
			},
		},
		{
			"CLEARCHAT with target user",
			&irc.Message{
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
			nil,
			&Event{
				Type: EventTypeBan,
				Payload: &Payload{
					Ban: &PayloadBan{
						UserId: "90790024",
					},
				},
			},
		},
		{
			"CLEARCHAT for entire log",
			&irc.Message{
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
			nil,
			&Event{
				Type: EventTypeClear,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EventFromMessage(tt.message)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
