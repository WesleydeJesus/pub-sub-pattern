package main

import ( "sync"
		 "log"
		 "github.com/gorilla/websocket"
		 "github.com/gin-gonic/gin"
)

type Broker struct {
	stopChannel chan struct{}
	publishChannel chan string
	subscribeChannel map[chan string]struct{}
	mutex sync.RWMutex
	upgrader websocket.Upgrader
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
	defer broker.Stop()

	route := gin.Default()

	route.POST("/ws/publish", func(c *gin.Context) {
		message := c.PostForm("message")
		broker.Publish(message)
		c.JSON(200, gin.H{"status": "ok"})
	})

	route.GET("/ws/subscribe", func(c *gin.Context) {
		webSocket, err := broker.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer webSocket.Close()

		subscriber := broker.Subscribe()
		defer broker.Unsubscribe(subscriber)

		for {
			message := <-subscriber
			err := webSocket.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	route.Run(":8080")
}