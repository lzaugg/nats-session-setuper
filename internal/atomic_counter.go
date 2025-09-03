package internal

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const MaxValue = 99

type AtomicCounter struct {
	js  jetstream.JetStream
	kv  jetstream.KeyValue
	key string
}

func NewAtomicCounter(ctxt context.Context, nc *nats.Conn, bucketName, key string) (*AtomicCounter, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	// Create or get the KV bucket
	kv, err := js.CreateKeyValue(ctxt, jetstream.KeyValueConfig{
		Bucket:  bucketName,
		History: 10,
	})
	if err != nil {
		return nil, err
	}

	return &AtomicCounter{
		js:  js,
		kv:  kv,
		key: key,
	}, nil
}

// GetNextValue atomically increments and returns the next value
func (ac *AtomicCounter) GetNextValue(ctxt context.Context) (int64, error) {
	for {
		// Get current value with a lock
		entry, err := ac.kv.Get(ctxt, ac.key)
		if err != nil {
			// If key doesn't exist, start at 0
			if errors.Is(err, jetstream.ErrKeyNotFound) {
				return ac.initializeCounter(ctxt, 0)
			}
			return 0, err
		}

		currentValue, err := strconv.ParseInt(string(entry.Value()), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid counter value: %v", err)
		}

		// Try to update with optimistic locking
		newValue := currentValue + 1
		_, err = ac.kv.Update(ctxt,
			ac.key,
			[]byte(strconv.FormatInt(newValue, 10)),
			entry.Revision())

		if err != nil {
			// If the update failed due to a conflict (another process updated it)
			// Retry the operation
			time.Sleep(time.Millisecond * 10)
			continue
			//}
			//return 0, err
		}

		if newValue > MaxValue {
			return 0, fmt.Errorf("counter has reached the maximum value")
		}

		return newValue, nil
	}
}

// Initialize counter with a starting value
func (ac *AtomicCounter) initializeCounter(ctxt context.Context, initialValue int64) (int64, error) {
	_, err := ac.kv.Put(ctxt, ac.key, []byte(strconv.FormatInt(initialValue, 10)))
	if err != nil {
		return 0, err
	}
	return initialValue, nil
}

// GetCurrentValue returns the current value without incrementing
func (ac *AtomicCounter) GetCurrentValue(ctxt context.Context) (int64, error) {
	entry, err := ac.kv.Get(ctxt, ac.key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return 0, nil
		}
		return 0, err
	}

	currentValue, err := strconv.ParseInt(string(entry.Value()), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid counter value: %v", err)
	}

	return currentValue, nil
}
