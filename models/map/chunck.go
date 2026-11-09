package worldmap

import "github.com/google/uuid"

type Chunk struct {
	ID          uuid.UUID
	GridX       int
	GridY       int
	Width       float64
	Height      float64
	PixelWidth  int
	PixelHeight int
	EntityIDs   []string
}
