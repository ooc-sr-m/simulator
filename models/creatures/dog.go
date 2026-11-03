package creatures

type Dog struct {
    Creature
s
    Breed 	string
	Name 	*string
	Father	*uuid.UUID
	Mother 	*uuid.UUID
   	Personality cognitive_functions.Personality
}