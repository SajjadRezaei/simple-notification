package rabbit

import (
	"context"
	"log"
	"notification-service/src/infra/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"

	"notification-service/src/config"
	"notification-service/src/pkg/socket"
)

// ConsumeMessage dynamically consumes messages for all queues in the special goroutine
func ConsumeMessage(cfg *config.RabbitMQConfig, rmq *rabbitmq.RabbitMQ, ctx context.Context) {
	for _, queueName := range cfg.ServiceToQueue {
		go func(queue string) {
			log.Printf("Starting consumer for queue: %s", queue)

			// Start consuming messages from the queue
			mags, err := rmq.Ch.Consume(
				queue,
				"", // consumer tag (auto-generated if empty)
				false,
				false,
				false,
				false,
				nil,
			)

			if err != nil {
				log.Fatalf("Failed to start consuming messages from queue %s", err)
			}

			for {
				select {
				case msg := <-mags:
					processMessage(queueName, msg, err)
				case <-ctx.Done():
					log.Printf("Stopping consumer for queue: %s", queueName)
					return
				}
			}
		}(queueName)
	}
}

// processMessage handles the processing of a single message
func processMessage(topic string, msg amqp.Delivery, err error) {
	if len(msg.Body) > 0 {
		log.Printf("Received message from topic %s: %s", topic, string(msg.Body))

		success := socket.BroadcastMessage(topic, msg.Body)
		if success {
			// Acknowledge the message if it was successfully sent
			if err = msg.Ack(false); err != nil {
				log.Printf("Failed to acknowledge message: %v", err)
			}
		} else {
			//  prevent requeue and send to  DLQ
			if err = msg.Nack(false, false); err != nil {
				log.Printf("Failed to reject message: %v", err)
			}
		}
	}
}
