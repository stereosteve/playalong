package repo

import (
	"encoding/json"

	"go.etcd.io/bbolt"
)

type Song struct {
	ID    string
	Title string
	Genre string
	Key   string
	BPM   string
	Stems []StemFile
}

type StemFile struct {
	Name string
	Path string
}

//
// DB STUFF
//

func SaveSong(song *Song) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SongBucket)
		j, err := json.Marshal(song)
		if err != nil {
			return err
		}
		return b.Put([]byte(song.ID), j)
	})
}

func ListSongs() ([]*Song, error) {
	all := []*Song{}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SongBucket)
		return b.ForEach(func(k, v []byte) error {
			var p *Song
			err := json.Unmarshal(v, &p)
			if err != nil {
				return err
			}
			all = append(all, p)
			return nil
		})
	})
	return all, err
}

func GetSong(id string) (*Song, error) {
	var p *Song
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SongBucket)
		v := b.Get([]byte(id))
		return json.Unmarshal(v, &p)
	})
	return p, err
}
