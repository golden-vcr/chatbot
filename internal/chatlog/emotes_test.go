package chatlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseEmotes(t *testing.T) {
	tests := []struct {
		name    string
		emotes  string
		wantErr string
		want    []emoteInfo
	}{
		{
			"no emotes",
			"",
			"",
			[]emoteInfo{},
		},
		{
			"single emote",
			"1529862:0-8",
			"",
			[]emoteInfo{
				{
					id: "1529862",
					spans: []emoteSpan{
						{0, 8},
					},
				},
			},
		},
		{
			"multiple emotes, multiple spans",
			"555555584:49-50/emotesv2_9d94d65bbef64763b7c09401156ea0bc:2-15,17-30,34-47",
			"",
			[]emoteInfo{
				{
					id: "555555584",
					spans: []emoteSpan{
						{49, 50},
					},
				},
				{
					id: "emotesv2_9d94d65bbef64763b7c09401156ea0bc",
					spans: []emoteSpan{
						{2, 15},
						{17, 30},
						{34, 47},
					},
				},
			},
		},
		{
			"bad id/span delimiter",
			"555555584 49-50",
			"emote attribute is not delimited by a colon",
			nil,
		},
		{
			"bad start/end delimiter",
			"555555584:49 50",
			"span in emote attribute is not delimited by a hyphen",
			nil,
		},
		{
			"non-numeric span start",
			"555555584:foo-50",
			"span in emote attribute has non-numeric start",
			nil,
		},
		{
			"non-numeric span end",
			"555555584:49-foo",
			"span in emote attribute has non-numeric end",
			nil,
		},
		{
			"bad span delimiter",
			"555555584:49-50 52-53",
			"span in emote attribute has non-numeric end",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEmotes(tt.emotes)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
