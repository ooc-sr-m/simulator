package actions

import "simulator/models/capabilities"

var actionDefinitions = [...]ActionDefinition{
	{
		Type:        ActionMove,
		Name:        "Move",
		Description: "Attempt to move to a destination position.",
		Parameters: []ActionParameter{
			{Name: "destination", Type: "position", Required: true, Description: "Position the creature attempts to move to."},
		},
		Tags: []string{"generic", "movement"},
	},
	{
		Type:        ActionRun,
		Name:        "Run",
		Description: "Attempt to run to a destination position.",
		Parameters: []ActionParameter{
			{Name: "destination", Type: "position", Required: true, Description: "Position the creature attempts to run to."},
		},
		Tags: []string{"generic", "movement"},
	},
	{
		Type:        ActionJump,
		Name:        "Jump",
		Description: "Attempt to jump.",
		Parameters:  []ActionParameter{},
		Tags:        []string{"generic", "movement"},
	},
	{
		Type:        ActionFollow,
		Name:        "Follow",
		Description: "Attempt to follow another entity.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: true, Description: "Identifier of the entity the creature attempts to follow."},
		},
		Tags: []string{"generic", "movement", "social"},
	},
	{
		Type:        ActionEat,
		Name:        "Eat",
		Description: "Attempt to consume compatible food.",
		Parameters: []ActionParameter{
			{Name: "item_id", Type: "uuid", Required: true, Description: "Identifier of the food the creature attempts to consume."},
			{Name: "quantity", Type: "number", Required: false, Description: "Amount the creature attempts to consume."},
		},
		Tags: []string{"generic", "physiology"},
	},
	{
		Type:        ActionDrink,
		Name:        "Drink",
		Description: "Attempt to consume a compatible drink.",
		Parameters: []ActionParameter{
			{Name: "item_id", Type: "uuid", Required: true, Description: "Identifier of the drink the creature attempts to consume."},
			{Name: "quantity", Type: "number", Required: false, Description: "Amount the creature attempts to consume."},
		},
		Tags: []string{"generic", "physiology"},
	},
	{
		Type:        ActionSleep,
		Name:        "Sleep",
		Description: "Attempt to sleep for an optional duration.",
		Parameters: []ActionParameter{
			{Name: "duration", Type: "number", Required: false, Description: "Requested sleep duration."},
		},
		Tags: []string{"generic", "physiology"},
	},
	{
		Type:        ActionRest,
		Name:        "Rest",
		Description: "Attempt to rest without assuming the creature sleeps.",
		Parameters: []ActionParameter{
			{Name: "duration", Type: "number", Required: false, Description: "Requested rest duration."},
		},
		Tags: []string{"generic", "physiology"},
	},
	{
		Type:        ActionLook,
		Name:        "Look",
		Description: "Attempt to observe the surroundings or a specific entity.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: false, Description: "Entity the creature attempts to observe, when there is a specific target."},
		},
		Tags: []string{"generic", "perception"},
	},
	{
		Type:        ActionListen,
		Name:        "Listen",
		Description: "Attempt to pay attention to sounds in the surroundings.",
		Parameters:  []ActionParameter{},
		Tags:        []string{"generic", "perception"},
	},
	{
		Type:        ActionSniff,
		Name:        "Sniff",
		Description: "Attempt to perceive scents in the surroundings or from a specific entity.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: false, Description: "Entity the creature attempts to sniff, when there is a specific target."},
		},
		Tags: []string{"generic", "perception"},
	},
	{
		Type:        ActionBite,
		Name:        "Bite",
		Description: "Attempt to bite a target within reach.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: true, Description: "Identifier of the target the creature attempts to bite."},
		},
		Tags: []string{"generic", "combat", "interaction"},
	},
	{
		Type:        ActionLick,
		Name:        "Lick",
		Description: "Attempt to lick a target within reach.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: true, Description: "Identifier of the target the creature attempts to lick."},
		},
		Tags: []string{"generic", "interaction"},
	},
	{
		Type:        ActionPickUp,
		Name:        "Pick up",
		Description: "Attempt to pick up an item.",
		Parameters: []ActionParameter{
			{Name: "item_id", Type: "uuid", Required: true, Description: "Identifier of the item the creature attempts to pick up."},
		},
		Tags: []string{"generic", "interaction"},
	},
	{
		Type:        ActionDrop,
		Name:        "Drop",
		Description: "Attempt to drop an item.",
		Parameters: []ActionParameter{
			{Name: "item_id", Type: "uuid", Required: true, Description: "Identifier of the item the creature attempts to drop."},
		},
		Tags: []string{"generic", "interaction"},
	},
	{
		Type:        ActionDig,
		Name:        "Dig",
		Description: "Attempt to dig at the current position.",
		Parameters:  []ActionParameter{},
		Tags:        []string{"generic", "interaction"},
	},
	{
		Type:        ActionPlay,
		Name:        "Play",
		Description: "Attempt to start or continue a play activity.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: false, Description: "Entity or item the creature attempts to play with."},
		},
		Tags: []string{"generic", "social"},
	},
	{
		Type:        ActionGroom,
		Name:        "Groom",
		Description: "Attempt to groom itself or another entity.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: false, Description: "Entity the creature attempts to groom; when omitted, it attempts to groom itself."},
		},
		Tags: []string{"generic", "social", "physiology"},
	},
	{
		Type:        ActionApproach,
		Name:        "Approach",
		Description: "Attempt to approach an entity.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: true, Description: "Identifier of the entity the creature attempts to approach."},
		},
		Tags: []string{"generic", "movement", "social"},
	},
	{
		Type:        ActionAvoid,
		Name:        "Avoid",
		Description: "Attempt to keep away from an entity.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: true, Description: "Identifier of the entity the creature attempts to avoid."},
		},
		Tags: []string{"generic", "movement"},
	},
	{
		Type:        ActionHide,
		Name:        "Hide",
		Description: "Attempt to hide in the surroundings.",
		Parameters:  []ActionParameter{},
		Tags:        []string{"generic", "movement"},
	},
	{
		Type:        ActionFlee,
		Name:        "Flee",
		Description: "Attempt to move away from a threat or situation.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: false, Description: "Entity the creature attempts to flee from, when there is a specific threat."},
		},
		Tags: []string{"generic", "movement"},
	},
	{
		Type:        ActionDefecate,
		Name:        "Defecate",
		Description: "Attempt to defecate.",
		Parameters:  []ActionParameter{},
		Tags:        []string{"generic", "physiology"},
	},
	{
		Type:          ActionBark,
		Name:          "Bark",
		Description:   "Attempt to bark with an optional intensity.",
		AllowedLabels: capabilities.Set{capabilities.CanBark: true},
		Parameters: []ActionParameter{
			{Name: "intensity", Type: "number", Required: false, Description: "Requested bark intensity."},
		},
		Tags: []string{"expression"},
	},
	{
		Type:        ActionGrowl,
		Name:        "Growl",
		Description: "Attempt to growl, optionally directing it at an entity.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: false, Description: "Entity at which the creature attempts to direct the growl."},
			{Name: "intensity", Type: "number", Required: false, Description: "Requested growl intensity."},
		},
		Tags: []string{"expression"},
	},
	{
		Type:        ActionWhine,
		Name:        "Whine",
		Description: "Attempt to whine with an optional intensity.",
		Parameters: []ActionParameter{
			{Name: "intensity", Type: "number", Required: false, Description: "Requested whine intensity."},
		},
		Tags: []string{"expression"},
	},
	{
		Type:        ActionHowl,
		Name:        "Howl",
		Description: "Attempt to howl with an optional intensity.",
		Parameters: []ActionParameter{
			{Name: "intensity", Type: "number", Required: false, Description: "Requested howl intensity."},
		},
		Tags: []string{"expression"},
	},
	{
		Type:        ActionMarkTerritory,
		Name:        "Mark territory",
		Description: "Attempt to mark an area or item in the surroundings.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: false, Description: "Entity or item the creature attempts to mark, when there is a specific target."},
		},
		Tags: []string{"expression"},
	},
	{
		Type:        ActionSniffRear,
		Name:        "Sniff rear",
		Description: "Attempt to sniff another creature's rear as a social interaction.",
		Parameters: []ActionParameter{
			{Name: "target_id", Type: "uuid", Required: true, Description: "Identifier of the creature whose rear is being sniffed."},
		},
		Tags: []string{"social", "perception"},
	},
}

func AllDefinitions() []ActionDefinition {
	definitions := make([]ActionDefinition, len(actionDefinitions))
	for i, definition := range actionDefinitions {
		definitions[i] = cloneDefinition(definition)
	}
	return definitions
}

func GetDefinition(actionType ActionType) (ActionDefinition, bool) {
	for _, definition := range actionDefinitions {
		if definition.Type == actionType {
			return cloneDefinition(definition), true
		}
	}

	return ActionDefinition{}, false
}

func cloneDefinition(definition ActionDefinition) ActionDefinition {
	clone := definition
	clone.Parameters = cloneSlice(definition.Parameters)
	clone.Tags = cloneSlice(definition.Tags)
	clone.AllowedLabels = cloneCapabilitySet(definition.AllowedLabels)
	clone.DeniedLabels = cloneCapabilitySet(definition.DeniedLabels)
	return clone
}

func cloneSlice[T any](source []T) []T {
	if source == nil {
		return nil
	}

	clone := make([]T, len(source))
	copy(clone, source)
	return clone
}

func cloneCapabilitySet(source capabilities.Set) capabilities.Set {
	if source == nil {
		return nil
	}

	clone := make(capabilities.Set, len(source))
	for label, enabled := range source {
		clone[label] = enabled
	}
	return clone
}
