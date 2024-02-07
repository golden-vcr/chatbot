package chatlog

import "sync"

// messageBuffer is a fixed-size ring buffer recording the user IDs associated with the
// N most recent messages
type messageBuffer struct {
	messages  []*PayloadAppend
	capacity  int
	size      int
	headIndex int

	mu sync.Mutex
}

// newMessageBuffer initializes an empty messageBuffer that will hold 'append' payloads
// up to the given capacity
func newMessageBuffer(capacity int) *messageBuffer {
	return &messageBuffer{
		messages:  make([]*PayloadAppend, capacity, capacity),
		capacity:  capacity,
		size:      0,
		headIndex: 0,
	}
}

// first returns the index of the oldest item in the buffer, or -1 if no items are
// buffered
func (b *messageBuffer) first() int {
	if b.size == 0 {
		return -1
	}
	if b.size < b.capacity {
		return 0
	}
	return b.headIndex
}

// filter modifies the buffer in place so that it only includes items that match the
// given predicate
func (b *messageBuffer) filter(pred func(*PayloadAppend) bool) {
	// Get an index to the oldest item in the buffer, and early-out if no items exist
	i := b.first()
	if i < 0 {
		return
	}

	// Prepare a copy of our messages slice into which we can write our updated values
	newMessages := make([]*PayloadAppend, b.capacity, b.capacity)
	newSize := 0
	for n := 0; n < b.size; n++ {
		if pred(b.messages[i]) {
			newMessages[newSize] = b.messages[i]
			newSize++
		}
		i = (i + 1) % b.capacity
	}

	// If the size is unchanged (we ended up with exactly the same number of messages),
	// then we have the same set of elements and can discard our copy without changing
	// buffer state
	if newSize == b.size {
		return
	}

	// Otherwise, we've shrunk the buffer, and we can swap in the new slice of messages:
	// we're back to having our oldest element at 0, and since the buffer has shrunk we
	// know there's enough slack for it to hold newly-inserted elements thereafter
	b.messages = newMessages
	b.size = newSize
	b.headIndex = newSize
}

// append registers a new item recording the fact that the given user sent a message
// with the given ID, potentially ejecting the oldest item from the buffer in the
// process
func (b *messageBuffer) append(payload *PayloadAppend) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages[b.headIndex] = payload
	b.headIndex = (b.headIndex + 1) % b.capacity
	b.size = min(b.size+1, b.capacity)
}

// take returns a properly-ordered slice contaning up to n buffered messages
func (b *messageBuffer) take(n int) []*PayloadAppend {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Take up to n, but no more than the number of itmes actually buffered
	i := b.first()
	numToTake := n
	if numToTake < b.size {
		// If we're taking fewer elements than we have, shift forward to make sure we
		// get the n newest elements
		i = (i + b.size - n) % b.capacity
	} else if numToTake > b.size {
		numToTake = b.size
	}

	// Iterate from startIndex until we've appended numToTake elements into our result
	result := make([]*PayloadAppend, 0, numToTake)
	for n := 0; n < numToTake; n++ {
		result = append(result, b.messages[i])
		i = (i + 1) % b.capacity
	}
	return result
}

// ban searches the buffer for all recent messages sent by the user with the given ID,
// and removes all matching items
func (b *messageBuffer) ban(userId string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.filter(func(message *PayloadAppend) bool {
		return message.UserId != userId
	})
}

// delete removes all messages that match the given messageId
func (b *messageBuffer) delete(messageId string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.filter(func(message *PayloadAppend) bool {
		return message.MessageId != messageId
	})
}
