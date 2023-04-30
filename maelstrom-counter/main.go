package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func main() {
	log.SetOutput(os.Stderr)
	lock := sync.Mutex{}

	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, nodeId := range n.NodeIDs() {
					if nodeId == n.ID() {
						continue
					}
					propagate := make(map[string]any)

					propagate["type"] = "propagate"
					counters := make(map[string]int)
					for _, nodeId := range n.NodeIDs() {
						val, _ := kv.ReadInt(context.Background(), nodeId)
						counters[nodeId] = int(val)
					}
					propagate["counters"] = counters

					n.Send(nodeId, propagate)
				}
			}
		}
	}()

	n.Handle("add", func(msg maelstrom.Message) error {
		log.Println("handling add")
		var request map[string]any
		if err := json.Unmarshal(msg.Body, &request); err != nil {
			return err
		}

		delta := int(request["delta"].(float64))
		log.Println("delta received", delta)

		current, _ := kv.ReadInt(context.Background(), n.ID())
		to := current + delta

		kv.Write(context.Background(), n.ID(), to)

		response := make(map[string]any)
		response["type"] = "add_ok"
		return n.Reply(msg, response)
	})

	n.Handle("propagate", func(msg maelstrom.Message) error {
		lock.Lock()
		defer lock.Unlock()

		log.Println("handling propagate")
		var request map[string]any
		if err := json.Unmarshal(msg.Body, &request); err != nil {
			return err
		}

		counters := request["counters"].(map[string]interface{})

		for nodeId, value := range counters {
			mine, _ := kv.ReadInt(context.Background(), nodeId)
			their := int(value.(float64))
			kv.Write(context.Background(), nodeId, maxInt(mine, their))
		}

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		response := make(map[string]any)
		response["type"] = "read_ok"
		sum := 0
		for _, nodeId := range n.NodeIDs() {
			val, _ := kv.ReadInt(context.Background(), nodeId)
			sum += val
		}
		response["value"] = sum
		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
