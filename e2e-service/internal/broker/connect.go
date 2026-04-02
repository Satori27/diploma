package broker

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/Satori27/diploma/e2e-service/internal/storage"
	"github.com/google/uuid"
)

const batchSize = 1000


type Broker struct {
	topic   string
	cg      sarama.ConsumerGroup
	storage *storage.Storage
}

func NewBroker(topic string, storage *storage.Storage) *Broker {
	brokers := []string{os.Getenv("KAFKA_CONN")}

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Consumer.Fetch.Default = 10 * 1024 * 1024
	config.Consumer.MaxProcessingTime = 500 * time.Millisecond



	config.Consumer.Fetch.Max = 50 * 1024 * 1024

	config.ChannelBufferSize = 1024


	cg, err := sarama.NewConsumerGroup(brokers, "e2e-service", config)
	if err != nil {
		log.Fatal("Failed to create consumer group: ", err)
	}

	b := &Broker{
		topic:   topic,
		cg:      cg,
		storage: storage,
	}

	go b.startConsume()

	return b
}

func (b *Broker) startConsume() {
	ctx := context.Background()

	consumer := &batchConsumer{
		broker:    b,
		batchSize: batchSize,
	}

	for {
		if err := b.cg.Consume(ctx, []string{b.topic}, consumer); err != nil {
			log.Println("Error from consumer: ", err)
			time.Sleep(500 * time.Millisecond)
		}
	}
}


func (b *Broker) Close() {
	b.cg.Close()
}


type batchConsumer struct {
	broker    *Broker
	batchSize int
}

func (c *batchConsumer) Setup(sarama.ConsumerGroupSession) error { return nil }

func (c *batchConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *batchConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var messages []*sarama.ConsumerMessage
	t := 100 * time.Millisecond
	ticker := time.NewTicker(t)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-claim.Messages():
			slog.Debug("Received message", "partition", msg.Partition, "offset", msg.Offset)
			if !ok {
				if len(messages) > 0 {
					c.broker.processBatch(session.Context(), messages, session)
					messages = messages[:0]
					ticker.Reset(t)
				}
				return nil
			}

			messages = append(messages, msg)

			if len(messages) >= c.batchSize {
				c.broker.processBatch(session.Context(), messages, session)
				ticker.Reset(t)
				messages = messages[:0]
			}

		case <-ticker.C:
			if len(messages) > 0 {
				c.broker.processBatch(session.Context(), messages, session)
				ticker.Reset(t)
				messages = messages[:0]
			}
		}
	}
}

func (b *Broker) processBatch(ctx context.Context, msgs []*sarama.ConsumerMessage, session sarama.ConsumerGroupSession) {
	uuids := make(uuid.UUIDs, 0, len(msgs))

	for _, msg := range msgs {
		u, err := uuid.ParseBytes(msg.Value)
		if err != nil {
			slog.Error("Failed parse uuid", "error", err)
			return
		}
		uuids = append(uuids, u)
	}

	err := b.storage.UpdateOrders(ctx, uuids)
	if err != nil {
		slog.Error("UpdateOrders failed", "error", err)
		return
	}

	for _, msg := range msgs {
		session.MarkMessage(msg, "")
	}

	session.Commit()

	slog.Info("Processed batch", "count", len(uuids))
}
