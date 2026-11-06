# Simulator — Action & Creature State Lifecycle

## 1. Objetivo

Definir cómo evolucionan las características de una criatura mientras ejecuta una acción.

El sistema debe permitir que:

- Cada criatura tenga su propio ciclo de ejecución mediante una goroutine.
- Una criatura mantenga una acción activa durante un período de tiempo.
- Sus características fisiológicas y cognitivas cambien progresivamente.
- Las acciones produzcan efectos durante su ejecución, no únicamente al finalizar.
- Las necesidades naturales continúen evolucionando independientemente de la acción.
- Una acción pueda finalizar, fallar o ser interrumpida.
- Al finalizar una acción, el sistema pueda seleccionar la siguiente.

Esta especificación no introduce todavía eventos externos, inteligencia artificial ni acciones dinámicas.

---

## 2. Conceptos fundamentales

### Creature

Entidad que representa a un ser dentro del mundo.

Puede disponer de capacidades fisiológicas como:

- `Metabolism`
- `Energy`
- `Mobility`
- `VitalSigns`
- `Senses`

Y capacidades cognitivas como:

- `Attention`
- `Memory`
- `Emotion`
- `Personality`
- `Reasoning`
- `Goals`

No todas las criaturas disponen necesariamente de todas las capacidades.

### Action

Representa una actividad que una criatura puede realizar.

Ejemplos:

- `Eat`
- `Drink`
- `Sleep`
- `Rest`
- `Run`
- `Move`
- `Bark`

Una acción puede ser instantánea o mantenerse activa durante un período de tiempo.

### ActiveAction

Representa una instancia concreta de una acción que está ejecutándose.

Por ejemplo:

`Sleep` es una definición de acción.

`Toby está durmiendo desde las 22:00` es una acción activa.

Debe conservar información sobre su inicio, tiempo transcurrido, estado y parámetros.

### ActionEffect

Representa una modificación que una acción provoca en un atributo.

Ejemplos:

- `Sleep` reduce el cansancio.
- `Run` aumenta el cansancio.
- `Eat` reduce el hambre.
- `Drink` reduce la sed.

Una acción puede tener múltiples efectos simultáneos.

---

## 3. Evolución del estado

El estado de una criatura depende de dos factores:

1. Evolución natural de sus características.
2. Efectos provocados por la acción que está ejecutando.

La fórmula conceptual es:

**Estado siguiente = Estado actual + Cambio natural + Cambio producido por la acción**

Para un atributo `X`:

```text
X(t + dt) = clamp(
    X(t)
    + NaturalChange(X, dt)
    + ActionChange(X, action, dt)
)
```

Donde:

- `t`: tiempo actual de simulación.
- `dt`: tiempo transcurrido desde la última actualización.
- `NaturalChange`: cambio provocado por el metabolismo o las reglas naturales.
- `ActionChange`: cambio provocado por la acción actual.
- `clamp`: limita el valor dentro de los límites permitidos.

Los valores y coeficientes específicos se definirán posteriormente.

El modelo debe evitar que un atributo supere sus límites sin una regla que lo permita.

---

## 4. Ejemplo: un perro durmiendo

Estado inicial de Toby:

```text
Hunger: 20
Thirst: 15
Fatigue: 80

CurrentAction: Sleep
```

Durante el sueño:

- El hambre puede aumentar debido al metabolismo.
- La sed puede aumentar.
- El cansancio disminuye gracias a la acción `Sleep`.

Ejemplo ilustrativo:

| Tiempo | Hunger | Thirst | Fatigue |
| ------ | -----: | -----: | ------: |
| 00:00  |     20 |     15 |      80 |
| 00:01  |   20.1 |   15.1 |      79 |
| 00:02  |   20.2 |   15.2 |      78 |
| 00:03  |   20.3 |   15.3 |      77 |

Estos valores no representan reglas definitivas.

Lo importante es que **los cambios ocurren durante la acción**, no únicamente cuando termina.

---

## 5. Ejemplo: un perro comiendo

Una acción `Eat` debe producir al menos dos efectos relacionados:

1. Reducir progresivamente el hambre del perro.
2. Consumir progresivamente comida del mundo o del inventario.

El sistema debe garantizar que la reducción del hambre dependa de la cantidad real de comida ingerida.

Ejemplo:

```text
Toby inicia Eat.

Segundo 1:
    Consume comida.
    Disminuye su hambre.

Segundo 2:
    Consume comida.
    Disminuye su hambre.

Segundo 3:
    No queda comida.
    Eat termina.
```

La acción también puede producir efectos secundarios.

Por ejemplo, si el perro come demasiado, podría experimentar malestar.

Estos efectos adicionales se diseñarán posteriormente.

No deben implementarse todavía reglas de digestión, enfermedades o empacho sin definir previamente sus modelos.

---

## 6. Ciclo de ejecución de una criatura

Cada criatura activa dispone de una goroutine.

Su ciclo conceptual es:

```text
START
  |
  v
¿Existe una acción activa?
  |
  +-- NO --> Seleccionar siguiente acción
  |                |
  |                v
  |          Iniciar acción
  |                |
  +----------------+
  |
  v
Esperar siguiente actualización
  |
  v
Calcular tiempo transcurrido (dt)
  |
  v
Actualizar evolución fisiológica natural
  |
  v
Aplicar efectos de la acción activa
  |
  v
Comprobar límites y condiciones
  |
  v
¿La acción debe finalizar?
  |
  +-- NO --> Continuar acción
  |
  +-- SÍ --> Finalizar acción
                   |
                   v
             Seleccionar otra
```

Una criatura no debe ejecutar un proceso intensivo de decisión en cada actualización.

El ciclo de actualización y el proceso de toma de decisiones son responsabilidades distintas.

---

## 7. Frecuencia de actualización

Para el primer prototipo, se propone actualizar cada criatura aproximadamente una vez por segundo.

Sin embargo, las modificaciones deben calcularse utilizando `dt`, no suponiendo que cada iteración dura exactamente un segundo.

Ejemplo:

```go
deltaTime := 1.25 // Segundos transcurridos

hungerIncrease := hungerRate * deltaTime
```

Esto permite mantener una evolución proporcional al tiempo incluso cuando una actualización se retrasa.

**Nota:** el tiempo real y el tiempo simulado no tienen por qué ser equivalentes. En el futuro, el mundo podrá tener un reloj virtual centralizado. La lógica de actualización debe recibir tiempo simulado de forma explícita.

---

## 8. Modelo conceptual de los efectos

Se propone definir un modelo `ActionEffect` reutilizable.

Ejemplo ilustrativo en Go:

```go
type ActionEffect struct {
    Attribute string
    Rate      float64
}
```

Una acción `Sleep` podría declarar:

```go
[]ActionEffect{
    {
        Attribute: "energy.fatigue",
        Rate: -0.02,
    },
}
```

El valor de `Rate` representa el cambio por unidad de tiempo simulado.

La sintaxis definitiva de los identificadores de atributos queda pendiente.

### Consideración importante

No todas las acciones se pueden representar correctamente mediante una tasa constante.

En el futuro se contemplan:

- Efectos condicionados.
- Efectos no lineales.
- Efectos progresivos.
- Efectos limitados por recursos.
- Efectos de inicio o finalización.
- Efectos secundarios.

Para el MVP se utilizarán efectos lineales donde resulte apropiado.

---

## 9. Ciclo de vida de una acción

Una acción activa podrá tener estos estados:

| Estado        | Significado                       |
| ------------- | --------------------------------- |
| `Pending`     | Seleccionada, todavía no iniciada |
| `Running`     | En ejecución                      |
| `Completed`   | Terminada correctamente           |
| `Interrupted` | Interrumpida antes de terminar    |
| `Failed`      | No pudo ejecutarse correctamente  |

Ejemplo:

```text
Toby empieza a dormir.
    ↓
Sleep: Running
    ↓
Su cansancio disminuye progresivamente.
    ↓
Sleep: Completed
    ↓
Se selecciona otra acción.
```

En caso de interrupción:

```text
Toby empieza a dormir.
    ↓
Sleep: Running
    ↓
Se produce una interrupción.
    ↓
Sleep: Interrupted
    ↓
Se selecciona otra acción.
```

El origen y la propagación de las interrupciones externas se desarrollarán cuando se implemente el sistema de eventos.

---

## 10. Validación de acciones

Una acción no puede aplicarse únicamente porque haya sido seleccionada.

Antes de iniciarla deben comprobarse sus requisitos.

Durante su ejecución deben mantenerse las condiciones necesarias.

Ejemplos:

- `Eat` requiere alimento disponible.
- `Run` requiere movilidad suficiente.
- `Drink` requiere acceso a agua.
- `Sleep` requiere que la criatura pueda dormir.

Si deja de cumplirse una condición durante la ejecución, la acción podrá interrumpirse o fallar.

**Las capacidades ausentes deben tratarse explícitamente.**

Por ejemplo, una criatura sin `Metabolism` no debe provocar un acceso inválido a `Metabolism.Hunger`.

---

## 11. Concurrencia y estado del mundo

Cada criatura puede tener una goroutine independiente.

Sin embargo, dos criaturas pueden interactuar con el mismo recurso.

Ejemplo:

```text
Toby intenta comer una ración.

Un lobo intenta comer esa misma ración.
```

Las goroutines no deben modificar simultáneamente el recurso sin coordinación.

Se establece la siguiente regla arquitectónica:

**Una criatura puede calcular sus necesidades y proponer efectos, pero los cambios sobre recursos compartidos deben validarse y aplicarse de forma coordinada.**

Para los atributos internos de una criatura, se procurará que su propia goroutine sea la responsable de actualizarlos.

Las operaciones compartidas requerirán mecanismos de sincronización o un servicio central que aplique los cambios.

---

## 12. Relación futura con Event

`Action` y `Event` representan conceptos diferentes.

**Action:** actividad intentada o ejecutada por una criatura.

**Event:** acontecimiento producido dentro del mundo.

Una acción puede generar eventos en distintos momentos:

```text
Action: Bark
    |
    v
Event: SoundProduced
```

También puede generar eventos durante su ejecución:

```text
Action: Eat
    |
    +-- FoodConsumed
    |
    +-- FoodConsumed
    |
    +-- ActionCompleted
```

No es obligatorio generar un evento global por cada cambio numérico de un atributo.

La actualización interna del hambre puede ocurrir sin publicarse como evento.

Se definirán posteriormente las reglas de publicación, propagación y percepción.

---

## 13. Integración futura con la toma de decisiones

Cuando termina o se interrumpe una acción, la criatura debe seleccionar qué hacer a continuación.

Inicialmente se propone un algoritmo basado en utilidad (*Utility AI*).

La decisión tendrá en cuenta:

- Capacidades fisiológicas.
- Estado fisiológico actual.
- Personalidad.
- Emociones.
- Memoria.
- Acciones disponibles.
- Requisitos de ejecución.

El sistema de decisiones no debe modificar directamente el estado del mundo.

Su responsabilidad será elegir una acción candidata.

La ejecución será responsabilidad del sistema de acciones.

---

## 14. Organización del proyecto

Respetar la estructura existente de `simulator`:

```text
simulator/
├── adapters/
├── cmd/
├── models/
│   ├── actions/
│   │   ├── action.go
│   │   └── action_effect.go
│   ├── cognitive_functions/
│   ├── creatures/
│   ├── map/
│   └── physiological_capacities/
└── services/
    ├── creature_service.go
    ├── action_service.go
    └── decision_service.go
```

Esta distribución es conceptual. No requiere crear todos los archivos inmediatamente.

- `models/actions`: definiciones e instancias de acciones y efectos.
- `models/creatures`: estructuras de criaturas.
- `models/physiological_capacities`: características fisiológicas.
- `models/cognitive_functions`: características cognitivas.
- `services/creature_service.go`: ciclo de actualización de cada criatura.
- `services/action_service.go`: ejecución y validación de acciones.
- `services/decision_service.go`: selección de la siguiente acción.

---

## 15. Reglas de diseño

1. Los modelos representan datos, no ejecutan bucles ni llaman a LLMs.
2. Las acciones pueden durar varios ciclos de actualización.
3. Los efectos se aplican progresivamente mientras la acción está activa.
4. El metabolismo y los efectos de acciones deben poder combinarse.
5. Las actualizaciones utilizan tiempo transcurrido, no un segundo fijo asumido.
6. Las acciones no deben crear recursos inexistentes.
7. Los cambios compartidos deben estar coordinados.
8. Las criaturas pueden tener capacidades opcionales.
9. Las decisiones y las actualizaciones de estado son procesos diferentes.
10. El sistema debe admitir en el futuro acciones dinámicas y eventos externos sin cambiar estos principios.

---

## 16. Alcance inicial

El primer prototipo implementará únicamente:

- Una criatura `Dog`.
- Un ciclo de actualización.
- Acciones `Sleep`, `Eat` y `Rest`.
- Cambios progresivos en hambre y cansancio.
- Inicio y finalización de acciones.
- Registro de cambios de estado.

Quedan fuera de alcance:

- LLM.
- Eventos externos.
- Propagación de sonidos y visión.
- Acciones dinámicas.
- Digestión detallada.
- Sistema complejo de heridas.
- Persistencia en base de datos.
- Sincronización distribuida.

### Criterio de éxito

El prototipo será válido cuando podamos iniciar una acción de varios minutos simulados, observar cómo cambian los atributos de la criatura durante su ejecución y comprobar que al finalizar puede seleccionar otra acción.

**Principio fundamental: una acción no es simplemente algo que ocurre; es una actividad que transforma progresivamente el estado de la criatura y, cuando corresponda, del mundo.**
