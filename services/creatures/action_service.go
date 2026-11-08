package creatures

import (
	"simulator/models/actions"
	"simulator/models/capabilities"
	"simulator/models/creatures"
)

type ActionService struct{}

func (ActionService) GetActions(creature *creatures.Creature) []actions.ActionDefinition {
	if creature == nil {
		return nil
	}

	catalog := actions.AllDefinitions()
	compatible := make([]actions.ActionDefinition, 0, len(catalog))
	for _, definition := range catalog {
		if IsActionCompatible(creature.Capabilities, definition) {
			compatible = append(compatible, definition)
		}
	}
	return compatible
}

func IsActionCompatible(creatureLabels capabilities.Set, definition actions.ActionDefinition) bool {
	for label, denied := range definition.DeniedLabels {
		if denied && creatureLabels[label] {
			return false
		}
	}

	hasAllowedLabel := false
	for label, allowed := range definition.AllowedLabels {
		if !allowed {
			continue
		}

		hasAllowedLabel = true
		if creatureLabels[label] {
			return true
		}
	}

	return !hasAllowedLabel
}
