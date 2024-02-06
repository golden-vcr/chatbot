package chatlog

import "encoding/json"

type EventType string

const (
	EventTypeAppend EventType = "append"
	EventTypeDelete EventType = "delete"
	EventTypeBan    EventType = "ban"
	EventTypeClear  EventType = "clear"
)

type Event struct {
	Type    EventType `json:"type"`
	Payload *Payload  `json:"payload,omitempty"`
}

type Payload struct {
	Append *PayloadAppend
	Delete *PayloadDelete
	Ban    *PayloadBan
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type fields struct {
		Type    EventType       `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	var f fields
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}

	e.Type = f.Type
	switch f.Type {
	case EventTypeAppend:
		e.Payload = &Payload{}
		return json.Unmarshal(f.Payload, &e.Payload.Append)
	case EventTypeDelete:
		e.Payload = &Payload{}
		return json.Unmarshal(f.Payload, &e.Payload.Delete)
	case EventTypeBan:
		e.Payload = &Payload{}
		return json.Unmarshal(f.Payload, &e.Payload.Ban)
	}
	return nil
}

func (p Payload) MarshalJSON() ([]byte, error) {
	if p.Append != nil {
		return json.Marshal(p.Append)
	}
	if p.Delete != nil {
		return json.Marshal(p.Delete)
	}
	if p.Ban != nil {
		return json.Marshal(p.Ban)
	}
	return json.Marshal(nil)
}

type PayloadAppend struct {
	MessageId string         `json:"messageId"`
	UserId    string         `json:"userId"`
	Username  string         `json:"username"`
	Color     string         `json:"color"`
	Text      string         `json:"text"`
	Emotes    []EmoteDetails `json:"emotes"`
}

type EmoteDetails struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PayloadDelete struct {
	MessageId string `json:"messageId"`
}

type PayloadBan struct {
	UserId string `json:"userId"`
}
