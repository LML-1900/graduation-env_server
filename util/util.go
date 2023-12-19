package util

import (
	"env_server/data"
	"math"
	"strconv"
	"strings"
)

func LonLatPositionToTileId(position data.LonLatPosition, level int) string {
	n := math.Pow(2, float64(level+1))
	X := (position.Longitude + 180) / 360 * n
	Y := (position.Latitude + 90) / 360 * n
	x := int(math.Floor(X))
	y := int(math.Floor(Y))
	builder := strings.Builder{}
	builder.WriteString(strconv.Itoa(level))
	builder.WriteString("-")
	builder.WriteString(strconv.Itoa(x))
	builder.WriteString("-")
	builder.WriteString(strconv.Itoa(y))
	return builder.String()
}
