package consumer

import (
	"context"
	"log/slog"

	"practrice/pkg/message"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (mc *MessageConsumer) batchConsume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Список для накопления сообщений в пачке
			batch := make([]*kafka.Message, 0, mc.cfg.MaxPollRecords)

			// Собираем сообщения до достижения minBatchSize или истечения времени
			for len(batch) < mc.cfg.MaxPollRecords {
				ev := mc.KafkaClient.Poll(100) // Короткий таймаут для частых проверок
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					batch = append(batch, e)
				case kafka.Error:
					slog.Error("BatchMessageConsumer", "err", e)
					continue
				}
			}

			// Если собрали сообщения, обрабатываем их
			if len(batch) > 0 {
				slog.Error("BatchMessageConsumer собрал сообщения", "кол-во", len(batch))

				for _, msg := range batch {
					var message message.MessageBody
					if err := mc.Deserializer.DeserializeInto(mc.Topic, msg.Value, &message); err != nil {
						slog.Error("BatchMessageConsumer десериализация", "err", err.Error())
						continue
					}

					slog.Info("BatchMessageConsumer получено", "сообщения", message.Message)
				}

				if _, err := mc.KafkaClient.Commit(); err != nil {
					slog.Error("BatchMessageConsumer ошибка коммита", "err", err)
				}
			}
		}
	}
}
