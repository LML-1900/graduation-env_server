package data

import "time"

type DemData struct {
	TileId   string    `bson:"tile_id"`
	Level    int       `bson:"level"`
	Content  []byte    `bson:"content"`
	InsertAt time.Time `bson:"insert_at"`
}

type LonLatPosition struct {
	Longitude float64 `bson:"longitude"`
	Latitude  float64 `bson:"latitude"`
}

type Crater struct {
	Position LonLatPosition
	Width    float64 `bson:"width"`
	Depth    float64 `bson:"depth"`
}
