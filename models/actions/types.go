package actions

import (
	"simulator/models/actions/combat"
	"simulator/models/actions/expression"
	"simulator/models/actions/interaction"
	"simulator/models/actions/movement"
	"simulator/models/actions/perception"
	"simulator/models/actions/physiology"
	"simulator/models/actions/social"
)

const (
	ActionMove          = movement.ActionMove
	ActionRun           = movement.ActionRun
	ActionJump          = movement.ActionJump
	ActionFollow        = movement.ActionFollow
	ActionEat           = physiology.ActionEat
	ActionDrink         = physiology.ActionDrink
	ActionSleep         = physiology.ActionSleep
	ActionRest          = physiology.ActionRest
	ActionLook          = perception.ActionLook
	ActionListen        = perception.ActionListen
	ActionSniff         = perception.ActionSniff
	ActionBite          = combat.ActionBite
	ActionLick          = interaction.ActionLick
	ActionPickUp        = interaction.ActionPickUp
	ActionDrop          = interaction.ActionDrop
	ActionDig           = interaction.ActionDig
	ActionPlay          = social.ActionPlay
	ActionGroom         = social.ActionGroom
	ActionApproach      = movement.ActionApproach
	ActionAvoid         = movement.ActionAvoid
	ActionHide          = movement.ActionHide
	ActionFlee          = movement.ActionFlee
	ActionDefecate      = physiology.ActionDefecate
	ActionBark          = expression.ActionBark
	ActionGrowl         = expression.ActionGrowl
	ActionWhine         = expression.ActionWhine
	ActionHowl          = expression.ActionHowl
	ActionMarkTerritory = expression.ActionMarkTerritory
	ActionSniffRear     = social.ActionSniffRear
)
