package database

import (
	"call-center-api/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *KafkaProducer) PublishIncomingCall(ctx context.Context, call models.IncomingCall) error {
	data, err := json.Marshal(call)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(call.CallID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

func (p *KafkaProducer) PublishAssignedCall(ctx context.Context, call models.AssignedCall) error {
	data, err := json.Marshal(call)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "assigned_calls", // Always use assigned_calls topic
		Key:   sarama.StringEncoder(call.CallID),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		fmt.Printf("Error publishing assigned call to Kafka: %v\n", err)
		return err
	}
	fmt.Printf("Published assigned call %s to Kafka topic assigned_calls partition %d offset %d\n", call.CallID, partition, offset)
	return nil
}

// PublishMessage publishes a generic message with key and value
func (p *KafkaProducer) PublishMessage(ctx context.Context, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		fmt.Printf("Error publishing message to Kafka: %v\n", err)
		return err
	}
	fmt.Printf("Published message with key '%s' to Kafka topic %s partition %d offset %d\n", key, p.topic, partition, offset)
	return nil
}

func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// KafkaConsumer for consuming messages
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
}

func NewKafkaConsumer(brokers []string, topic, groupID string) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		consumer: consumer,
		topic:    topic,
	}, nil
}

func (c *KafkaConsumer) ConsumeMessages(ctx context.Context, handler func(models.IncomingCall) error) error {
	handlerWrapper := &consumerGroupHandler{
		handler: c.topic,
		processIncoming: func(call models.IncomingCall) error {
			return handler(call)
		},
	}

	// Keep consuming until context is canceled
	for {
		if err := c.consumer.Consume(ctx, []string{c.topic}, handlerWrapper); err != nil {
			// Check if it's a context cancellation
			if ctx.Err() != nil {
				fmt.Printf("Consumer context canceled for topic %s\n", c.topic)
				return ctx.Err()
			}
			// For other errors, log and return
			fmt.Printf("Consumer error for topic %s: %v\n", c.topic, err)
			return err
		}

		// Consume returns when rebalancing or on errors
		// Check if context is done before continuing
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *KafkaConsumer) ConsumeAssignedCalls(ctx context.Context, handler func(models.AssignedCall) error) error {
	handlerWrapper := &consumerGroupHandler{
		handler: c.topic,
		processAssigned: func(call models.AssignedCall) error {
			return handler(call)
		},
	}

	// Keep consuming until context is canceled
	for {
		if err := c.consumer.Consume(ctx, []string{c.topic}, handlerWrapper); err != nil {
			// Check if it's a context cancellation
			if ctx.Err() != nil {
				fmt.Printf("Consumer context canceled for topic %s\n", c.topic)
				return ctx.Err()
			}
			// For other errors, log and return
			fmt.Printf("Consumer error for topic %s: %v\n", c.topic, err)
			return err
		}

		// Consume returns when rebalancing or on errors
		// Check if context is done before continuing
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// ConsumeRawMessages consumes raw messages with access to key and value
func (c *KafkaConsumer) ConsumeRawMessages(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return c.consumer.Consume(ctx, topics, handler)
}

func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler         string
	processIncoming func(models.IncomingCall) error
	processAssigned func(models.AssignedCall) error
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	fmt.Printf("Consumer group handler setup - MemberID: %s, GenerationID: %d\n", session.MemberID(), session.GenerationID())
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	fmt.Printf("Consumer group handler cleanup\n")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fmt.Printf("Starting ConsumeClaim for topic %s partition %d at offset %d\n", claim.Topic(), claim.Partition(), claim.InitialOffset())

	for message := range claim.Messages() {
		fmt.Printf("Received Kafka message from topic %s partition %d offset %d\n", message.Topic, message.Partition, message.Offset)

		// Try to unmarshal as AssignedCall first (has more fields)
		var assignedCall models.AssignedCall
		if err := json.Unmarshal(message.Value, &assignedCall); err == nil && assignedCall.AssignedAgentID != "" {
			fmt.Printf("Decoded as AssignedCall: %s for agent %s\n", assignedCall.CallID, assignedCall.AssignedAgentID)
			if h.processAssigned != nil {
				if err := h.processAssigned(assignedCall); err != nil {
					fmt.Printf("Error processing assigned call: %v\n", err)
				}
			}
		} else {
			// Try to unmarshal as IncomingCall
			var incomingCall models.IncomingCall
			if err := json.Unmarshal(message.Value, &incomingCall); err == nil && incomingCall.CallID != "" {
				fmt.Printf("Decoded as IncomingCall: %s\n", incomingCall.CallID)
				if h.processIncoming != nil {
					if err := h.processIncoming(incomingCall); err != nil {
						fmt.Printf("Error processing incoming call: %v\n", err)
					}
				}
			} else {
				fmt.Printf("Error unmarshaling message: %v, data: %s\n", err, string(message.Value))
			}
		}
		session.MarkMessage(message, "")
	}
	fmt.Printf("ConsumeClaim loop exited for topic %s partition %d\n", claim.Topic(), claim.Partition())
	return nil
}
