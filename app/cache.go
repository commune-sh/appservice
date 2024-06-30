package app

import (
	config "commune/config"

	"github.com/tidwall/buntdb"
)

type Cache struct {
	PublicRooms *buntdb.DB
}

func NewCache(conf *config.Config) (*Cache, error) {

	db, err := buntdb.Open(":memory:")
	if err != nil {
		panic(err)
	}

	c := &Cache{
		PublicRooms: db,
	}

	return c, nil
}
