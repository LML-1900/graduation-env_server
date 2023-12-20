package util

import (
	"env_server/data"
	pb "env_server/grpc_env_service"
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
	return generateTileId(level, x, y)
}

func LonLatAreaToTileIds(area *pb.Area, level int) (tileIds []string) {
	minlon := area.Bottomleft.Longitude
	minlat := area.Bottomleft.Latitude
	maxlon := area.Topright.Longitude
	maxlat := area.Topright.Latitude
	n := math.Pow(2, float64(level+1))
	xmin := int(math.Floor((minlon + 180) / 360 * n))
	ymin := int(math.Floor((minlat + 90) / 360 * n))
	xmax := int(math.Floor((maxlon + 180) / 360 * n))
	ymax := int(math.Floor((maxlat + 90) / 360 * n))

	for i := xmin; i <= xmax; i++ {
		for j := ymin; j <= ymax; j++ {
			tileIds = append(tileIds, generateTileId(level, i, j))
		}
	}
	return tileIds
}

func generateTileId(level, x, y int) string {
	builder := strings.Builder{}
	builder.WriteString(strconv.Itoa(level))
	builder.WriteString("-")
	builder.WriteString(strconv.Itoa(x))
	builder.WriteString("-")
	builder.WriteString(strconv.Itoa(y))
	return builder.String()
}
