import test from 'node:test';
import assert from 'node:assert/strict';
import { OpenRideClient, OpenRideError } from '../index.js';

function response(status, payload) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'content-type': 'application/json' },
  });
}

test('createRequest sends bearer token and idempotency key', async () => {
  let captured;
  const client = new OpenRideClient({
    baseURL: 'https://example.test/',
    token: 'token-123',
    fetch: async (url, init) => {
      captured = { url, init };
      return response(200, { data: { id: 'req_1', status: 'open', service_type: 'passenger.car', version: 1 } });
    },
  });

  const result = await client.createRequest({
    service_type: 'passenger.car',
    pickup: { lat: 19.8, lng: 105.7 },
    destination: { lat: 19.7, lng: 105.8 },
  }, { idempotencyKey: 'idem-1' });

  assert.equal(result.id, 'req_1');
  assert.equal(captured.url, 'https://example.test/v2/requests');
  assert.equal(captured.init.headers.get('authorization'), 'Bearer token-123');
  assert.equal(captured.init.headers.get('idempotency-key'), 'idem-1');
});

test('API errors become OpenRideError', async () => {
  const client = new OpenRideClient({
    baseURL: 'https://example.test',
    fetch: async () => response(409, { error: { code: 'QUOTE_EXPIRED', message: 'Quote expired' } }),
  });

  await assert.rejects(
    () => client.acceptQuote('quote_1'),
    (error) => error instanceof OpenRideError && error.status === 409 && error.code === 'QUOTE_EXPIRED',
  );
});

test('submitQuote accepts Number.MAX_SAFE_INTEGER exactly', async () => {
  let body;
  const client = new OpenRideClient({
    baseURL: 'https://example.test',
    fetch: async (_url, init) => {
      body = JSON.parse(init.body);
      return response(201, { data: { quote_id: 'q_1', fare_total_minor: Number.MAX_SAFE_INTEGER, currency: 'VND' } });
    },
  });

  await client.submitQuote('req_1', {
    fare_total_minor: Number.MAX_SAFE_INTEGER,
    currency: 'VND',
  }, { idempotencyKey: 'quote-safe-max' });

  assert.equal(body.fare_total_minor, Number.MAX_SAFE_INTEGER);
});

test('submitQuote rejects money above JavaScript safe integer range before network I/O', async () => {
  let calls = 0;
  const client = new OpenRideClient({
    baseURL: 'https://example.test',
    fetch: async () => {
      calls += 1;
      return response(201, { data: {} });
    },
  });

  assert.throws(
    () => client.submitQuote('req_1', {
      fare_total_minor: Number.MAX_SAFE_INTEGER + 1,
      currency: 'VND',
    }, { idempotencyKey: 'quote-too-large' }),
    /non-negative safe integer/,
  );
  assert.equal(calls, 0);
});

test('createDriverTariff rejects unsafe minor-unit fields before network I/O', async () => {
  let calls = 0;
  const client = new OpenRideClient({
    baseURL: 'https://example.test',
    fetch: async () => {
      calls += 1;
      return response(201, { data: {} });
    },
  });

  assert.throws(
    () => client.createDriverTariff({
      service_type: 'passenger.car',
      quote_mode: 'manual',
      currency: 'VND',
      per_km_minor: Number.MAX_SAFE_INTEGER + 1,
    }, { idempotencyKey: 'tariff-too-large' }),
    /non-negative safe integer/,
  );
  assert.equal(calls, 0);
});
