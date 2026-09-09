export class OpenRideError extends Error {
  constructor(message, { status = 0, code = '', details = undefined } = {}) {
    super(message);
    this.name = 'OpenRideError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

function joinURL(baseURL, path) {
  return `${String(baseURL).replace(/\/$/, '')}${path.startsWith('/') ? path : `/${path}`}`;
}

export class OpenRideClient {
  constructor({ baseURL, token, fetch: fetchImpl } = {}) {
    if (!baseURL) throw new TypeError('baseURL is required');
    this.baseURL = baseURL;
    this.token = token || '';
    this.fetch = fetchImpl || globalThis.fetch;
    if (typeof this.fetch !== 'function') throw new TypeError('A fetch implementation is required');
  }

  withToken(token) {
    return new OpenRideClient({ baseURL: this.baseURL, token, fetch: this.fetch });
  }

  async request(path, { method = 'GET', body, idempotencyKey, headers = {} } = {}) {
    const requestHeaders = new Headers(headers);
    requestHeaders.set('accept', 'application/json');
    if (body !== undefined) requestHeaders.set('content-type', 'application/json');
    if (this.token) requestHeaders.set('authorization', `Bearer ${this.token}`);
    if (idempotencyKey) requestHeaders.set('idempotency-key', idempotencyKey);

    const response = await this.fetch(joinURL(this.baseURL, path), {
      method,
      headers: requestHeaders,
      body: body === undefined ? undefined : JSON.stringify(body),
    });

    const text = await response.text();
    let payload;
    if (text) {
      try { payload = JSON.parse(text); }
      catch { throw new OpenRideError('OpenRide returned invalid JSON', { status: response.status, details: text }); }
    }

    if (!response.ok) {
      const apiError = payload?.error || payload;
      throw new OpenRideError(apiError?.message || `OpenRide request failed with HTTP ${response.status}`, {
        status: response.status,
        code: apiError?.code || '',
        details: apiError?.details,
      });
    }
    return payload?.data ?? payload;
  }

  listServices() {
    return this.request('/v2/services');
  }

  previewRoute(input) {
    return this.request('/v2/routes/preview', { method: 'POST', body: input });
  }

  createRequest(input, { idempotencyKey } = {}) {
    return this.request('/v2/requests', { method: 'POST', body: input, idempotencyKey });
  }

  getRequest(requestID) {
    return this.request(`/v2/requests/${encodeURIComponent(requestID)}`);
  }

  cancelRequest(requestID, { idempotencyKey, reason } = {}) {
    return this.request(`/v2/requests/${encodeURIComponent(requestID)}/cancel`, {
      method: 'POST', body: reason ? { reason } : {}, idempotencyKey,
    });
  }

  listOffers(requestID) {
    return this.request(`/v2/requests/${encodeURIComponent(requestID)}/offers`);
  }

  acceptQuote(quoteID, { idempotencyKey } = {}) {
    return this.request(`/v2/quotes/${encodeURIComponent(quoteID)}/accept`, {
      method: 'POST', body: {}, idempotencyKey,
    });
  }

  listDriverTariffs() {
    return this.request('/v2/drivers/me/tariffs');
  }

  createDriverTariff(input, { idempotencyKey } = {}) {
    return this.request('/v2/drivers/me/tariffs', { method: 'POST', body: input, idempotencyKey });
  }

  listEligibleRequests() {
    return this.request('/v2/drivers/me/requests');
  }

  submitQuote(requestID, input, { idempotencyKey } = {}) {
    return this.request(`/v2/requests/${encodeURIComponent(requestID)}/quotes`, {
      method: 'POST', body: input, idempotencyKey,
    });
  }

  withdrawQuote(quoteID, { idempotencyKey } = {}) {
    return this.request(`/v2/quotes/${encodeURIComponent(quoteID)}/withdraw`, {
      method: 'POST', body: {}, idempotencyKey,
    });
  }
}
