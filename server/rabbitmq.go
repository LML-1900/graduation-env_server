package server

import (
	"context"
	"encoding/json"
	"env_server/data"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type RabbitMq struct {
	Conn  *amqp.Connection
	Ch    *amqp.Channel
	qname string
}

func CreateMessageQueue(qname string, url string) (*RabbitMq, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("fail to connect to RabbitMQ, err:%v", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("fail to open a channel, err:%v", err)
		return nil, err
	}

	err = ch.ExchangeDeclare(
		qname,   // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange, err:%v", err)
		return nil, err
	}
	mq := RabbitMq{
		Conn:  conn,
		Ch:    ch,
		qname: qname,
	}
	return &mq, nil
}

func (mq *RabbitMq) BroadCastCraters(crater data.Crater) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	craterBytes, err := json.Marshal(crater)
	if err != nil {
		log.Printf("Failed to marshal crater, err:%v\n", err)
		return
	}

	err = mq.Ch.PublishWithContext(ctx,
		mq.qname,                // exchange
		data.CRATER_ROUTING_KEY, // routing key
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        craterBytes,
		})
	if err != nil {
		log.Printf("Failed to publish a message, err:%v\n", err)
		return
	}

	log.Printf(" [x] Sent Crater at [lon:%v lat:%v]", crater.Position.Longitude, crater.Position.Latitude)
}
