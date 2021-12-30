package makeless_go_event_redis

import (
	"encoding/json"
	"github.com/makeless/makeless-go/event/basic"
	"sync"
)

type Message struct {
	UserId    uint                               `json:"userId"` // 0: broadcast
	EventData *makeless_go_event_basic.EventData `json:"eventData"`

	*sync.RWMutex
}

func (message *Message) GetUserId() uint {
	message.RLock()
	defer message.RUnlock()

	return message.UserId
}

func (message *Message) GetEventData() *makeless_go_event_basic.EventData {
	message.RLock()
	defer message.RUnlock()

	return message.EventData
}

func (message *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(message)
}

func (message *Message) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, message)
}
