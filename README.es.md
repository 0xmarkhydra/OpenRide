<div align="center">

# OpenRide

### Infraestructura open source para mercados de movilidad

**Los conductores fijan sus condiciones. Los pasajeros eligen. Los algoritmos conectan. Las comunidades pueden autoalojarlo.**

[English](README.md) · [简体中文](README.zh-CN.md) · [हिन्दी](README.hi.md) · **Español**

</div>

---

## ¿Y si el transporte bajo demanda fuera infraestructura en vez de un intermediario cerrado?

OpenRide **no es otro clon de Grab/Uber**. Es una base open source para que comunidades de conductores, cooperativas, operadores locales y startups puedan operar sus propios mercados de movilidad sin reconstruir toda la tecnología desde cero.

```text
El pasajero crea la demanda
        ↓
Los conductores ofrecen sus propias condiciones
        ↓
OpenRide ordena las ofertas de forma transparente
        ↓
El pasajero elige — o Quick Match elige dentro de sus restricciones
        ↓
Agreement congela las condiciones aceptadas
        ↓
Ride ejecuta ese acuerdo
```

La plataforma puede recomendar, pero no puede modificar silenciosamente el precio o las condiciones que ambas partes aceptaron.

> **Estado real de implementación:** existen Core, módulos first-party, contracts, el SDK JS/TS y un proceso Marketplace Service desplegable de forma independiente. El esquema de persistencia de Marketplace ya tiene una base, pero el flujo durable completo request → quote → agreement, el Outbox relay, Location Service y Ride Service todavía no están terminados. `docs/PROJECT_STATUS.md` es la fuente de verdad sobre el estado.

## Regla central del mercado

**Cada conductor posee su propio tariff.** No existe un único precio por kilómetro que todos los conductores de un país deban configurar. La configuración de país/instancia puede definir moneda, restricciones legales o límites de seguridad; cada conductor define independientemente condiciones comerciales como `per_km`.

Para la misma solicitud de 10 km:

```text
Driver A → 5.000 VND/km → 50.000 VND → llegada en 10 min
Driver B → 6.000 VND/km → 60.000 VND → llegada en  3 min
Driver C → 5.500 VND/km → 55.000 VND → llegada en  6 min
```

OpenRide no reduce el mercado a `sort(price ASC)`. El pasajero puede comprender el equilibrio entre tarifa, ETA, calidad, fiabilidad y preferencias. El conductor puede cotizar de forma manual, automática o híbrida dentro de límites que él mismo controla.

### Principios no negociables

1. **Los conductores controlan sus condiciones comerciales, incluido su propio precio por km.**
2. **El pasajero conserva la decisión final.**
3. **La oferta más barata no es automáticamente la mejor.**
4. **Rechazar una solicitud inadecuada no es automáticamente mal comportamiento.**
5. **Las condiciones aceptadas se congelan y no pueden cambiarse silenciosamente.**
6. **Pricing y ranking deben ser explicables.**
7. **OpenRide seguirá siendo autoalojable e independiente del operador.**

## Dominio Marketplace

```text
MobilityRequest
DriverTariff
Quote
Marketplace ranking
Agreement
Outbox / Inbox
```

Invariantes principales:

```text
Request → Quote → Agreement → Ride
el dinero usa unidades monetarias menores enteras
cada driver es dueño de su tariff
las condiciones aceptadas forman un snapshot inmutable
ranking puede puntuar/reordenar pero no reescribir quotes
```

## Dirección de microservicios

OpenRide divide servicios por bounded context y ownership, no para crear carpetas vacías que parezcan una arquitectura de microservicios.

```text
one service → one bounded context
each service owns its data + migrations
no cross-service database queries
sync internal calls → gRPC when an immediate answer is required
async integration → NATS JetStream
state change + event → transactional outbox
consumer side effects → inbox / idempotency
public traffic → Edge Gateway / BFF
```

Marketplace es el primer domain service V2 que se está extrayendo. Ya existen el proceso, schema, health/readiness, service catalog y request validation; el API durable completo de Marketplace sigue en implementación.

## Fiabilidad orientada a eventos

```text
DB transaction
  ├── change owned domain state
  └── append outbox event
COMMIT
      ↓
outbox relay
      ↓
NATS JetStream
      ↓
consumer inbox dedupe
      ↓
consumer-owned state change
```

## Probarlo

```bash
make packages-test
make core-example
make marketplace-test
make marketplace-run
```

Stack de desarrollo de microservicios:

```bash
make micro-config
make micro-up
make micro-logs
make micro-down
```

## Documentación

El README y la documentación técnica siguen la misma política de cuatro idiomas: **English, 简体中文, हिन्दी y Español**. El índice de idiomas y el estado de las traducciones están en `docs/README.md`. Si una traducción queda temporalmente detrás de un cambio de código, el documento técnico en inglés es la fuente canónica.

Documentos principales: `PROJECT_STATUS`, `PRODUCT_VISION`, `OPENRIDE_MANIFESTO`, `MICROSERVICES_ARCHITECTURE`, `PACKAGE_ARCHITECTURE`, `API_CONTRACT_V2`.

## Estado del proyecto

OpenRide está en **pre-1.0** y todavía no afirma estar listo para transportar pasajeros reales en producción. Un operador real también debe validar regulación local, seguros, KYC, pagos, respuesta a incidentes, prevención de fraude, privacidad y seguridad.

## Licencia

OpenRide se distribuye bajo **GNU AGPL-3.0-or-later**. Consulta `LICENSE`.

---

<div align="center">

### Construyamos infraestructura para comunidades de movilidad, no otra plataforma cerrada de la que dependan.

⭐ Star · 🧩 Crea un módulo · 🛠️ Opera tu propia instancia · 🤝 Contribuye

</div>
