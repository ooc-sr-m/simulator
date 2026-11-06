Simulator


Es un proyecto experimental para construir un mundo medieval persistente donde personajes no jugadores (NPCs) actúan de forma autónoma mediante modelos de lenguaje (LLMs), mientras un motor de simulación escrito en Go gobierna el estado, el tiempo y las consecuencias.

Estado del proyecto: diseño y prototipo inicial. Las capacidades descritas como previstas no están implementadas todavía.

Objetivo
Crear una simulación donde los NPCs tengan objetivos, recursos, conocimientos y relaciones propios; puedan actuar aunque el jugador no esté interactuando; y produzcan consecuencias verificables en un mundo compartido.

El principio fundamental es:
El LLM propone lo que un personaje quiere hacer. El motor de Go decide si puede hacerlo y qué sucede.

Por ejemplo, un comerciante puede proponer comprar 12 raciones. El motor comprobará que existe un vendedor, que dispone de existencias, que el comprador tiene dinero y que ambos pueden interactuar. Solo entonces registrará la compra y actualizará los inventarios.

Principios de diseño

- Coherencia causal: nada aparece o desaparece sin una acción o regla que lo justifique.
- Estado compartido y consistente: todas las acciones afectan al mismo mundo; los cambios deben validarse de forma segura.
- Conocimiento limitado: cada NPC recibe únicamente información que conoce, observa o aprende.
- Autonomía real: los NPCs persiguen objetivos propios y no esperan instrucciones narrativas del jugador.
- Tiempo simulado: viajar, fabricar, descansar o comerciar requiere tiempo.
- Memoria y consecuencias: los agentes recuerdan acontecimientos pertinentes y sus relaciones evolucionan.
- Trazabilidad: registrar propuestas del LLM, validaciones, decisiones y cambios de estado para poder auditar la simulación.
- Modelos intercambiables: poder utilizar un LLM local mediante Ollama o un proveedor remoto como OpenAI.

## Ejecutar el simulador

Con Go 1.26 o posterior instalado, ejecuta desde la raíz del proyecto:

```powershell
go run ./cmd/simulator
```

El comando inicializa `WorldService`, crea diez perros con genomas aleatorios dentro de los rangos definidos para la especie y muestra sus datos y capacidades en la consola.
