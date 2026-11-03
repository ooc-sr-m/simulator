type GoalStatus string

const (
    GoalActive    GoalStatus = "active"
    GoalCompleted GoalStatus = "completed"
    GoalFailed    GoalStatus = "failed"
    GoalSuspended GoalStatus = "suspended"
)

type Goal struct {
    ID          uuid.UUID
    Description string
    Priority    int
    Status      GoalStatus
    Progress    float64
}
