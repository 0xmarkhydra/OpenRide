# OpenRide Public Money Contract

OpenRide V2 represents commercial amounts as **integer minor units** plus a three-letter currency code.

Example:

```json
{
  "fare_total_minor": 52000,
  "currency": "VND"
}
```

## Exact public range

Every public JSON field whose name ends in `_minor` must be an integer in this inclusive range:

```text
0 .. 9,007,199,254,740,991
```

The upper bound is JavaScript `Number.MAX_SAFE_INTEGER` (`2^53 - 1`). OpenRide deliberately uses this narrower range instead of Go/PostgreSQL's full signed `int64` range so JavaScript/TypeScript clients can represent every public commercial amount exactly.

This rule also applies to nested request fields such as:

```json
{
  "constraints": {
    "max_fare_minor": 70000
  }
}
```

Decimal, negative, or larger `_minor` values are invalid public input.

## Enforcement layers

- Marketplace HTTP parses request JSON with `json.Number` before domain decoding and rejects invalid/unsafe `*_minor` values with `422 MONEY_MINOR_INVALID`.
- Marketplace PostgreSQL constrains persisted tariff, quote and Agreement money columns to the same maximum.
- The JavaScript/TypeScript SDK validates tariff and quote money input with `Number.isSafeInteger` before network I/O.
- Go Core continues to use integer minor units internally; public serialization must still obey this contract.

## Why numbers instead of decimal strings

Pre-1.0 V2 keeps JSON numbers for SDK simplicity and compatibility with the existing flat contract. If OpenRide ever needs amounts outside the JS-safe range, that requires a versioned public contract change (for example decimal strings), not an implicit widening of this range.
