type Creature struct {
    ID       uuid.UUID
    Species  string
    Age      int
    Location uuid.UUID
    Position Position

    Senses   Senses
    Mobility Mobility
}