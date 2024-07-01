package app

import (
	config "commune/config"

	"github.com/tidwall/buntdb"
)

type Cache struct {
	JoinedRooms *buntdb.DB
}

func NewCache(conf *config.Config) (*Cache, error) {

	db, err := buntdb.Open(":memory:")
	if err != nil {
		panic(err)
	}

	c := &Cache{
		JoinedRooms: db,
	}

	return c, nil
}
