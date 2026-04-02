package broker

import (
	"log"
	"os"

	"github.com/IBM/sarama"
)

func NewBroker(topic string) *Broker {
	brokers := []string{os.Getenv("KAFKA_CONN")}

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0


	config.Producer.RequiredAcks = sarama.WaitForAll 
	config.Producer.Retry.Max = 5          
	config.Producer.Return.Successes = true

	config.Producer.Flush.MaxMessages = 1000                 
	config.Producer.Compression = sarama.CompressionSnappy 
	config.Net.MaxOpenRequests = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatal("Failed to start Kafka producer:", err)
	}

	return &Broker{
		topic:    topic,
		producer: producer,
	}
}

type Broker struct {
	topic    string
	producer sarama.SyncProducer
}

func (b *Broker) Publish(messages [][]byte) error {
	msgs := make([]*sarama.ProducerMessage, 0, len(messages))
	for _, message := range messages {
		msgs = append(msgs, &sarama.ProducerMessage{
			Topic: b.topic,
			Value: sarama.ByteEncoder(message),
		})
	}

	err := b.producer.SendMessages(msgs)
	if err != nil {
		return err
	}

	return nil
}

func (b *Broker) Close() {
	err := b.producer.Close()
	if err != nil {
		log.Println("Error closing Kafka producer:", err)
	}
}
