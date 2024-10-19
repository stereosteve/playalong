package repo

import (
	"log"

	"go.etcd.io/bbolt"
)

var (
	db         *bbolt.DB
	SongBucket = []byte("songs")
)

func Dial() {
	var err error
	db, err = bbolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(SongBucket)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
}

func Close() {
	db.Close()

}
