package main

import (
	"bytes"
	"context"

	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/exp/rand"
)

type EventType string

const (
	USER_SIGNUP     EventType = "signup"
	ORDER_PLACEMENT EventType = "orderplacement"
)

type MessageTriggerEvent struct {
	BrandId    int               `json:"brandId,omitempty"`
	Recipient  string            `json:"recipient,omitempty"`
	UserId     int               `json:"userId,omitempty"`
	EventType  EventType         `json:"eventType,omitempty"`
	CustomData map[string]string `json:"customData,omitempty"`
}

const timeInterval = 5

func main() {
	conn, err := amqp.Dial("amqp://yt_user:password@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"IntegrationQueue",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to create queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		message, _ := json.Marshal(generateSignupEvent())
		err = ch.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})

		failOnError(err, "Failed to publish  message")

		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, message, "", "  ")
		if err != nil {
			log.Printf("Failed to generate pretty JSON: %s", err)
		} else {
			log.Printf("[x] Sent %s\n", prettyJSON.String())
		}

		time.Sleep(time.Second * timeInterval)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func generateSignupEvent() *MessageTriggerEvent {

	events := []*MessageTriggerEvent{
		{
			BrandId:   123,
			Recipient: "911234567890",
			UserId:    123456,
			CustomData: map[string]string{
				"firstname": "John",
				"surname":   "Kented",
				"reward":    "300",
			},
			EventType: "signup",
		},
		{
			BrandId:   123,
			Recipient: "911234567890",
			UserId:    123456,
			CustomData: map[string]string{
				"firstname": "John",
				"surname":   "Kented",
				"reward":    "300",
			},
			EventType: "orderplacement",
		},
	}

	return events[rand.Intn(len(events))]
}
