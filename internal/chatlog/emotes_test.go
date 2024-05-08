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

func Test_substituteEmotes(t *testing.T) {
	tests := []struct {
		name             string
		body             string
		emoteInfos       []emoteInfo
		wantErr          string
		wantText         string
		wantEmoteDetails []EmoteDetails
	}{
		{
			"message with no emotes is unchanged",
			"boy howdy, would you look at that",
			[]emoteInfo{},
			"",
			"boy howdy, would you look at that",
			[]EmoteDetails{},
		},
		{
			"literal dollar is still escaped in message with no emotes",
			"I have $5 in my pocket",
			[]emoteInfo{},
			"",
			"I have $$5 in my pocket",
			[]EmoteDetails{},
		},
		{
			"message with a single emote",
			"looky here it's our boy abe presidAbe , howdy doody",
			[]emoteInfo{
				{
					id: "4151865",
					spans: []emoteSpan{
						{28, 36},
					},
				},
			},
			"",
			"looky here it's our boy abe $0 , howdy doody",
			[]EmoteDetails{
				{
					Name: "presidAbe",
					Url:  "https://static-cdn.jtvnw.net/emoticons/v2/4151865/default/dark/1.0",
				},
			},
		},
		{
			"message with a multiple emotes, multiple spans, literal dollar",
			"Lincoln presidAbe presidAbe is on the $5 bill FrankerZ !",
			[]emoteInfo{
				{
					id: "4151865",
					spans: []emoteSpan{
						{8, 16},
						{18, 26},
					},
				},
				{
					id: "65",
					spans: []emoteSpan{
						{46, 53},
					},
				},
			},
			"",
			"Lincoln $0 $0 is on the $$5 bill $1 !",
			[]EmoteDetails{
				{
					Name: "presidAbe",
					Url:  "https://static-cdn.jtvnw.net/emoticons/v2/4151865/default/dark/1.0",
				},
				{
					Name: "FrankerZ",
					Url:  "https://static-cdn.jtvnw.net/emoticons/v2/65/default/dark/1.0",
				},
			},
		},
		{
			"message with multi-byte characters",
			"broadcast 69—Nice lullaSup",
			[]emoteInfo{
				{
					id: "emotesv2_c697ab2d9be341c693b23aae6ed1e101",
					spans: []emoteSpan{
						{18, 25},
					},
				},
			},
			"",
			"broadcast 69—Nice $0",
			[]EmoteDetails{
				{
					Name: "lullaSup",
					Url:  "https://static-cdn.jtvnw.net/emoticons/v2/emotesv2_c697ab2d9be341c693b23aae6ed1e101/default/dark/1.0",
				},
			},
		},
		{
			"bad span ordering",
			"Lincoln presidAbe presidAbe",
			[]emoteInfo{
				{
					id: "4151865",
					spans: []emoteSpan{
						{16, 8},
						{18, 26},
					},
				},
			},
			"invalid span bounds 16-8",
			"",
			nil,
		},
		{
			"bad span bounds",
			"Lincoln presidAbe presidAbe",
			[]emoteInfo{
				{
					id: "4151865",
					spans: []emoteSpan{
						{8, 99},
						{18, 26},
					},
				},
			},
			"invalid span bounds 8-99",
			"",
			nil,
		},
		{
			"inconsistent emote name",
			"Lincoln presidAbe presidJohn",
			[]emoteInfo{
				{
					id: "4151865",
					spans: []emoteSpan{
						{8, 16},
						{18, 27},
					},
				},
			},
			"name of emote 4151865 was detected as 'presidAbe' in first span; 'presidJohn' subsequently",
			"",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotEmoteDetails, err := substituteEmotes(tt.body, tt.emoteInfos)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Empty(t, gotText)
				assert.Nil(t, gotEmoteDetails)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantText, gotText)
				assert.Equal(t, tt.wantEmoteDetails, gotEmoteDetails)
			}
		})
	}
}
