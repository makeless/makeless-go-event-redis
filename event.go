package makeless_go_event_redis

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/makeless/makeless-go/event"
	"github.com/makeless/makeless-go/event/basic"
)

type Event struct {
	makeless_go_event.Event

	Name     string
	Addr     string
	Password string
	Db       int

	client *redis.Client
	pubSub *redis.PubSub

	*sync.RWMutex
}

func (event *Event) getAddr() string {
	event.RLock()
	defer event.RUnlock()

	return event.Addr
}

func (event *Event) getDb() int {
	event.RLock()
	defer event.RUnlock()

	return event.Db
}

func (event *Event) getPassword() string {
	event.RLock()
	defer event.RUnlock()

	return event.Password
}

func (event *Event) getPubSub() *redis.PubSub {
	event.RLock()
	defer event.RUnlock()

	return event.pubSub
}

func (event *Event) setPubSub(pubSub *redis.PubSub) {
	event.Lock()
	defer event.Unlock()

	event.pubSub = pubSub
}

func (event *Event) Init() error {
	client := redis.NewClient(&redis.Options{
		Addr:     event.getAddr(),
		Password: event.getPassword(),
		DB:       event.getDb(),
	})

	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		return err
	}

	event.setClient(client)

	event.setPubSub(
		event.GetClient().Subscribe(event.GetClient().Context(), event.GetName()),
	)

	if _, err := event.getPubSub().Receive(event.GetClient().Context()); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case channel := <-event.getPubSub().Channel():
				var message = &Message{
					RWMutex: new(sync.RWMutex),
					EventData: &makeless_go_event_basic.EventData{
						RWMutex: new(sync.RWMutex),
					},
				}

				if err := message.UnmarshalBinary([]byte(channel.Payload)); err != nil {
					event.TriggerError(err)
					continue
				}

				switch message.UserId {
				case 0:
					if err := event.Broadcast(message.GetEventData().GetChannel(), message.GetEventData().GetId(), message.GetEventData().GetData()); err != nil {
						event.TriggerError(err)
						continue
					}
				default:
					if err := event.Trigger(message.GetUserId(), message.GetEventData().GetChannel(), message.GetEventData().GetId(), message.GetEventData().GetData()); err != nil {
						event.TriggerError(err)
						continue
					}
				}
			case <-event.GetClient().Context().Done():
				return
			}
		}
	}()

	return nil
}

func (event *Event) GetClient() *redis.Client {
	event.RLock()
	defer event.RUnlock()

	return event.client
}

func (event *Event) setClient(client *redis.Client) {
	event.RLock()
	defer event.RUnlock()

	event.client = client
}

func (event *Event) Trigger(userId uint, channel string, id string, data interface{}) error {
	return event.GetClient().Publish(event.GetClient().Context(), event.GetName(), &Message{
		UserId: userId,
		EventData: &makeless_go_event_basic.EventData{
			Channel: channel,
			Id:      id,
			Data:    data,
			RWMutex: new(sync.RWMutex),
		},
		RWMutex: new(sync.RWMutex),
	}).Err()
}

func (event *Event) Broadcast(channel string, id string, data interface{}) error {
	return event.GetClient().Publish(event.GetClient().Context(), event.GetName(), &Message{
		UserId: 0,
		EventData: &makeless_go_event_basic.EventData{
			Channel: channel,
			Id:      id,
			Data:    data,
			RWMutex: new(sync.RWMutex),
		},
		RWMutex: new(sync.RWMutex),
	}).Err()
}
