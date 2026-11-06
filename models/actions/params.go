package actions

import (
	"simulator/models/actions/core"
	"simulator/models/actions/expression"
	"simulator/models/actions/movement"
	"simulator/models/actions/physiology"
)

type TargetParams = core.TargetParams
type ItemParams = core.ItemParams
type MoveParams = movement.MoveParams
type EatParams = physiology.EatParams
type DrinkParams = physiology.DrinkParams
type SleepParams = physiology.SleepParams
type BarkParams = expression.BarkParams
