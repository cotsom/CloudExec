package modules

import (
	"clx/utils"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Topics struct{}

func (m Topics) RunModule(target string, flags map[string]string, conn *kafka.Conn) {
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
