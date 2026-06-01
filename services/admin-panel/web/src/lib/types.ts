import { env } from '@shopee/shared-config';

export const API_BASE_URL = env.API_GATEWAY_URL;

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
  correlation_id?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface ApiError {
  error: string;
  message: string;
  status_code: number;
  correlation_id?: string;
}
