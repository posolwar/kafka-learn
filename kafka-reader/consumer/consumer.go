package consumer

import (
	"context"
	"errors"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/jsonschema"
)

// Для обработки сообщений по одному
type MessageConsumer struct {
	Topic string

	cfg          *ConsumerConfig
	SrClient     schemaregistry.Client
	Deserializer serde.Deserializer
	KafkaClient  *kafka.Consumer
}

func NewMessageConsumer(cfg *ConsumerConfig) *MessageConsumer {
	return &MessageConsumer{cfg: cfg}
}

// SetSrClient настраивает Schema Registry клиент
func (mc *MessageConsumer) SetSrClient(srAddress string) error {
	if srAddress == "" {
		return errors.New("ошибка указания sr адреса")
	}

	srClient, err := schemaregistry.NewClient(schemaregistry.NewConfig(srAddress))
	if err != nil {
		return fmt.Errorf("ошибка заведения sr клиента: %w", err)
	}

	mc.SrClient = srClient

	return nil
}

// SetSerializer настраивает десериализатор
func (mc *MessageConsumer) SetSerializer() error {
	if mc.SrClient == nil {
		return fmt.Errorf("schema registry клиент не установлен")
	}

	deserializer, err := jsonschema.NewDeserializer(mc.SrClient, serde.ValueSerde, jsonschema.NewDeserializerConfig())
	if err != nil {
		return fmt.Errorf("ошибка создания десериализатора: %w", err)
	}

	mc.Deserializer = deserializer
	return nil
}

// SetTopic устанавливает топик
func (mc *MessageConsumer) SetTopic(topic string) error {
	if mc.KafkaClient == nil {
		return errors.New("Кафка клиент не может быть пустым при установке топика")
	}

	if topic == "" {
		return fmt.Errorf("топик не может быть пустым")
	}

	mc.Topic = topic

	// Подписываемся на топик
	if err := mc.KafkaClient.SubscribeTopics([]string{topic}, nil); err != nil {
		return fmt.Errorf("ошибка подписки на топик %s: %w", mc.Topic, err)
	}

	return nil
}

// SetKafkaClient настраивает Kafka-клиента
func (mc *MessageConsumer) SetKafkaClient() error {
	// Создаем конфигурацию Kafka
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":  mc.cfg.BootstrapServers,
		"enable.auto.commit": mc.cfg.AutoCommit,
		"group.id":           mc.cfg.GroupID,
		"auto.offset.reset":  "earliest",
	}

	if mc.cfg.FetchMinBytes != 0 {
		kafkaConfig.SetKey("fetch.min.bytes", mc.cfg.FetchMinBytes)
	}

	if mc.cfg.FetchMaxWaitMs != 0 {
		kafkaConfig.SetKey("fetch.wait.max.ms", mc.cfg.FetchMaxWaitMs)
	}

	// Создаем Kafka-консьюмера
	client, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return fmt.Errorf("ошибка создания Kafka-консьюмера: %w", err)
	}

	mc.KafkaClient = client

	return nil
}

func (mc *MessageConsumer) Consume(ctx context.Context) error {
	switch mc.cfg.Mode {
	case SingleMode:
		return mc.singleConsume(ctx)
	case BatchMode:
		return mc.batchConsume(ctx)
	default:
		return errors.New("Неожиданный мод consumer")
	}
}
