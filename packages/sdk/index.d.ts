export type ServiceType = string;
export type Point = { lat: number; lng: number };

export type MoneyInput = {
  currency: string;
  minor: number;
};

export type ServiceManifest = {
  id: ServiceType;
  version: string;
  display_name: string;
  description?: string;
  category: string;
  capabilities?: string[];
  contracts?: Record<string, string>;
};

export type MobilityRequestInput = {
  service_type: ServiceType;
  pickup: Point;
  destination?: Point;
  attributes?: Record<string, unknown>;
  preferences?: Record<string, unknown>;
  constraints?: Record<string, unknown>;
};

export type MobilityRequest = MobilityRequestInput & {
  id: string;
  status: string;
  version: number;
  expires_at?: string;
  route?: { distance_m: number; duration_s: number };
};

export type Offer = {
  quote_id: string;
  fare_total_minor: number;
  currency: string;
  pickup_eta_s: number;
  pickup_distance_m: number;
  expires_at: string;
  rank?: number;
  reasons?: string[];
  recommended?: boolean;
  driver?: Record<string, unknown>;
  vehicle?: Record<string, unknown>;
};

export type DriverTariffInput = {
  service_type: ServiceType;
  quote_mode: 'manual' | 'auto' | 'hybrid';
  currency: string;
  base_fare_minor?: number;
  minimum_fare_minor?: number;
  per_km_minor?: number;
  per_minute_minor?: number;
  pickup_fee_minor?: number;
  auto_quote_min_minor?: number;
  auto_quote_max_minor?: number;
};

export type QuoteInput = {
  fare_total_minor: number;
  currency: string;
};

export type RequestOptions = {
  method?: string;
  body?: unknown;
  idempotencyKey?: string;
  headers?: HeadersInit;
};

export declare class OpenRideError extends Error {
  status: number;
  code: string;
  details?: unknown;
  constructor(message: string, options?: { status?: number; code?: string; details?: unknown });
}

export declare class OpenRideClient {
  readonly baseURL: string;
  readonly token: string;
  readonly fetch: typeof fetch;
  constructor(options: { baseURL: string; token?: string; fetch?: typeof fetch });
  withToken(token: string): OpenRideClient;
  request<T = unknown>(path: string, options?: RequestOptions): Promise<T>;
  listServices(): Promise<ServiceManifest[]>;
  previewRoute(input: { pickup: Point; destination: Point }): Promise<unknown>;
  createRequest(input: MobilityRequestInput, options?: { idempotencyKey?: string }): Promise<MobilityRequest>;
  getRequest(requestID: string): Promise<MobilityRequest>;
  cancelRequest(requestID: string, options?: { idempotencyKey?: string; reason?: string }): Promise<MobilityRequest>;
  listOffers(requestID: string): Promise<Offer[]>;
  acceptQuote(quoteID: string, options?: { idempotencyKey?: string }): Promise<unknown>;
  listDriverTariffs(): Promise<unknown[]>;
  createDriverTariff(input: DriverTariffInput, options?: { idempotencyKey?: string }): Promise<unknown>;
  listEligibleRequests(): Promise<MobilityRequest[]>;
  submitQuote(requestID: string, input: QuoteInput, options?: { idempotencyKey?: string }): Promise<unknown>;
  withdrawQuote(quoteID: string, options?: { idempotencyKey?: string }): Promise<unknown>;
}
