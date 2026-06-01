import { env } from '@shopee/shared-config';
import type { ApiError } from './types';

const MAX_RETRIES = 3;
const INITIAL_BACKOFF_MS = 1000;
const REQUEST_TIMEOUT_MS = 30000;

interface RequestConfig extends RequestInit {
  params?: Record<string, string | number | boolean>;
  retryCount?: number;
  skipAuth?: boolean;
}

export class ApiClientError extends Error {
  status: number;
  correlationId?: string;

  constructor(message: string, status: number, correlationId?: string) {
    super(message);
    this.name = 'ApiClientError';
    this.status = status;
    this.correlationId = correlationId;
  }
}

function getCorrelationId(): string {
  return `admin-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}

function getAccessToken(): string | null {
  if (typeof document === 'undefined') return null;
  return document.cookie
    .split('; ')
    .find(row => row.startsWith('access_token='))
    ?.split('=')[1] ?? null;
}

function buildUrl(path: string, params?: Record<string, string | number | boolean>): string {
  const baseUrl = `${env.API_GATEWAY_URL}${path}`;
  if (!params) return baseUrl;
  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      searchParams.set(key, String(value));
    }
  });
  const query = searchParams.toString();
  return query ? `${baseUrl}?${query}` : baseUrl;
}

async function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function fetchWithTimeout(
  url: string,
  config: RequestInit,
  timeoutMs: number
): Promise<Response> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      ...config,
      signal: controller.signal,
    });
    return response;
  } finally {
    clearTimeout(timeoutId);
  }
}

function isRetryableStatus(status: number): boolean {
  return status === 429 || status === 502 || status === 503 || status === 504;
}

export async function apiRequest<T>(
  path: string,
  config: RequestConfig = {}
): Promise<T> {
  const { params, retryCount = 0, skipAuth = false, ...fetchConfig } = config;

  const correlationId = getCorrelationId();
  const url = buildUrl(path, params);

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Correlation-ID': correlationId,
    ...(fetchConfig.headers as Record<string, string>),
  };

  if (!skipAuth) {
    const token = getAccessToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  try {
    const response = await fetchWithTimeout(url, { ...fetchConfig, headers }, REQUEST_TIMEOUT_MS);

    if (!response.ok) {
      if (isRetryableStatus(response.status) && retryCount < MAX_RETRIES) {
        const backoff = INITIAL_BACKOFF_MS * Math.pow(2, retryCount);
        await delay(backoff);
        return apiRequest(path, { ...config, retryCount: retryCount + 1 });
      }

      let errorBody: ApiError;
      try {
        errorBody = await response.json();
      } catch {
        errorBody = {
          error: 'unknown_error',
          message: `HTTP ${response.status}: ${response.statusText}`,
          status_code: response.status,
        };
      }

      const error = new ApiClientError(
        errorBody.message || 'Request failed',
        response.status,
        errorBody.correlation_id || correlationId
      );
      throw error;
    }

    if (response.status === 204) {
      return undefined as T;
    }

    const data = await response.json();
    return data as T;
  } catch (err) {
    if (err instanceof ApiClientError) throw err;

    if (err instanceof DOMException && err.name === 'AbortError') {
      throw new ApiClientError('Request timeout', 408, correlationId);
    }

    if (retryCount < MAX_RETRIES) {
      const backoff = INITIAL_BACKOFF_MS * Math.pow(2, retryCount);
      await delay(backoff);
      return apiRequest(path, { ...config, retryCount: retryCount + 1 });
    }

    throw new ApiClientError(
      err instanceof Error ? err.message : 'Network error',
      0,
      correlationId
    );
  }
}

export const api = {
  get: <T>(path: string, params?: Record<string, string | number | boolean>) =>
    apiRequest<T>(path, { method: 'GET', params }),

  post: <T>(path: string, body?: unknown) =>
    apiRequest<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),

  put: <T>(path: string, body?: unknown) =>
    apiRequest<T>(path, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),

  patch: <T>(path: string, body?: unknown) =>
    apiRequest<T>(path, {
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    }),

  delete: <T>(path: string) => apiRequest<T>(path, { method: 'DELETE' }),
};
