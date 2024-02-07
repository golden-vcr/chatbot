package chatlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MessageBuffer(t *testing.T) {
	b := newMessageBuffer(4)

	assert.Len(t, b.messages, 4)
	assert.Equal(t, 4, b.capacity)
	assert.Equal(t, 0, b.size)
	assert.Equal(t, 0, b.headIndex)

	b.append(&PayloadAppend{MessageId: "1", UserId: "alice", Text: "hello from alice"})
	b.append(&PayloadAppend{MessageId: "2", UserId: "bob", Text: "hello from bob"})
	b.append(&PayloadAppend{MessageId: "3", UserId: "alice", Text: "hello again from alice"})

	assert.Len(t, b.messages, 4)
	assert.Equal(t, 4, b.capacity)
	assert.Equal(t, 3, b.size)
	assert.Equal(t, 3, b.headIndex)

	assert.Equal(t, b.take(4), []*PayloadAppend{
		{MessageId: "1", UserId: "alice", Text: "hello from alice"},
		{MessageId: "2", UserId: "bob", Text: "hello from bob"},
		{MessageId: "3", UserId: "alice", Text: "hello again from alice"},
	})

	assert.Equal(t, b.take(2), []*PayloadAppend{
		{MessageId: "2", UserId: "bob", Text: "hello from bob"},
		{MessageId: "3", UserId: "alice", Text: "hello again from alice"},
	})

	b.append(&PayloadAppend{MessageId: "4", UserId: "charlie", Text: "hello from charlie"})
	b.append(&PayloadAppend{MessageId: "5", UserId: "bob", Text: "hello again from bob"})
	b.append(&PayloadAppend{MessageId: "6", UserId: "alice", Text: "hello for the last time from alice"})

	assert.Len(t, b.messages, 4)
	assert.Equal(t, 4, b.capacity)
	assert.Equal(t, 4, b.size)
	assert.Equal(t, 2, b.headIndex)

	assert.Equal(t, b.take(4), []*PayloadAppend{
		{MessageId: "3", UserId: "alice", Text: "hello again from alice"},
		{MessageId: "4", UserId: "charlie", Text: "hello from charlie"},
		{MessageId: "5", UserId: "bob", Text: "hello again from bob"},
		{MessageId: "6", UserId: "alice", Text: "hello for the last time from alice"},
	})

	assert.Equal(t, b.take(2), []*PayloadAppend{
		{MessageId: "5", UserId: "bob", Text: "hello again from bob"},
		{MessageId: "6", UserId: "alice", Text: "hello for the last time from alice"},
	})

	b.ban("alice")
	b.ban("nobody")

	assert.Equal(t, b.take(4), []*PayloadAppend{
		{MessageId: "4", UserId: "charlie", Text: "hello from charlie"},
		{MessageId: "5", UserId: "bob", Text: "hello again from bob"},
	})

	b.append(&PayloadAppend{MessageId: "7", UserId: "dnitra", Text: "hello from dnitra"})
	b.append(&PayloadAppend{MessageId: "8", UserId: "dnitra", Text: "hello again from dnitra"})

	assert.Equal(t, b.take(4), []*PayloadAppend{
		{MessageId: "4", UserId: "charlie", Text: "hello from charlie"},
		{MessageId: "5", UserId: "bob", Text: "hello again from bob"},
		{MessageId: "7", UserId: "dnitra", Text: "hello from dnitra"},
		{MessageId: "8", UserId: "dnitra", Text: "hello again from dnitra"},
	})

	b.append(&PayloadAppend{MessageId: "9", UserId: "dnitra", Text: "hello for the last time from dnitra"})

	assert.Equal(t, b.take(4), []*PayloadAppend{
		{MessageId: "5", UserId: "bob", Text: "hello again from bob"},
		{MessageId: "7", UserId: "dnitra", Text: "hello from dnitra"},
		{MessageId: "8", UserId: "dnitra", Text: "hello again from dnitra"},
		{MessageId: "9", UserId: "dnitra", Text: "hello for the last time from dnitra"},
	})

	b.delete("8")
	b.delete("99")

	assert.Equal(t, b.take(4), []*PayloadAppend{
		{MessageId: "5", UserId: "bob", Text: "hello again from bob"},
		{MessageId: "7", UserId: "dnitra", Text: "hello from dnitra"},
		{MessageId: "9", UserId: "dnitra", Text: "hello for the last time from dnitra"},
	})

	b.clear()

	assert.Equal(t, b.take(4), []*PayloadAppend{})
}
