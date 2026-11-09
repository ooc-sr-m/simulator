package worldmap

type Chunk struct {
	GridX       int
	GridY       int
	Width       float64
	Height      float64
	PixelWidth  int
	PixelHeight int
	EntityIDs   []string
}
