package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/posolwar/kafka-learn/kafka-writer/messager"
	"github.com/posolwar/kafka-learn/pkg/message"
)

type Flags struct {
	OptionRetries           int
	OptionAcks              string
	Topic                   string
	Message                 string
	AddressBootstrapServers string
	AddressSchemaRegistry   string
}

func (f *Flags) BindFlags() error {
	flag.StringVar(&f.AddressBootstrapServers, "bootstrap-servers", "", "Cписок брокеров Kafka для подключения клиента")
	flag.StringVar(&f.AddressSchemaRegistry, "schema-registry", "", "Адрес schema registry")
	flag.StringVar(&f.Topic, "topic", "", "Топик для передачи сообщения")
	flag.StringVar(&f.Message, "message", "", "Передаваемое сообщение")
	flag.IntVar(&f.OptionRetries, "retries", 3, "Количество попыток повторной отправки сообщения продюсером в случае ошибки")

	flag.Parse()

	if f.AddressBootstrapServers == "" {
		return errors.New("bootstrap-servers не может быть пустым")
	}

	if f.AddressSchemaRegistry == "" {
		return errors.New("schema registry не указан")
	}

	if f.Topic == "" {
		return errors.New("topic не может быть пустым")
	}
	if f.Message == "" {
		return errors.New("message не может быть пустым")
	}

	return nil
}

func main() {
	var flags Flags

	if err := flags.BindFlags(); err != nil {
		log.Fatal("ошибка получения флагов: " + err.Error())
	}

	producer := messager.NewMessageProducer(
		messager.NewProducerConfig(
			flags.AddressBootstrapServers,
			flags.OptionRetries,
		),
	)

	if err := producer.SetKafkaClient(); err != nil {
		slog.Error("ошибка создания kafka клиента", "err", err.Error())
		os.Exit(1)
	}
	defer producer.KafkaClient.Close()

	if err := producer.SetSrClient(flags.AddressSchemaRegistry); err != nil {
		slog.Error("ошибка создания schema registry клиента", "err", err.Error())
		os.Exit(1)
	}

	if err := producer.SetSerializer(); err != nil {
		slog.Error("ошибка создания сериализатора", "err", err.Error())
		os.Exit(1)
	}

	if err := producer.SetTopic(flags.Topic); err != nil {
		slog.Error("ошибка установки топика", "err", err.Error())
		os.Exit(1)
	}

	if err := producer.Produce(context.Background(), message.MessageBody{Message: flags.Message}); err != nil {
		slog.Error("ошибка отправки сообщения", "err", err.Error())
		os.Exit(1)
	}
}
