# OpenRide Product Vision

## 1. Purpose

OpenRide exists to provide an open, self-hostable foundation for fair mobility marketplaces.

The system should help riders and drivers find each other, exchange transparent commercial offers, agree on terms, execute a ride or mobility job safely, and settle payment without requiring one central platform to dictate the entire market.

## 2. Product thesis

Traditional ride-hailing products commonly collapse four different business concepts into one object called a trip:

1. demand;
2. price;
3. assignment;
4. execution.

OpenRide deliberately separates them:

```text
MobilityRequest -> Quote -> Agreement -> Ride
```

This separation allows drivers to control their pricing policy and allows riders to compare meaningful alternatives.

## 3. Primary actors

### Rider

A rider wants to:
- declare pickup and destination;
- specify service requirements;
- receive transparent offers;
- compare price, pickup ETA, driver quality and reliability;
- select a driver or use quick match;
- track the accepted ride realtime;
- pay and rate safely.

### Driver

A driver wants to:
- complete identity/KYC requirements;
- declare vehicle and capabilities;
- set availability;
- configure pricing/tariff rules;
- receive relevant requests;
- quote manually or automatically within self-approved bounds;
- skip unsuitable requests without arbitrary punishment;
- execute accepted rides;
- understand earnings and settlement.

### Operator

An operator exists to run the marketplace infrastructure, not to become the owner of every commercial decision.

Operator responsibilities include:
- KYC and trust;
- moderation and fraud controls;
- legal/service-area policy;
- payment configuration;
- safety and support;
- dispute resolution;
- system health;
- marketplace health;
- transparent guardrails.

### Community / Instance owner

OpenRide should later support different instances operated by:
- driver communities;
- cooperatives;
- taxi groups;
- local startups;
- regional organizations.

The codebase should therefore avoid assumptions that there can only ever be one global operator.

## 4. Marketplace modes

### 4.1 Marketplace Mode

The rider creates a request. Eligible drivers can produce quotes. The rider sees multiple offers and chooses.

Example:

```text
Request: 10 km

Driver A: 50,000 VND - pickup 10 min - rating 4.9
Driver B: 60,000 VND - pickup 3 min  - rating 4.8
Driver C: 55,000 VND - pickup 6 min  - rating 5.0
```

The UI should explain useful distinctions such as best overall, cheapest and fastest pickup instead of pretending one number is objectively best.

### 4.2 Quick Match

The rider can delegate the choice to OpenRide using constraints such as:
- maximum fare;
- maximum pickup ETA;
- minimum rating;
- service preferences.

Quick Match must select from driver-authorized quotes. It must not silently replace driver pricing with a platform-owned fare.

## 5. Driver pricing model

Pricing should be modeled as `DriverTariff`, not as one global platform rule.

A tariff may include:
- base fare;
- minimum fare;
- per-km rate;
- per-minute rate;
- pickup fee;
- maximum pickup distance;
- night rule;
- holiday rule;
- long-distance rule;
- zone rule;
- manual/auto/hybrid mode;
- minimum and maximum automatic quote bounds.

The system may produce market recommendations, but a recommendation is not ownership of the price.

## 6. Ranking philosophy

Marketplace ranking should be multi-objective.

Possible factors:
- pickup ETA;
- price fit;
- rating;
- completion reliability;
- driver/rider preferences;
- service capability;
- marketplace fairness/exposure.

OpenRide must not rank only by lowest fare because that creates a race to the bottom for drivers.

Every ranking should be explainable enough that a user can understand why an offer was recommended.

## 7. Agreement

When a rider accepts a quote, OpenRide creates an immutable commercial snapshot called `Agreement`.

At minimum it records:
- request;
- rider;
- driver;
- selected vehicle/service capability;
- pickup and destination;
- accepted price;
- pricing breakdown;
- currency;
- accepted terms;
- timestamps.

After this point the agreed fare cannot be silently recomputed because supply/demand changed.

## 8. Ride execution

`Ride` represents execution after agreement.

Typical passenger ride lifecycle:

```text
ASSIGNED
-> DRIVER_EN_ROUTE
-> DRIVER_ARRIVED
-> PASSENGER_ONBOARD
-> IN_PROGRESS
-> COMPLETED
```

Cancellation, incident and dispute paths are separate explicit transitions.

## 9. Future service types

The marketplace core should not assume every request is one passenger ride.

Potential verticals:
- car ride;
- motorbike ride;
- carpool;
- intercity ride;
- delivery;
- designated driver;
- vehicle inspection assistance;
- other local mobility jobs.

Legacy FlashX workflows should eventually become service-specific verticals on top of the same request/quote/agreement foundation.

## 10. Product success criteria

OpenRide is successful when:
- a driver can clearly control their pricing rules;
- a rider can clearly compare offers;
- accepted terms are auditable;
- matching remains fast enough for real-world mobility;
- operators can self-host and configure local policy;
- new mobility verticals can be added without rewriting the core marketplace;
- contributors can understand the product model from public documentation and tests.
