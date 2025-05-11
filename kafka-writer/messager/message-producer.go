package messager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"practrice/pkg/message"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/jsonschema"
)

// MessageProducer отвечает за отправку сообщений в Kafka
type MessageProducer struct {
	Topic string

	cfg         *ProducerConfig
	SrClient    schemaregistry.Client
	Serializer  serde.Serializer
	KafkaClient *kafka.Producer
}

// NewMessageProducer Заведение MessageProducer c инициализацией
func NewMessageProducer(cfg *ProducerConfig) *MessageProducer {
	return &MessageProducer{cfg: cfg}
}

// SetKafkaClient настраивает Kafka-клиента
func (mp *MessageProducer) SetKafkaClient() error {
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": mp.cfg.BootstrapServers,
		"acks":              "all",
		"retries":           mp.cfg.Retries,
	}

	client, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return fmt.Errorf("ошибка создания Kafka-продюсера: %w", err)
	}

	mp.KafkaClient = client
	return nil
}

// SetSrClient настраивает Schema Registry клиент
func (mp *MessageProducer) SetSrClient(srAddress string) error {
	if srAddress == "" {
		return errors.New("ошибка указания sr адреса")
	}

	srClient, err := schemaregistry.NewClient(schemaregistry.NewConfig(srAddress))
	if err != nil {
		return fmt.Errorf("ошибка заведения sr клиента: %w", err)
	}

	mp.SrClient = srClient
	return nil
}

// SetSerializer настраивает сериализатор
func (mp *MessageProducer) SetSerializer() error {
	if mp.SrClient == nil {
		return fmt.Errorf("schema registry клиент не установлен")
	}

	serializer, err := jsonschema.NewSerializer(mp.SrClient, serde.ValueSerde, jsonschema.NewSerializerConfig())
	if err != nil {
		return fmt.Errorf("ошибка создания сериализатора: %w", err)
	}

	mp.Serializer = serializer
	return nil
}

// SetTopic устанавливает топик
func (mp *MessageProducer) SetTopic(topic string) error {
	if topic == "" {
		return fmt.Errorf("топик не может быть пустым")
	}

	mp.Topic = topic
	return nil
}

// Produce отправляет сообщение в Kafka
func (mp *MessageProducer) Produce(ctx context.Context, message message.MessageBody) error {
	if mp.KafkaClient == nil {
		return errors.New("Kafka клиент не инициализирован")
	}

	if mp.Serializer == nil {
		return errors.New("сериализатор не инициализирован")
	}

	if mp.Topic == "" {
		return errors.New("топик не установлен")
	}

	// Проверка контекста
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("контекст отменен: %w", err)
	}

	// Сериализация сообщения
	payload, err := mp.Serializer.Serialize(mp.Topic, &message)
	if err != nil {
		return fmt.Errorf("сериализация сообщения: %w", err)
	}

	// Канал для получения результата доставки
	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	// Отправка сообщения
	err = mp.KafkaClient.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &mp.Topic, Partition: kafka.PartitionAny},
			Value:          payload,
			Key:            nil, // Без ключа для round-robin (чтоб рандомно выбиралась партиция у топика)
		},
		deliveryChan,
	)
	if err != nil {
		return fmt.Errorf("отправка сообщения: %w", err)
	}

	// Ожидание подтверждения доставки
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		msg := e.(*kafka.Message)
		if msg.TopicPartition.Error != nil {
			return fmt.Errorf("ошибка доставки: %w", msg.TopicPartition.Error)
		}

		slog.Info("Сообщение доставлено", "message", message, "topic", mp.Topic, "partition", msg.TopicPartition.Partition, "offset", msg.TopicPartition.Offset)
	}

	return nil
}
