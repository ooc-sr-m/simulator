package movement

import worldmap "simulator/models/map"

type MoveParams struct {
	Destination worldmap.Position `json:"destination"`
}
