package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const (
	workerIDBits = uint64(10) // 5bit workerID out of 10bit worker machine ID
	sequenceBits = uint64(12)

	maxSequence = -1 ^ (-1 << sequenceBits)
	maxWorkerID = -1 ^ (-1 << workerIDBits)

	timeLeft = uint8(22) // timeLeft = workerIDBits + sequenceBits // Timestamp offset left
	workLeft = uint8(12) // workLeft = sequenceBits // Node IDx offset to the left

	epoch = int64(1262304000000) // constant timestamp (milliseconds)
)

type Snowflake struct {
	LastStamp    int64
	WorkerID     int64
	Sequence     int64
	SequenceLock sync.Mutex
}

func (snowflake *Snowflake) GenerateUID() uint64 {
	snowflake.SequenceLock.Lock()
	defer snowflake.SequenceLock.Unlock()

	timeStamp := getMilliSeconds()
	if timeStamp < snowflake.LastStamp {
		panic(fmt.Sprintf("Time moved backwards: %d < %d", timeStamp, snowflake.LastStamp))
	}

	if snowflake.LastStamp == timeStamp {
		snowflake.Sequence = (snowflake.Sequence + 1) & maxSequence
		if snowflake.Sequence == 0 {
			for timeStamp <= snowflake.LastStamp {
				timeStamp = getMilliSeconds()
			}
		}
	} else {
		snowflake.Sequence = 0
	}

	snowflake.LastStamp = timeStamp

	id := ((timeStamp - epoch) << timeLeft) |
		(snowflake.WorkerID << workLeft) |
		snowflake.Sequence

	return uint64(id)
}

func generateUUID() string {
	id := uuid.New()
	return id.String()
}

func getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1000000
}

func getStringHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// generateUID generates a unique ID using the worker ID and the current time
// implementation of the snowflake algorithm
func main() {
	log.SetOutput(os.Stderr)

	n := maelstrom.NewNode()
	worker := Snowflake{
		WorkerID: 0,
	}

	// { "src": "", "dest": "", "body": { "type": "generate" } }
	n.Handle("generate", func(msg maelstrom.Message) error {
		workerID := 0
		for i := 0; i < len(n.NodeIDs()); i++ {
			if n.NodeIDs()[i] == n.ID() {
				workerID = i
				break
			}
		}
		worker.WorkerID = int64(workerID)

		body := make(map[string]any)

		body["type"] = "generate_ok"
		body["id"] = strconv.FormatUint(worker.GenerateUID(), 10)

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
