# @openride/sdk

Zero-dependency JavaScript/TypeScript client for OpenRide Marketplace V2.

```js
import { OpenRideClient } from '@openride/sdk';

const openride = new OpenRideClient({
  baseURL: 'https://api.example.com',
  token: session.accessToken,
});

const request = await openride.createRequest({
  service_type: 'passenger.car',
  pickup: { lat: 19.8067, lng: 105.7852 },
  destination: { lat: 19.7724, lng: 105.7762 },
}, { idempotencyKey: crypto.randomUUID() });

const offers = await openride.listOffers(request.id);
const agreement = await openride.acceptQuote(offers[0].quote_id, {
  idempotencyKey: crypto.randomUUID(),
});
```

The SDK intentionally contains no UI framework, state manager or transport dependency beyond the Web Fetch API. You can inject a custom `fetch` implementation for Node, mobile shells, tests or edge runtimes.

## Supported V2 flows

- service catalog;
- route preview;
- create/get/cancel mobility requests;
- rider offer list and quote acceptance;
- driver tariffs;
- eligible driver requests;
- submit/withdraw driver quotes.

`@openride/sdk` follows `docs/API_CONTRACT_V2.md`. It is pre-1.0 and should version breaking contract changes explicitly.
