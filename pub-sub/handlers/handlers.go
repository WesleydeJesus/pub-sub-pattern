package handlers

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "pub-sub/broker"
)

func InitializeRouters(broker *broker.Broker) *gin.Engine {
    route := gin.Default()

    	route.POST("/ws/publish", func(c *gin.Context) {
		message := c.PostForm("message")
		broker.Publish(message)
		c.JSON(200, gin.H{"status": "ok"})
	})

	route.GET("/ws/subscribe", func(c *gin.Context) {
		webSocket, err := broker.Upgrader.Upgrade(c.Writer, c.Request, nil)
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

    return route
}