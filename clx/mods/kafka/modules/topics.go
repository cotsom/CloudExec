package modules

import (
	"clx/utils"
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Topics struct{}

func (m Topics) RunModule(target string, flags map[string]string, conn *kafka.Conn) {
	if flags["topic"] != "" {
		// make a new reader that consumes from topic-A, partition 0, at offset 42
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{target},
			Topic:     flags["topic"],
			MaxBytes:  1,
			Partition: 0,
		})
		r.SetOffset(0)
		for {
			fmt.Println("qwe")
			t, err := r.ReadMessage(context.Background())
			fmt.Println("qwe")
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Printf("message at offset %d: %s = %s\n", t.Offset, string(t.Key), string(t.Value))
		}

		if err := r.Close(); err != nil {
			log.Fatal("failed to close reader:", err)
		}
	} else {
		partitions, err := conn.ReadPartitions()
		if err != nil {
			panic(err.Error())
		}

		t := map[string]struct{}{}

		for _, p := range partitions {
			t[p.Topic] = struct{}{}
		}
		for k := range t {
			fmt.Println(utils.ClearLine, k)
		}
	}
}
