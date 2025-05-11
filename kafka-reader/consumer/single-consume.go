package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/posolwar/kafka-learn/pkg/message"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (mc *MessageConsumer) singleConsume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := mc.KafkaClient.ReadMessage(time.Duration(mc.cfg.FetchMaxWaitMs))
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				slog.Error("SingleMessageConsumer", "err", err.Error())

				continue
			}

			var message message.MessageBody
			if err := mc.Deserializer.DeserializeInto(mc.Topic, msg.Value, &message); err != nil {
				slog.Error("SingleMessageConsumer десериализация", "err", err.Error())
				continue
			}

			// TODO: В дальнейшем вынести в канал для вывода наружу
			slog.Info("SingleMessageConsumer получено", "сообщение:", message.Message)
		}
	}
}
