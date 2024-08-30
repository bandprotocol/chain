package emitter

import abci "github.com/cometbft/cometbft/abci/types"

// EventQuerier is a helper struct that helps to find the event with some conditions.
type EventQuerier struct {
	events []abci.Event
}

// NewEventQuerier creates a new EventQuerier with the given events.
func NewEventQuerier(events []abci.Event) *EventQuerier {
	return &EventQuerier{events: events}
}

// FindEventWithTypeBeforeIdx returns the last event of the given type that is emitted before the given index.
func (es *EventQuerier) FindEventWithTypeBeforeIdx(eventType string, idx int) (abci.Event, bool) {
	for i := idx; i >= 0; i-- {
		if es.events[i].Type == eventType {
			return es.events[i], true
		}
	}

	return abci.Event{}, false
}

// FindEventWithTypeAfterIdx returns the first event of the given type that is emitted after the given index.
func (es *EventQuerier) FindEventWithTypeAfterIdx(eventType string, idx int) (abci.Event, bool) {
	for i := idx + 1; i < len(es.events); i++ {
		if es.events[i].Type == eventType {
			return es.events[i], true
		}
	}

	return abci.Event{}, false
}
