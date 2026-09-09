# OpenRide दस्तावेज़

[English](../../README.md) · [简体中文](../zh-CN/README.md) · **हिन्दी** · [Español](../es/README.md)

OpenRide के public docs चार भाषाओं में रखे जाते हैं: English, 简体中文, हिन्दी और Español। English technical docs canonical source हैं। API paths, field names, event names, code और business invariants translation में नहीं बदलने चाहिए।

## सबसे महत्वपूर्ण Marketplace invariant

हर driver अपना tariff own करता है। `per_km` **driver + service tariff** के स्तर पर configure होता है; किसी देश के सभी drivers के लिए एक shared rate नहीं है। Country/instance currency, regulation, safety constraints और operator limits तय कर सकता है, लेकिन driver-owned commercial pricing को replace नहीं करता।

## संरचना

English source `docs/*.md` में है और Hindi mirror समान filename के साथ `docs/i18n/hi/` में रहता है। Maintained public docs को अंततः चारों भाषाओं में पूरा होना चाहिए। यदि translation पीछे हो तो उसे stale mark करना अनिवार्य है।

Priority docs: `PROJECT_STATUS.md`, `PRODUCT_VISION.md`, `OPENRIDE_MANIFESTO.md`, `MICROSERVICES_ARCHITECTURE.md`, `PACKAGE_ARCHITECTURE.md`, `DOMAIN_MODEL.md`, `DATA_MODEL.md`, `API_CONTRACT_V2.md`, `OPENRIDE_MIGRATION_PLAN_V2.md`, `ADR_OPENRIDE_V2.md`।
