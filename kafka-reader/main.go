package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/posolwar/kafka-learn/kafka-reader/consumer"
)

type Flags struct {
	ConsumerGroupID         int
	Topic                   string
	AddressBootstrapServers string
	AddressSchemaRegistry   string
	ConsumerMode            string
}

func (f *Flags) Bind() error {
	flag.StringVar(&f.AddressBootstrapServers, "bootstrap-servers", "localhost:9094", "Список брокеров Kafka")
	flag.StringVar(&f.AddressSchemaRegistry, "schema-registry", "http://localhost:8081", "Адрес Schema Registry")
	flag.StringVar(&f.Topic, "topic", "", "Топик для подписки на сообщения")
	flag.StringVar(&f.ConsumerMode, "mode", string(consumer.SingleMode), "Тип обработки сообщения (single/batch)")
	flag.IntVar(&f.ConsumerGroupID, "group-id", 0, "ID группы для SingleMessageConsumer")

	flag.Parse()

	if f.AddressBootstrapServers == "" {
		return errors.New("bootstrap-servers не может быть пустым")
	}

	if f.ConsumerMode != string(consumer.SingleMode) && f.ConsumerMode != string(consumer.BatchMode) {
		return errors.New("неизвестный тип консьюмера, ожидается " + string(consumer.SingleMode) + " или " + string(consumer.BatchMode))
	}

	if f.AddressSchemaRegistry == "" {
		return errors.New("schema registry не указан")
	}

	if f.Topic == "" {
		return errors.New("topic не может быть пустым")
	}

	return nil
}

func main() {
	var flags Flags

	if err := flags.Bind(); err != nil {
		log.Fatal("ошибка получения флагов: " + err.Error())
	}

	consumer := consumer.NewMessageConsumer(
		consumer.NewConsumerConfig(
			flags.AddressBootstrapServers,
			flags.ConsumerGroupID,
			consumer.ConsumerMode(flags.ConsumerMode)))

	if err := consumer.SetKafkaClient(); err != nil {
		slog.Error("ошибка создания kafka клиента", "err", err.Error())
		os.Exit(1)
	}
	defer consumer.KafkaClient.Close()

	if err := consumer.SetSrClient(flags.AddressSchemaRegistry); err != nil {
		slog.Error("ошибка создания schema registry клиента", "err", err.Error())
		os.Exit(1)
	}

	if err := consumer.SetSerializer(); err != nil {
		slog.Error("ошибка создания сериализатора", "err", err.Error())
		os.Exit(1)
	}

	if err := consumer.SetTopic(flags.Topic); err != nil {
		slog.Error("ошибка установки топика", "err", err.Error())
		os.Exit(1)
	}

	if err := consumer.Consume(context.Background()); err != nil {
		slog.Error("ошибка получения сообщения", "err", err.Error())
		os.Exit(1)
	}
}
