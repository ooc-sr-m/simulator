type Creature struct {
    ID       uuid.UUID
    Species  string
    Age      int
    Health   int
    Location uuid.UUID
    Position Position
    Senses   Senses
    Hunger  int
}