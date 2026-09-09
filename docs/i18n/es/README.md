# Documentación de OpenRide

[English](../../README.md) · [简体中文](../zh-CN/README.md) · [हिन्दी](../hi/README.md) · **Español**

La documentación pública de OpenRide sigue una política de cuatro idiomas: English, 简体中文, हिन्दी y Español. La documentación técnica en inglés es la fuente canónica. Las traducciones deben conservar rutas API, nombres de campos, eventos, código e invariantes de negocio.

## Invariante principal de Marketplace

Cada conductor es dueño de su propio tariff. `per_km` se configura por **driver + service tariff**, no como una única tarifa compartida por todos los conductores de un país. La configuración de país/instancia puede definir moneda, regulación, límites de seguridad y restricciones del operador, pero no sustituye el precio comercial definido por cada conductor.

## Estructura

Los documentos fuente en inglés viven en `docs/*.md`. El espejo en español utiliza el mismo nombre de archivo bajo `docs/i18n/es/`. Todo documento público mantenido debe terminar cubierto en los cuatro idiomas. Si una traducción queda atrasada, debe marcarse explícitamente como stale.

Prioridad: `PROJECT_STATUS.md`, `PRODUCT_VISION.md`, `OPENRIDE_MANIFESTO.md`, `MICROSERVICES_ARCHITECTURE.md`, `PACKAGE_ARCHITECTURE.md`, `DOMAIN_MODEL.md`, `DATA_MODEL.md`, `API_CONTRACT_V2.md`, `OPENRIDE_MIGRATION_PLAN_V2.md`, `ADR_OPENRIDE_V2.md`.
