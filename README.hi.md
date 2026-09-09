<div align="center">

# OpenRide

### ओपन-सोर्स मोबिलिटी मार्केटप्लेस इन्फ्रास्ट्रक्चर

**ड्राइवर अपनी शर्तें तय करते हैं। राइडर चुनते हैं। एल्गोरिदम जोड़ते हैं। समुदाय स्वयं होस्ट कर सकते हैं।**

[English](README.md) · [简体中文](README.zh-CN.md) · **हिन्दी** · [Español](README.es.md)

</div>

---

## अगर राइड-हेलिंग किसी बंद प्लेटफ़ॉर्म की जगह सार्वजनिक रूप से उपयोग योग्य इन्फ्रास्ट्रक्चर हो?

OpenRide **Grab/Uber का एक और क्लोन नहीं है**। यह ड्राइवर समुदायों, सहकारी संस्थाओं, स्थानीय ऑपरेटरों और स्टार्टअप्स के लिए एक ओपन-सोर्स आधार है, ताकि वे पूरा तकनीकी स्टैक शून्य से बनाए बिना अपना मोबिलिटी मार्केटप्लेस चला सकें।

```text
राइडर मांग बनाता है
        ↓
ड्राइवर अपनी शर्तों पर ऑफ़र देते हैं
        ↓
OpenRide पारदर्शी तरीके से रैंक करता है
        ↓
राइडर चुनता है — या उसकी सीमाओं के भीतर Quick Match चुनता है
        ↓
Agreement स्वीकार की गई शर्तों का snapshot बनाता है
        ↓
Ride उसी agreement के अनुसार चलती है
```

प्लेटफ़ॉर्म सुझाव दे सकता है, लेकिन दोनों पक्षों द्वारा स्वीकार की गई कीमत या शर्तों को चुपचाप बदल नहीं सकता।

> **वास्तविक implementation स्थिति:** Core, first-party modules, contracts, JS/TS SDK और स्वतंत्र रूप से deploy होने वाला Marketplace Service process मौजूद हैं। Marketplace persistence schema की नींव मौजूद है, लेकिन पूरा durable request → quote → agreement flow, Outbox relay, Location Service और Ride Service अभी पूर्ण नहीं हैं। सही स्थिति `docs/PROJECT_STATUS.md` में है।

## मार्केटप्लेस का मुख्य नियम

**हर ड्राइवर का अपना tariff होता है।** कीमत देश के सभी ड्राइवरों के लिए एक समान per-km rate नहीं है। Country/instance configuration currency, कानूनी नियम या safety limits तय कर सकती है, लेकिन हर ड्राइवर अपनी `per_km` जैसी commercial terms स्वतंत्र रूप से तय करता है।

एक ही 10 km request के लिए:

```text
Driver A → 5,000 VND/km → 50,000 VND → pickup 10 min
Driver B → 6,000 VND/km → 60,000 VND → pickup  3 min
Driver C → 5,500 VND/km → 55,000 VND → pickup  6 min
```

OpenRide मार्केटप्लेस को केवल `sort(price ASC)` तक सीमित नहीं करता। राइडर fare, pickup ETA, quality, reliability और preferences के बीच अंतर समझ सकता है। ड्राइवर manual, automatic या hybrid quoting चुन सकता है और अपनी price limits नियंत्रित करता है।

### गैर-समझौतावादी सिद्धांत

1. **ड्राइवर अपनी commercial terms और अपना per-km price नियंत्रित करते हैं।**
2. **अंतिम चुनाव राइडर का रहता है।**
3. **सबसे सस्ता ऑफ़र अपने-आप सबसे अच्छा नहीं माना जाएगा।**
4. **अनुपयुक्त request को मना करना अपने-आप खराब व्यवहार नहीं है।**
5. **स्वीकार की गई शर्तें snapshot होती हैं और चुपचाप बदली नहीं जा सकतीं।**
6. **Pricing और ranking explainable होने चाहिए।**
7. **OpenRide self-hostable और operator-independent रहेगा।**

## Marketplace domain

```text
MobilityRequest
DriverTariff
Quote
Marketplace ranking
Agreement
Outbox / Inbox
```

मुख्य invariants:

```text
Request → Quote → Agreement → Ride
money integer minor units में रहता है
हर driver अपना tariff own करता है
accepted commercial terms immutable snapshot बनते हैं
ranking score/reorder कर सकता है लेकिन quote rewrite नहीं कर सकता
```

## Microservices दिशा

OpenRide services को bounded context और ownership के आधार पर विभाजित करता है।

```text
one service → one bounded context
each service owns its data + migrations
no cross-service database queries
sync internal calls → gRPC when immediate response is required
async integration → NATS JetStream
state change + event → transactional outbox
consumer side effects → inbox / idempotency
public traffic → Edge Gateway / BFF
```

Marketplace पहला V2 domain service है जिसे अलग process boundary में निकाला जा रहा है। Process, schema, health/readiness, service catalog और request validation मौजूद हैं; पूरा durable Marketplace API अभी implement हो रहा है।

## Reliability flow

```text
DB transaction
  ├── owned domain state change
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

## चलाएँ

```bash
make packages-test
make core-example
make marketplace-test
make marketplace-run
```

Development microservices stack:

```bash
make micro-config
make micro-up
make micro-logs
make micro-down
```

## Documentation

README और technical docs चार भाषाओं की policy का पालन करते हैं: **English, 简体中文, हिन्दी और Español**। Language index और translation status `docs/README.md` में हैं। Code change के बाद translation अस्थायी रूप से पीछे हो तो English technical document canonical source रहेगा।

मुख्य docs: `PROJECT_STATUS`, `PRODUCT_VISION`, `OPENRIDE_MANIFESTO`, `MICROSERVICES_ARCHITECTURE`, `PACKAGE_ARCHITECTURE`, `API_CONTRACT_V2`।

## Project status

OpenRide अभी **pre-1.0** है और वास्तविक यात्रियों के लिए production-ready होने का दावा नहीं करता। किसी वास्तविक operator को स्थानीय transport regulation, insurance, KYC, payment, incident response, fraud prevention, privacy और safety requirements सत्यापित करनी होंगी।

## License

OpenRide **GNU AGPL-3.0-or-later** के अंतर्गत licensed है। `LICENSE` देखें।

---

<div align="center">

### मोबिलिटी समुदायों के लिए इन्फ्रास्ट्रक्चर बनाइए — एक और बंद प्लेटफ़ॉर्म नहीं।

⭐ Star · 🧩 Module बनाइए · 🛠️ अपना operator चलाइए · 🤝 योगदान दीजिए

</div>
