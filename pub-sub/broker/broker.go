package broker

import ( "sync"
		 "github.com/gorilla/websocket"
		 "net/http"
)

type Broker struct {
	stopChannel chan struct{}
	publishChannel chan string
	subscribeChannel map[chan string]struct{}
	mutex sync.RWMutex
	Upgrader websocket.Upgrader
}

func NewBroker() *Broker {
	return &Broker{
		stopChannel: make(chan struct{}),
		publishChannel: make(chan string),
		subscribeChannel: make(map[chan string]struct{}),
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

func (broker *Broker) Subscribe() chan string {
	channel := make(chan string)
	broker.mutex.Lock()
	broker.subscribeChannel[channel] = struct{}{}
	broker.mutex.Unlock()
	return channel
}

func (broker *Broker) Unsubscribe(channel chan string) {
	broker.mutex.Lock()
	delete(broker.subscribeChannel, channel)
	broker.mutex.Unlock()
	close(channel)
}

func (broker *Broker) Publish(message string) {
	broker.publishChannel <- message
}
