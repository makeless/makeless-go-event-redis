package go_saas_event_redis

import (
	"encoding/json"
	"sync"
)

type Message struct {
	UserId  uint        `json:"userId"` // 0: broadcast
	Channel string      `json:"channel"`
	Id      string      `json:"id"`
	Data    interface{} `json:"data"`

	*sync.RWMutex
}

func (message *Message) GetUserId() uint {
	message.RLock()
	defer message.RUnlock()

	return message.UserId
}

func (message *Message) GetChannel() string {
	message.RLock()
	defer message.RUnlock()

	return message.Channel
}

func (message *Message) GetId() string {
	message.RLock()
	defer message.RUnlock()

	return message.Id
}

func (message *Message) GetData() interface{} {
	message.RLock()
	defer message.RUnlock()

	return message.Data
}

func (message *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(message)
}

func (message *Message) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, message)
}
