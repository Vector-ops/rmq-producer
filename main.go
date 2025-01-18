package main

import (
	"context"
	"encoding/json"
	"os"

	"log"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageTriggerEvent struct {
	Pattern string `json:"pattern,omitempty"`
	Data    Data   `json:"data,omitempty"`
}

type Data struct {
	BrandId    int                    `json:"brandId,omitempty"`
	Recipient  string                 `json:"recipient,omitempty"`
	UserId     int                    `json:"userId,omitempty"`
	EventType  string                 `json:"eventType,omitempty"`
	CustomData map[string]interface{} `json:"customData,omitempty"`
}

var channel *amqp.Channel
var queue amqp.Queue

func main() {
	loadEnv()

	go HttpServer()

	rmqURI := os.Getenv("RMQ_URI")
	if rmqURI == "" {
		panic("RabbitMQ URI not present in env file")
	}

	queueName := os.Getenv("RMQ_QUEUE")
	if queueName == "" {
		panic("RabbitMQ queue not present in env file")
	}

	conn, err := amqp.Dial(rmqURI)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		true,  // durability
		false, // autodelete
		false, // exclusive
		false, // nowait
		nil,   // args
	)
	failOnError(err, "Failed to create queue")

	channel = ch
	queue = q

	select {}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func PublishMessage(payload *MessageTriggerEvent) error {
	message, _ := json.Marshal(payload)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return channel.PublishWithContext(ctx, "", queue.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        message,
	})

}

func loadEnv() {
	err := godotenv.Load("./.env")
	if err != nil {
		panic("env file not found.")
	}
}
