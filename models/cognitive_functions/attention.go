type AttentionSource string

const (
    AttentionVision  AttentionSource = "vision"
    AttentionHearing AttentionSource = "hearing"
    AttentionSmell   AttentionSource = "smell"
    AttentionThought AttentionSource = "thought"
)

type AttentionTarget struct {
    TargetID uuid.UUID
    Source   AttentionSource
    Weight   int
}

type Attention struct {
    Capacity int
    Targets  []AttentionTarget
}