package dmdb

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
)

const (
	dbLocation = "db"
)

var (
	once sync.Once
	db   *badger.DB
)

func initDB() {
	once.Do(func() {
		opts := badger.DefaultOptions
		opts.Dir = dbLocation
		opts.ValueDir = dbLocation
		opts.ValueLogLoadingMode = options.FileIO //maybe?

		var err error
		log.L.Infof("Opening local database at: %s", dbLocation)

		db, err = badger.Open(opts)
		if err != nil {
			log.L.Fatalf("failed to open badger db: %s", err)
		}

		// close the database connection if the program is killed
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			log.L.Infof("Closing local database.")
			err := db.Close()
			if err != nil {
				log.L.Warnf("failed to close local database: %s", err)
			}

			os.Exit(0)
		}()
	})
}

// Put puts a key/value into the database
func Put(key string, value []byte) *nerr.E {
	initDB()

	err := db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
	if err != nil {
		return nerr.Translate(err).Addf("failed to put key %s into the database", key)
	}

	return nil
}

// Get returns a value at the given key
func Get(key string) ([]byte, *nerr.E) {
	initDB()
	value := []byte{}

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return value, nil
		}

		return value, nerr.Translate(err).Addf("failed to get '%s' from the local database", key)
	}

	return value, nil
}
