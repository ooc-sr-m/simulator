package dog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

//go:embed dog-names.json
var namesJSON []byte

var (
	loadNamesOnce sync.Once
	loadedNames   []string
	loadNamesErr  error
)

// Names returns every valid name embedded from dog-names.json. The file is
// decoded only once and callers receive a copy so the shared catalog cannot be
// modified accidentally.
func Names() ([]string, error) {
	names, err := nameCatalog()
	if err != nil {
		return nil, err
	}
	return append([]string(nil), names...), nil
}

func nameCatalog() ([]string, error) {
	loadNamesOnce.Do(func() {
		if err := json.Unmarshal(namesJSON, &loadedNames); err != nil {
			loadNamesErr = fmt.Errorf("decode embedded dog names: %w", err)
			return
		}

		validNames := loadedNames[:0]
		for _, name := range loadedNames {
			name = strings.TrimSpace(name)
			if name != "" {
				validNames = append(validNames, name)
			}
		}
		loadedNames = validNames

		if len(loadedNames) == 0 {
			loadNamesErr = fmt.Errorf("embedded dog name catalog is empty")
		}
	})

	if loadNamesErr != nil {
		return nil, loadNamesErr
	}
	return loadedNames, nil
}

// RandomName picks a name by using a random index in the embedded catalog.
func RandomName(random *rand.Rand) (string, error) {
	if random == nil {
		return "", fmt.Errorf("dog name random generator is not initialized")
	}

	names, err := nameCatalog()
	if err != nil {
		return "", err
	}
	return names[random.Intn(len(names))], nil
}
