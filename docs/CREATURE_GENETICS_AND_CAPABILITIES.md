# Simulator — Especies, genoma, capacidades y contrato de acciones

> **Especificación para Codex**. Proyecto Go `simulator`. Este documento complementa `ACTION_STATE_LIFECYCLE.md`; no lo reemplaza.
>
> **Importante**: inspecciona primero el repositorio real, especialmente `go.mod`, `models/creatures`, `models/physiological_capacities`, `models/cognitive_functions` y `models/actions/core`. Adapta nombres e imports a los tipos existentes. No reescribas modelos ni implementes funcionalidades fuera del alcance sin necesidad.

## 1. Propósito

Separar tres niveles de datos:

1. **Definición de especie (`SpeciesDefinition`)**: capacidades típicas, restricciones y rangos admisibles para generar individuos (ej.: perros, humanos).
2. **Genoma individual (`Genome`)**: rasgos relativamente estables que se generan dentro de esos rangos. Es un **modelo simplificado de juego**, no una representación biológica real.
3. **Estado de la criatura (`Creature`)**: valores cambiantes (hambre, sed, cansancio, salud, emoción, posición, etc.).

El genoma **no** sustituye al estado ni a las reglas de ejecución. Debe ayudar a inicializar atributos y capacidades; enfermedades, lesiones, aprendizaje y efectos mágicos pueden modificar la capacidad efectiva sin cambiar el genoma.

Objetivo principal: obtener un conjunto de acciones **compatibles** con una criatura y validar, en un segundo paso, cuáles puede **iniciar ahora**.

## 2. Arquitectura existente que debe respetarse

```text
simulator/
├── adapters/
├── cmd/
├── models/
│   ├── creatures/
│   │   ├── creature.go
│   │   ├── dog.go
│   │   └── human.go
│   ├── actions/
│   │   └── core/                 # Ya existe ActionDefinition
│   ├── cognitive_functions/
│   ├── physiological_capacities/
│   └── map/
└── services/
```

No introducir un directorio `internal/`, un `Cognition` obligatorio, un ORM ni un motor completo de simulación.

## 3. Entidades y responsabilidades

### `SpeciesDefinition`

Registro compartido por especie. Contiene un `SpeciesID` estable, límites de rasgos fisiológicos, características típicas de capacidades y etiquetas base. No contiene el hambre actual de un perro.

### `Genome`

Se genera por individuo (posiblemente usando una semilla reproducible). Contiene valores propios dentro de los rangos de su especie, por ejemplo su tasa metabólica o límites individuales. Puede derivar etiquetas inherentes. No es una cadena opaca de ADN que deba interpretar cada acción.

### `Creature`

Conserva su estructura actual y sus capacidades opcionales mediante punteros. Incorporar referencia a especie y genoma **solo tras revisar compatibilidad con el modelo existente**. El estado actual debe mantenerse independiente de límites y predisposiciones.

### `CapabilitySet`

Conjunto final de etiquetas efectivas de una criatura. Resulta de combinar etiquetas de especie, genoma y modificaciones de estado/equipamiento, sin confundir etiquetas con atributos numéricos. Es recomendable almacenarlas como `map[string]bool` o `map[string]struct{}` según la API existente; para una interfaz serializable simple, `map[string]bool` es suficiente. No poner una copia completa del catálogo en cada criatura.

**Distinción de etiquetas:**
- Etiquetas descriptivas (`dog`, `animal`) clasifican especies.
- Etiquetas de capacidad (`can_bark`, `can_walk`, `can_sleep`) indican aptitudes.
- Etiquetas de situación (`immobilized`, `unconscious`) son temporales y **no son genéticas**. Preferiblemente las reglas leen directamente el estado para decidir si una acción es ejecutable.

No deducir automáticamente `can_bark` simplemente porque aparece `dog`, salvo que el perfil de especie lo declare. Esto facilita especies fantásticas.

## 4. Definiciones Go ilustrativas (ajustar al repositorio real)

```go
package creatures

type FloatRange struct {
    Min float64 `json:"min"`
    Max float64 `json:"max"`
}

type SpeciesDefinition struct {
    ID             string            `json:"id"`
    Labels         map[string]bool   `json:"labels"`
    HungerMax      FloatRange        `json:"hunger_max"`
    HungerRate     FloatRange        `json:"hunger_rate"`
    // Ampliar con otros rasgos SOLO cuando el modelo real los necesite.
}

type Genome struct {
    SpeciesID      string            `json:"species_id"`
    Labels         map[string]bool   `json:"labels"`
    HungerMax      float64           `json:"hunger_max"`
    HungerRate     float64           `json:"hunger_rate"`
}
```

Las definiciones son **ejemplos**, no una orden de duplicar campos existentes. Si `Metabolism.Hunger` es `int`, conservarlo y concretar las unidades/conversión cuando se implemente el generador. No crear `HungerMax` en varios lugares como fuentes independientes de verdad. Escoger una única fuente autoritativa para el límite individual.

### Validación del perfil y generación

- Rechazar rangos con `Min > Max`, tasas negativas donde no correspondan, identificadores vacíos y especies inexistentes.
- Generar individuos con aleatoriedad inyectable/semilla para pruebas reproducibles; no utilizar un generador global opaco.
- Nunca modificar `SpeciesDefinition` cuando cambia el hambre de una criatura.
- Permitir características opcionales: ausencia de metabolismo (`nil`) no significa hambre cero.
- Prever validación de combinaciones imposibles (`HungerMax < Hunger` al inicializar) sin inventar reglas adicionales.
- Si los datos vienen de JSON/YAML, validar antes de crear criaturas.

## 5. Asociación de acciones a las criaturas

El catálogo de `ActionDefinition` es **único y compartido**. No crear una clase de acción nueva por especie ni asignar una lista duplicada de acciones a cada perro.

Mantener `Tags` de las acciones (ej.: `generic`, `physiology`) **solo para clasificación**, separados de los requisitos de compatibilidad.

Ejemplo de ampliación de `models/actions/core`:

```go
type ActionDefinition struct {
    Type          ActionType
    Name          string
    Description   string
    Parameters    []ActionParameter
    Tags          []string
    AllowedLabels map[string]bool
    DeniedLabels  map[string]bool
}
```

Si `ActionDefinition` ya tiene otros campos, **conservarlos**. Usar mapas para permitidas/prohibidas conforme a la decisión de diseño; no renombrar `Tags` a `AllowedLabels`.

### Semántica de compatibilidad propuesta

- `AllowedLabels` vacío/nil: no impone restricción positiva.
- Si hay etiquetas permitidas activas, basta con **al menos una coincidencia** (OR).
- Si hay alguna coincidencia con etiquetas prohibidas, **se deniega**, incluso cuando también exista una permitida.
- `false` no cuenta como etiqueta presente, tanto en criatura como en definición.
- Para condiciones AND (ej.: `can_grasp` **y** `can_speak`), no abusar del mapa OR; usar más adelante requisitos explícitos o un predicado de validación. Documentar esa limitación.

**Ejemplos**:

```go
// Bark: perro o criatura ficticia que tenga esta capacidad.
AllowedLabels: map[string]bool{"can_bark": true},

// Move: acción genérica; necesitará validación de movilidad y estado.
AllowedLabels: nil,

// Sleep: genérica, pero se valida que pueda dormir en ese momento.
AllowedLabels: nil,
```

`DeniedLabels` puede usarse para exclusiones declarativas puntuales, pero **no** debe sustituir validación fisiológica, disponibilidad de recursos, objetivos, distancias o herramientas.

## 6. API conceptual: `GetActions` no es `CanExecute`

```go
// Selección estable desde el catálogo compartido.
func GetActions(creature CreatureView, catalog []ActionDefinition) []ActionDefinition

// Comprobación contextual inmediatamente antes de iniciar.
func CanExecute(creature CreatureView, action ActionDefinition, params ActionParams, world WorldView) ValidationResult
```

Las firmas concretas se adaptarán a los tipos ya existentes; `CreatureView` y `WorldView` son conceptos, no modelos nuevos obligatorios.

**`GetActions`**: filtra por capacidades/etiquetas inherentes y efectivas; no modifica estado. Es válido incluir `Bark` aunque el perro esté dormido, si solo queremos conocer su repertorio general.

**`CanExecute`**: comprueba estado actual, parámetros, recursos disponibles y requisitos de la acción. Puede rechazar `Bark` por inconsciencia o `Eat` por falta de comida.

Cuando el estado cambia, volver a validar **antes de iniciar** y **durante la ejecución**. Una caché de compatibilidad es opcional y requiere invalidación si cambian las capacidades efectivas.

## 7. Contrato de comportamiento de las acciones (CLAVE para Codex)

Se debe distinguir con claridad:

| Concepto | Responsabilidad |
|---|---|
| `ActionDefinition` | Qué acción existe, qué significa, parámetros y restricciones declarativas; **compartida e inmutable** |
| `ActionRequest` | La intención concreta de una criatura, con `CreatureID`, tipo y argumentos |
| `ActiveAction` | Instancia activa: inicio, tiempo transcurrido, estado, progreso y datos necesarios |
| `ActionEffect` | Cambio de atributo o recurso que puede producir la acción durante su ejecución |
| `ActionService` | Validación, inicio, actualización y finalización/interrupción |
| `DecisionService` (futuro) | Elección de acción; nunca mutación directa del mundo |

**No ejecutar la acción al construir `ActionDefinition`.** No aplicar todos los efectos cuando se crea `ActionRequest` ni únicamente al finalizar `ActiveAction`.

### Ciclo esperado

1. `GetActions(creature)` obtiene repertorio compatible.
2. La criatura (reglas/Utility AI/LLM futuro) propone un `ActionRequest` con parámetros concretos.
3. `CanExecute` valida y se inicia una `ActiveAction`.
4. En cada actualización, con `dt` simulado, se calcula primero la evolución natural y después se aplican los efectos de la acción activa.
5. Cada actualización respeta límites, consume recursos **reales** y verifica condiciones de continuación.
6. Al alcanzar duración/condición, acción `Completed`, `Interrupted` o `Failed`; se desencadena nueva decisión.
7. Cuando proceda, se emiten eventos de hechos observables; los eventos NO sustituyen el cambio de estado.

```text
Creature + Species/Genome + estado actual
             |
        GetActions (catálogo)
             |
     ActionRequest elegido
             |
     CanExecute / Start
             |
       ActiveAction
             |
       Update(dt)
         |       |
  Cambio natural  Efectos acción
         \       /
        Estado actualizado
             |
    ¿Continuar / terminar?
             |
      Siguiente decisión
```

### Ejemplo `Sleep`

- Repertorio: permitido si el individuo posee la capacidad de dormir; no asumirlo para gólems.
- Ejecución: necesita estar en un estado que permita dormir.
- Durante `Update(dt)`: disminuye fatiga según tasa vigente, mientras hambre/sed pueden aumentar por evolución natural.
- Finaliza al cumplirse duración o condición definida; puede interrumpirse.
- No emitir un evento global por cada décima de cansancio modificada.

### Ejemplo `Eat`

- Repertorio: `Eat` puede ser genérica, no exclusiva de perros/humanos.
- Ejecución: validar alimento concreto y su compatibilidad alimentaria; **permitir la acción no implica que cualquier comida sea válida**.
- Durante `Update(dt)`: consumir cantidad disponible y reducir hambre en proporción a lo realmente ingerido.
- Cuando no queda alimento, terminar/interrumpir según la causa; no generar alimento ni saciedad gratuitamente.
- La dieta depende del perfil de especie y, si procede, de modificaciones del individuo; no de `Tags` cosméticos.

### Ejemplo `Bark`

- `AllowedLabels: {"can_bark": true}`.
- El estado puede impedir temporalmente ladrar.
- Al ejecutar, genera efecto de sonido/evento observable si el mundo lo permite.
- No asignar `can_bark` a todos los animales automáticamente.

## 8. Datos para que el LLM comprenda las acciones

Sí conviene preparar metadatos legibles por modelos de lenguaje: `Type`, `Name`, `Description`, parámetros tipados, restricciones y consecuencias descritas **sin prometer resultado**.

Ejemplo de catálogo expuesto al agente:

```json
{
  "type": "eat",
  "description": "Intentar ingerir comida disponible y compatible con tu organismo",
  "parameters": [{"name":"target_id","type":"uuid","required":true}],
  "notes": "La acción requiere alimento; su resultado se valida durante la ejecución"
}
```

**Importante**: al LLM se le deben ofrecer únicamente acciones relevantes (y contexto que su criatura conoce). Su JSON es una **propuesta**, no autorización ni ejecución. No delegar las mutaciones de estado en el LLM.

## 9. Concurrencia, tiempo y rendimiento

- Se mantiene la idea del proyecto: una goroutine por criatura activa, pero la definición y validación de acciones son servicios comunes.
- Un `tick` usa `dt` de tiempo **simulado**, no presupone 1 segundo exacto por iteración de tiempo real.
- Las actualizaciones de la misma criatura no deben ejecutarse en paralelo.
- Acciones sobre recursos compartidos (dos perros comiendo la misma ración) deben validarse/aplicarse con coordinación transaccional; los sentidos no evitan condiciones de carrera.
- Evitar recalcular el catálogo completo para cada microcambio si es innecesario. Si se cachea, invalidarlo cuando cambien las capacidades efectivas.
- El modelo de especie puede compartirse por referencia: no duplicarlo en los miles de instancias.

## 10. Alcance de implementación ahora

**Implementar, tras inspeccionar el código real:**

1. Definición mínima de especie y de genoma por individuo, sin herencia genética real.
2. Método/servicio para generar rasgos dentro de rangos reproducibles.
3. Mapas `AllowedLabels` y `DeniedLabels` en `ActionDefinition`, sin destruir los campos existentes.
4. Resolver compatibilidad mediante etiquetas con semántica OR y denegación prioritaria.
5. Pruebas unitarias de generación, rangos, compatibilidad y ausencia de capacidades.
6. Integrar `GetActions` con el catálogo ya implementado, sin generar acciones que el proyecto todavía no soporte.

**No implementar ahora:**

- Acciones dinámicas o su cola de aprobación.
- Herencia genética real, mutaciones, recombinación o reproducción.
- Sistema completo de nutrición, digestión, enfermedades o temperamento.
- Nuevos eventos externos, física, navegación ni motor de percepción.
- Llamadas a LLM o selección mediante Utility AI.
- Reescritura del ciclo de vida ya descrito en `ACTION_STATE_LIFECYCLE.md`.

## 11. Pruebas de aceptación imprescindibles

- Perro con `can_bark` obtiene `Bark`; humano sin ella, no.
- Acción genérica sin permitidas ni prohibidas aparece para ambos, **pero sigue exigiendo validación para iniciarse**.
- Etiqueta prohibida prevalece sobre permitida.
- Etiquetas con valor `false` no cuentan.
- Dos perros generados de la misma especie pueden tener distintas tasas de hambre, siempre dentro de rango.
- Con la misma semilla y orden de generación se obtiene el mismo resultado.
- Una criatura sin metabolismo no provoca pánico al inspeccionar sus acciones.
- Cambiar el hambre actual no cambia el genoma ni el perfil de especie.
- `ActionDefinition` nunca modifica el estado al registrarse.
- `Sleep` o `Eat`, si ya se implementan, no aplican sus efectos completos solo al terminar; usan `Update(dt)`.
- `go test ./...` y, cuando sea viable, `go test -race ./...` funcionan.

## 12. Instrucción breve para Codex

> Revisa el repositorio `simulator` y aplica esta especificación mediante cambios incrementales. Preserva los paquetes `adapters`, `models` y `services`, el `ActionDefinition` real y el resto de modelos existentes. En particular, distingue `Tags` descriptivas de `AllowedLabels`/`DeniedLabels`; distingue compatibilidad de ejecución; y no mezcles definición, instancia activa y efectos de acción. Implementa únicamente el alcance de la sección 10. Muestra qué archivos modificaste y qué pruebas ejecutaste. Si una decisión exige cambiar contratos ya existentes, documenta el conflicto antes de hacerlo.
