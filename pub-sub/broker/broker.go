package broker

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type Broker struct {
	stopChannel      chan struct{}
	publishChannel   chan []byte
	subscribeChannel map[chan []byte]struct{}
	mutex            sync.RWMutex
	Upgrader         websocket.Upgrader
}

func NewBroker() *Broker {
	return &Broker{
		stopChannel:      make(chan struct{}),
		publishChannel:   make(chan []byte),
		subscribeChannel: make(map[chan []byte]struct{}),
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (broker *Broker) Start() {
	for {
		select {
		case <-broker.stopChannel:
			return
		case message := <-broker.publishChannel:
			broker.mutex.RLock()
			for channel := range broker.subscribeChannel {
				channel <- message
			}
			broker.mutex.RUnlock()
		}
	}
}

func (broker *Broker) Stop() {
	close(broker.stopChannel)
}

func (broker *Broker) Subscribe() chan []byte {
	channel := make(chan []byte)
	broker.mutex.Lock()
	broker.subscribeChannel[channel] = struct{}{}
	broker.mutex.Unlock()
	return channel
}

func (broker *Broker) Unsubscribe(channel chan []byte) {
	broker.mutex.Lock()
	delete(broker.subscribeChannel, channel)
	broker.mutex.Unlock()
	close(channel)
}

func (broker *Broker) Publish(message []byte) {
	broker.publishChannel <- message
}
