package chatlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MessageBuffer(t *testing.T) {
	b := newEventBuffer(4)

	assert.Len(t, b.events, 4)
	assert.Equal(t, 4, b.capacity)
	assert.Equal(t, 0, b.size)
	assert.Equal(t, 0, b.headIndex)

	b.push(&Event{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "1", UserId: "alice", Text: "hello from alice"}}})
	b.push(&Event{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "2", UserId: "bob", Text: "hello from bob"}}})
	b.push(&Event{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "3", UserId: "alice", Text: "hello again from alice"}}})

	assert.Len(t, b.events, 4)
	assert.Equal(t, 4, b.capacity)
	assert.Equal(t, 3, b.size)
	assert.Equal(t, 3, b.headIndex)

	assert.Equal(t, b.take(4), []*Event{
		{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "1", UserId: "alice", Text: "hello from alice"}}},
		{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "2", UserId: "bob", Text: "hello from bob"}}},
		{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "3", UserId: "alice", Text: "hello again from alice"}}},
	})

	assert.Equal(t, b.take(2), []*Event{
		{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "2", UserId: "bob", Text: "hello from bob"}}},
		{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "3", UserId: "alice", Text: "hello again from alice"}}},
	})

	b.push(&Event{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "4", UserId: "charlie", Text: "hello from charlie"}}})
	b.push(&Event{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "5", UserId: "bob", Text: "hello again from bob"}}})
	b.push(&Event{Type: EventTypeClear})
	b.push(&Event{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "6", UserId: "alice", Text: "hello for the last time from alice"}}})

	assert.Len(t, b.events, 4)
	assert.Equal(t, 4, b.capacity)
	assert.Equal(t, 4, b.size)
	assert.Equal(t, 3, b.headIndex)

	assert.Equal(t, b.take(4), []*Event{
		{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "4", UserId: "charlie", Text: "hello from charlie"}}},
		{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "5", UserId: "bob", Text: "hello again from bob"}}},
		{Type: EventTypeClear},
		{Type: EventTypeAppend, Payload: &Payload{Append: &PayloadAppend{MessageId: "6", UserId: "alice", Text: "hello for the last time from alice"}}},
	})
}
