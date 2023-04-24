package cli

import (
	"encoding/json"
	"os"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// PointRequests defines a repeated slice of point objects.
type PointRequests struct {
	Points []types.Point
}

func parsePoints(pointsFile string) ([]types.Point, error) {
	points := PointRequests{}

	if pointsFile == "" {
		return points.Points, nil
	}

	contents, err := os.ReadFile(pointsFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &points)
	if err != nil {
		return nil, err
	}

	return points.Points, nil
}
