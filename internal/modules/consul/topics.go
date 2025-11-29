package modules

import (
	"context"
	"fmt"
	"log"

	utils "github.com/cotsom/CloudExec/internal/utils"

	"github.com/segmentio/kafka-go"
)

type Topics struct{}

func (m Topics) RunModule(target string, flags map[string]string, conn *kafka.Conn, dialer *kafka.Dialer) {
	if flags["topic"] != "" {
		var r *kafka.Reader

		if dialer != nil {
			r = kafka.NewReader(kafka.ReaderConfig{
				Brokers:  []string{target},
				Topic:    flags["topic"],
				GroupID:  "consumer-group-id",
				Dialer:   dialer,
				MaxBytes: 10e6,
			})
		} else {
			r = kafka.NewReader(kafka.ReaderConfig{
				Brokers:  []string{target},
				Topic:    flags["topic"],
				GroupID:  "consumer-group-id",
				MaxBytes: 10e6,
				// Partition: 0,
			})
		}

		r.SetOffset(0)
		for {
			t, err := r.ReadMessage(context.Background())
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
