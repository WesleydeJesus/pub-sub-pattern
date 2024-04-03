package main

import ( "fmt"
		"sync"
		// "time"
		// "log"
)

type Broker struct {
	stopChannel chan struct{}
	publishChannel chan string
	subscribeChannel map[chan string]struct{}
	mutex sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		stopChannel: make(chan struct{}),
		publishChannel: make(chan string),
		subscribeChannel: make(map[chan string]struct{}),
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

func main() {
	broker := NewBroker()
	go broker.Start()

	subscriber := broker.Subscribe()
	go func() {
		for {
			message := <-subscriber
			fmt.Println(message)
		}
	}()

	broker.Publish("123")
	broker.Publish("som")
	broker.Publish("teste")
	broker.Publish("123")
	broker.Publish("som")

	broker.Unsubscribe(subscriber)
	broker.Stop()

}