import assert from 'node:assert/strict';
import { OpenRideClient, OpenRideError } from '../index.js';

const baseURL = process.env.OPENRIDE_E2E_BASE_URL || 'http://127.0.0.1:18080';

async function jsonRequest(path, { method = 'GET', body, headers = {} } = {}) {
  const response = await fetch(`${baseURL}${path}`, {
    method,
    headers: {
      accept: 'application/json',
      ...(body === undefined ? {} : { 'content-type': 'application/json' }),
      ...headers,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await response.text();
  const payload = text ? JSON.parse(text) : undefined;
  if (!response.ok) {
    throw new Error(`${method} ${path} -> ${response.status}: ${text}`);
  }
  return payload?.data ?? payload;
}

async function login(phone, role) {
  const challenge = await jsonRequest('/v1/auth/otp/request', {
    method: 'POST',
    body: { phone, role },
  });
  assert.ok(challenge.challenge_id, `missing challenge id for ${role}`);
  assert.match(challenge.debug_code || '', /^\d{6}$/, `missing development OTP for ${role}`);

  const verified = await jsonRequest('/v1/auth/otp/verify', {
    method: 'POST',
    body: {
      challenge_id: challenge.challenge_id,
      role,
      code: challenge.debug_code,
    },
  });
  assert.ok(verified.actor?.id, `missing actor for ${role}`);
  assert.ok(verified.tokens?.access_token, `missing access token for ${role}`);
  return verified;
}

async function expectOpenRideError(fn, status, code) {
  try {
    await fn();
  } catch (error) {
    assert.ok(error instanceof OpenRideError, `expected OpenRideError, got ${error}`);
    assert.equal(error.status, status);
    assert.equal(error.code, code);
    return error;
  }
  assert.fail(`expected ${status} ${code}`);
}

const riderAuth = await login('+84911111111', 'rider');
const driverAuth = await login('+84922222222', 'driver');

const rider = new OpenRideClient({ baseURL, token: riderAuth.tokens.access_token });
const driver = new OpenRideClient({ baseURL, token: driverAuth.tokens.access_token });

const services = await rider.listServices();
assert.ok(services.some((service) => service.id === 'passenger.car'), 'passenger.car missing from public catalog');

const requestInput = {
  service_type: 'passenger.car',
  pickup: { lat: 19.8067, lng: 105.7852 },
  destination: { lat: 19.7724, lng: 105.7762 },
  constraints: { max_fare_minor: 70000, max_pickup_eta_s: 900 },
};
const requestKey = 'e2e-create-request-1';
const created = await rider.createRequest(requestInput, { idempotencyKey: requestKey });
assert.ok(created.id?.startsWith('req_'), `unexpected request id ${created.id}`);
assert.equal(created.rider_id, riderAuth.actor.id, 'Edge did not reconstruct rider identity');
assert.equal(created.status, 'open');

const replayed = await rider.createRequest(requestInput, { idempotencyKey: requestKey });
assert.equal(replayed.id, created.id, 'same-key/same-payload did not replay the original resource');

await expectOpenRideError(
  () => rider.createRequest({
    ...requestInput,
    destination: { lat: 19.7000, lng: 105.7000 },
  }, { idempotencyKey: requestKey }),
  409,
  'IDEMPOTENCY_CONFLICT',
);

const tariff = await driver.createDriverTariff({
  service_type: 'passenger.car',
  quote_mode: 'manual',
  currency: 'VND',
  minimum_fare_minor: 20000,
  per_km_minor: 5000,
}, { idempotencyKey: 'e2e-driver-tariff-1' });
assert.equal(tariff.driver_id, driverAuth.actor.id, 'Edge did not reconstruct driver identity');

const quote = await driver.submitQuote(created.id, {
  fare_total_minor: 52000,
  currency: 'VND',
}, { idempotencyKey: 'e2e-quote-1' });
assert.ok(quote.quote_id?.startsWith('quote_'), `unexpected quote id ${quote.quote_id}`);
assert.equal(quote.fare_total_minor, 52000);

const offers = await rider.listOffers(created.id);
assert.equal(offers.length, 1);
assert.equal(offers[0].quote_id, quote.quote_id);
assert.equal(offers[0].fare_total_minor, 52000);
assert.equal(offers[0].rank, 1);
assert.equal(offers[0].recommended, true);
assert.ok(Array.isArray(offers[0].reasons) && offers[0].reasons.length > 0, 'ranker reasons missing');

const accepted = await rider.acceptQuote(quote.quote_id, { idempotencyKey: 'e2e-accept-1' });
assert.ok(accepted.agreement?.id?.startsWith('agr_'), 'agreement missing');
assert.equal(accepted.agreement.quote_id, quote.quote_id);
assert.equal(accepted.agreement.rider_id, riderAuth.actor.id);
assert.equal(accepted.agreement.driver_id, driverAuth.actor.id);
assert.equal(accepted.agreement.fare_total_minor, 52000);

const acceptedReplay = await rider.acceptQuote(quote.quote_id, { idempotencyKey: 'e2e-accept-1' });
assert.equal(acceptedReplay.agreement.id, accepted.agreement.id, 'accept replay returned a different Agreement');

assert.throws(
  () => driver.submitQuote(created.id, {
    fare_total_minor: Number.MAX_SAFE_INTEGER + 1,
    currency: 'VND',
  }, { idempotencyKey: 'e2e-unsafe-sdk-money' }),
  /non-negative safe integer/,
);

const rawUnsafe = await fetch(`${baseURL}/v2/requests/${encodeURIComponent(created.id)}/quotes`, {
  method: 'POST',
  headers: {
    authorization: `Bearer ${driverAuth.tokens.access_token}`,
    'content-type': 'application/json',
    accept: 'application/json',
    'idempotency-key': 'e2e-unsafe-raw-money',
    'x-openride-actor-id': 'spoofed-driver',
  },
  body: `{"fare_total_minor":9007199254740992,"currency":"VND"}`,
});
assert.equal(rawUnsafe.status, 422);
const rawUnsafePayload = await rawUnsafe.json();
assert.equal(rawUnsafePayload.error?.code, 'MONEY_MINOR_INVALID');

console.log(JSON.stringify({
  status: 'ok',
  rider_id: riderAuth.actor.id,
  driver_id: driverAuth.actor.id,
  request_id: created.id,
  quote_id: quote.quote_id,
  agreement_id: accepted.agreement.id,
}));
