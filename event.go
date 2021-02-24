package makeless_go_event_redis

import (
	"context"
	"sync"

	"github.com/gin-contrib/sse"
	"github.com/go-redis/redis/v8"
	"github.com/makeless/makeless-go/event"
)

type Event struct {
	Name     string
	Addr     string
	Password string
	Db       int

	Client *redis.Client
	pubSub *redis.PubSub

	BaseEvent makeless_go_event.Event
	*sync.RWMutex
}

func (event *Event) GetName() string {
	event.RLock()
	defer event.RUnlock()

	return event.Name
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
				}

				if err := message.UnmarshalBinary([]byte(channel.Payload)); err != nil {
					event.TriggerError(err)
					continue
				}

				switch message.UserId {
				case 0:
					if err := event.GetBaseEvent().Broadcast(message.GetChannel(), message.GetId(), message.GetData()); err != nil {
						event.TriggerError(err)
						continue
					}
				default:
					if err := event.GetBaseEvent().Trigger(message.GetUserId(), message.GetChannel(), message.GetId(), message.GetData()); err != nil {
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

	return event.Client
}

func (event *Event) setClient(client *redis.Client) {
	event.RLock()
	defer event.RUnlock()

	event.Client = client
}

func (event *Event) GetBaseEvent() makeless_go_event.Event {
	event.RLock()
	defer event.RUnlock()

	return event.BaseEvent
}

func (event *Event) NewClientId() string {
	return event.GetBaseEvent().NewClientId()
}

func (event *Event) GetHub() makeless_go_event.Hub {
	return event.GetBaseEvent().GetHub()
}

func (event *Event) Subscribe(userId uint, clientId string) {
	event.GetBaseEvent().Subscribe(userId, clientId)
}

func (event *Event) Unsubscribe(userId uint, clientId string) {
	event.GetBaseEvent().Unsubscribe(userId, clientId)
}

func (event *Event) Trigger(userId uint, channel string, id string, data interface{}) error {
	return event.GetClient().Publish(event.GetClient().Context(), event.GetName(), &Message{
		UserId:  userId,
		Channel: channel,
		Id:      id,
		Data:    data,
	}).Err()
}

func (event *Event) TriggerError(err error) {
	event.GetBaseEvent().TriggerError(err)
}

func (event *Event) Broadcast(channel string, id string, data interface{}) error {
	return event.GetClient().Publish(event.GetClient().Context(), event.GetName(), &Message{
		UserId:  0,
		Channel: channel,
		Id:      id,
		Data:    data,
	}).Err()
}

func (event *Event) Listen(userId uint, clientId string) chan sse.Event {
	return event.GetBaseEvent().Listen(userId, clientId)
}

func (event *Event) ListenError() chan error {
	return event.GetBaseEvent().ListenError()
}
