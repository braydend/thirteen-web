package main

import (
	"fmt"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var clients map[string]*websocket.Conn

func main() {
	clients = make(map[string]*websocket.Conn)
	app := fiber.New()

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	/*
		Joins connection pool and assigns a random UUID
	*/
	app.Get("/ws/join", websocket.New(func(c *websocket.Conn) {
		uuid, err := uuid.NewUUID()

		if err != nil {
			log.Fatalf("Failed to generate UUID")
		}

		id := uuid.String()

		clients[id] = c

		messageHandler(c, id)
	}))

	log.Fatal(app.Listen(":3000"))
}

func messageHandler(c *websocket.Conn, id string) {
	var (
		mt  int
		msg []byte
		err error
	)
	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", msg)

		broadcast(mt, msg, id)
	}
}

/*
Send a message to all established clients
*/
func broadcast(mt int, msg []byte, sender string) {
	signedMsg := fmt.Sprintf("%s: %s", sender, msg)
	for _, c := range clients {
		if err := c.WriteMessage(mt, []byte(signedMsg)); err != nil {
			log.Println("write:", err)
			break
		}
	}
}
