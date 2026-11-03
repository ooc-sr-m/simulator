type Memory struct {
    Capacity        int
    MinimumWeight   int
    Entries         []MemoryEntry
}

type MemoryEntry struct {
	ID        uuid.UUID
    Type      MemoryType
    Content   string
    Weight    int
    CreatedAt time.Time
}
