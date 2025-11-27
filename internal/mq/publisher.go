package mq

import (
	"log"

	"github.com/streadway/amqp"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func Connect(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	err = ch.ExchangeDeclare(
		"game.events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &Publisher{conn, ch}, nil
}

func (p *Publisher) Publish(routingKey, message string) {
	err := p.channel.Publish(
		"game.events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
	}
}

func (p *Publisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
