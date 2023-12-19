package data

import "time"

type DemData struct {
	TileId   string    `bson:"tile_id"`
	Level    int       `bson:"level"`
	Content  []byte    `bson:"content"`
	InsertAt time.Time `bson:"insert_at"`
}
