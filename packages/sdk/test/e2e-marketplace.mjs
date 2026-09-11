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
assert.equal(replayed.id, created.id, 'create request replay returned a different resource');

await expectOpenRideError(
  () => rider.createRequest({
    ...requestInput,
    destination: { lat: 19.7000, lng: 105.7000 },
  }, { idempotencyKey: requestKey }),
  409,
  'IDEMPOTENCY_CONFLICT',
);

const tariffInput = {
  service_type: 'passenger.car',
  quote_mode: 'manual',
  currency: 'VND',
  minimum_fare_minor: 20000,
  per_km_minor: 5000,
};
const tariffKey = 'e2e-driver-tariff-1';
const tariff = await driver.createDriverTariff(tariffInput, { idempotencyKey: tariffKey });
assert.equal(tariff.driver_id, driverAuth.actor.id, 'Edge did not reconstruct driver identity');
const tariffReplay = await driver.createDriverTariff(tariffInput, { idempotencyKey: tariffKey });
assert.equal(tariffReplay.id, tariff.id, 'tariff replay returned a different resource');

const quoteInput = { fare_total_minor: 52000, currency: 'VND' };
const quoteKey = 'e2e-quote-1';
const quote = await driver.submitQuote(created.id, quoteInput, { idempotencyKey: quoteKey });
assert.ok(quote.quote_id?.startsWith('quote_'), `unexpected quote id ${quote.quote_id}`);
assert.equal(quote.fare_total_minor, 52000);
const quoteReplay = await driver.submitQuote(created.id, quoteInput, { idempotencyKey: quoteKey });
assert.equal(quoteReplay.quote_id, quote.quote_id, 'quote replay returned a different resource');

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

const cancelRequest = await rider.createRequest({
  ...requestInput,
  pickup: { lat: 19.8100, lng: 105.7800 },
}, { idempotencyKey: 'e2e-cancel-request-create' });
const cancelled = await rider.cancelRequest(cancelRequest.id, { idempotencyKey: 'e2e-cancel-1', reason: 'e2e' });
assert.equal(cancelled.status, 'cancelled');
const cancelReplay = await rider.cancelRequest(cancelRequest.id, { idempotencyKey: 'e2e-cancel-1', reason: 'e2e' });
assert.equal(cancelReplay.status, 'cancelled', 'cancel replay did not return original status');

const withdrawRequest = await rider.createRequest({
  ...requestInput,
  pickup: { lat: 19.8200, lng: 105.7900 },
}, { idempotencyKey: 'e2e-withdraw-request-create' });
const withdrawQuote = await driver.submitQuote(withdrawRequest.id, {
  fare_total_minor: 53000,
  currency: 'VND',
}, { idempotencyKey: 'e2e-withdraw-quote-create' });
const withdrawn = await driver.withdrawQuote(withdrawQuote.quote_id, { idempotencyKey: 'e2e-withdraw-1' });
assert.equal(withdrawn.status, 'withdrawn');
const withdrawReplay = await driver.withdrawQuote(withdrawQuote.quote_id, { idempotencyKey: 'e2e-withdraw-1' });
assert.equal(withdrawReplay.status, 'withdrawn', 'withdraw replay did not return original status');

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

const spoofRequestResponse = await fetch(`${baseURL}/v2/requests`, {
  method: 'POST',
  headers: {
    authorization: `Bearer ${riderAuth.tokens.access_token}`,
    'content-type': 'application/json',
    accept: 'application/json',
    'idempotency-key': 'e2e-spoof-check',
    'x-openride-actor-id': 'attacker-rider',
    'x-openride-role': 'admin',
  },
  body: JSON.stringify({
    ...requestInput,
    pickup: { lat: 19.8300, lng: 105.8000 },
  }),
});
assert.equal(spoofRequestResponse.status, 201);
const spoofPayload = await spoofRequestResponse.json();
assert.equal(spoofPayload.data?.rider_id, riderAuth.actor.id, 'client trust header overrode authenticated actor');
assert.notEqual(spoofPayload.data?.rider_id, 'attacker-rider');

console.log(JSON.stringify({
  status: 'ok',
  rider_id: riderAuth.actor.id,
  driver_id: driverAuth.actor.id,
  request_id: created.id,
  quote_id: quote.quote_id,
  agreement_id: accepted.agreement.id,
  generic_idempotency: ['create_request', 'create_tariff', 'submit_quote', 'cancel_request', 'withdraw_quote'],
}));
