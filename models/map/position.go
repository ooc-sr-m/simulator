type Position struct {
    X float64
    Y float64
}

func (p Position) DistanceTo(other Position) float64 {
    dx := p.X - other.X
    dy := p.Y - other.Y

    return math.Hypot(dx, dy)
}
