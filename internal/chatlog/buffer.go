package chatlog

import "sync"

// eventBuffer is a fixed-size ring buffer that keeps track of the N most recent chatlog
// Events that have been handled
type eventBuffer struct {
	events    []*Event
	capacity  int
	size      int
	headIndex int

	mu sync.Mutex
}

// newEventBuffer initializes an empty eventBuffer that will hold Event structs up to
// to the given capacity
func newEventBuffer(capacity int) *eventBuffer {
	return &eventBuffer{
		events:    make([]*Event, capacity, capacity),
		capacity:  capacity,
		size:      0,
		headIndex: 0,
	}
}

// first returns the index of the oldest item in the buffer, or -1 if no items are
// buffered
func (b *eventBuffer) first() int {
	if b.size == 0 {
		return -1
	}
	if b.size < b.capacity {
		return 0
	}
	return b.headIndex
}

// push adds a new event into the buffer, potentially ejecting the oldest event in the
// process
func (b *eventBuffer) push(event *Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events[b.headIndex] = event
	b.headIndex = (b.headIndex + 1) % b.capacity
	b.size = min(b.size+1, b.capacity)
}

// take returns a properly-ordered slice contaning up to n buffered events
func (b *eventBuffer) take(n int) []*Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Take up to n, but no more than the number of events actually buffered
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
	result := make([]*Event, 0, numToTake)
	for n := 0; n < numToTake; n++ {
		result = append(result, b.events[i])
		i = (i + 1) % b.capacity
	}
	return result
}
