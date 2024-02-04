package chatlog

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Event(t *testing.T) {
	tests := []struct {
		name   string
		ev     Event
		jsonEv string
	}{
		{
			"append basic message",
			Event{
				Type: EventTypeAppend,
				Payload: &Payload{
					Append: &PayloadAppend{
						MessageId: "f5e05a31-57c8-4f34-bfd5-bc1ae222c279",
						Username:  "wasabimilkshake",
						Color:     "#00ff7f",
						Text:      "Hello world",
						Emotes:    []EmoteDetails{},
					},
				},
			},
			`{"type":"append","payload":{"messageId":"f5e05a31-57c8-4f34-bfd5-bc1ae222c279","username":"wasabimilkshake","color":"#00ff7f","text":"Hello world","emotes":[]}}`,
		},
		{
			"append message with emotes",
			Event{
				Type: EventTypeAppend,
				Payload: &Payload{
					Append: &PayloadAppend{
						MessageId: "f5e05a31-57c8-4f34-bfd5-bc1ae222c279",
						Username:  "wasabimilkshake",
						Color:     "#00ff7f",
						Text:      "I have $$52 $0 $0 $1",
						Emotes: []EmoteDetails{
							{
								Name: "someEmote",
								Url:  "https://my-cool-emotes.biz/some-emote.png",
							},
							{
								Name: "anotherEmote",
								Url:  "https://my-cool-emotes.biz/another-emote.gif",
							},
						},
					},
				},
			},
			`{"type":"append","payload":{"messageId":"f5e05a31-57c8-4f34-bfd5-bc1ae222c279","username":"wasabimilkshake","color":"#00ff7f","text":"I have $$52 $0 $0 $1","emotes":[{"name":"someEmote","url":"https://my-cool-emotes.biz/some-emote.png"},{"name":"anotherEmote","url":"https://my-cool-emotes.biz/another-emote.gif"}]}}`,
		},
		{
			"delete a single message",
			Event{
				Type: EventTypeDelete,
				Payload: &Payload{
					Delete: &PayloadDelete{
						MessageIds: []string{"f5e05a31-57c8-4f34-bfd5-bc1ae222c279"},
					},
				},
			},
			`{"type":"delete","payload":{"messageIds":["f5e05a31-57c8-4f34-bfd5-bc1ae222c279"]}}`,
		},
		{
			"clear the entire log",
			Event{
				Type: EventTypeClear,
			},
			`{"type":"clear"}`,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("marshal %s to JSON", tt.name), func(t *testing.T) {
			want := tt.jsonEv
			got, err := json.Marshal(tt.ev)
			assert.NoError(t, err)
			assert.Equal(t, want, string(got))
		})
		t.Run(fmt.Sprintf("unmarshal %s from JSON", tt.name), func(t *testing.T) {
			want := tt.ev
			var got Event
			err := json.Unmarshal([]byte(tt.jsonEv), &got)
			assert.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}
