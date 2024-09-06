package badger

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

type void struct{}

type set map[int64]void

func (s set) Add(element int64) (exists bool) {
	if _, exists = s[element]; !exists {
		s[element] = void{}
	}

	return exists
}

func (s set) Values() []int64 {
	values := make([]int64, 0, len(s))
	for k := range s {
		values = append(values, k)
	}

	return values
}

type DB struct {
	db *badger.DB
}

type TableKey []byte

var (
	debugStatusKey       = TableKey("debug_status")
	subscriptionsProdKey = TableKey("subscriptions_prod")
)

func NewDB(dbPath string) (*DB, error) {
	// It will be created if it doesn't exist.
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return errors.Join(d.db.Sync(), d.db.Close())
}

func (d *DB) AddSubscriber(id int64) error {
	err := d.db.Update(func(tx *badger.Txn) error {
		item, err := tx.Get(subscriptionsProdKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("error getting subscriptions: %w", err)
		}

		var subs set

		if err == nil {
			var itemValue []byte
			_ = item.Value(func(v []byte) error {
				itemValue = append(itemValue, v...)
				return nil
			})

			if err := json.Unmarshal(itemValue, &subs); err != nil {
				return fmt.Errorf("error unmarshalling item: %w", err)
			}

		} else {
			subs = make(set)
		}

		subs.Add(id)
		newItemValue, err := json.Marshal(subs)
		if err != nil {
			return fmt.Errorf("error marshalling new item: %w", err)
		}

		err = tx.Set(subscriptionsProdKey, newItemValue)
		if err != nil {
			return fmt.Errorf("error setting subscriptions: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error adding subscription for user ID [%d]: %w", id, err)
	}

	return nil
}

func (d *DB) RemoveSubscriber(id int64) error {
	err := d.db.Update(func(tx *badger.Txn) error {
		item, err := tx.Get(subscriptionsProdKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("error getting subscriptions: %w", err)
		}

		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}

		var itemValue []byte
		_ = item.Value(func(v []byte) error {
			itemValue = append(itemValue, v...)
			return nil
		})

		var subs set
		if err := json.Unmarshal(itemValue, &subs); err != nil {
			return fmt.Errorf("error unmarshalling item: %w", err)
		}

		delete(subs, id)

		newItemValue, err := json.Marshal(subs)
		if err != nil {
			return fmt.Errorf("error marshalling new item: %w", err)
		}

		err = tx.Set(subscriptionsProdKey, newItemValue)
		if err != nil {
			return fmt.Errorf("error setting subscriptions: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error removing subscription for user ID [%d]: %w", id, err)
	}

	return nil
}

func (d *DB) Subscribers() ([]int64, error) {
	var subs set

	err := d.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(subscriptionsProdKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("error getting subscriptions: %w", err)
		}

		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}

		var itemValue []byte
		_ = item.Value(func(v []byte) error {
			itemValue = append(itemValue, v...)
			return nil
		})

		return json.Unmarshal(itemValue, &subs)
	})
	if err != nil {
		return nil, fmt.Errorf("error on DB transaction seeking for subscribers: %w", err)
	}

	return subs.Values(), nil
}

func (d *DB) ManageDebug(enable bool) error {
	err := d.db.Update(func(tx *badger.Txn) error {
		newItemValue, err := json.Marshal(enable)
		if err != nil {
			return fmt.Errorf("error marshalling new item: %w", err)
		}

		err = tx.Set(debugStatusKey, newItemValue)
		if err != nil {
			return fmt.Errorf("error setting subscriptions: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error setting debug status: %w", err)
	}

	return nil
}

func (d *DB) DebugEnabled() (bool, error) {
	var debugEnabled = false

	err := d.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(debugStatusKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("error getting subscriptions: %w", err)
		}

		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}

		var itemValue []byte
		_ = item.Value(func(v []byte) error {
			itemValue = append(itemValue, v...)
			return nil
		})

		return json.Unmarshal(itemValue, &debugEnabled)
	})
	if err != nil {
		return false, fmt.Errorf("error on DB transaction getting debug status: %w", err)
	}

	return debugEnabled, nil
}
