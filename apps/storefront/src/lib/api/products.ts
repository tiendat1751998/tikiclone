const API_BASE = "/api/gateway";

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  const data = await res.json();
  return (data?.data || data) as T;
}

export const productsApi = {
  getFeatured: <T>(limit = 12) => fetchJSON<T[]>(`${API_BASE}/products?sort=featured&limit=${limit}`),
  getFlashSale: <T>(limit = 10) => fetchJSON<T[]>(`${API_BASE}/products?sort=flash_sale&limit=${limit}`),
  getRecommended: <T>(limit = 12) => fetchJSON<T[]>(`${API_BASE}/products?sort=recommended&limit=${limit}`),
  getDeals: <T>(limit = 20) => fetchJSON<T[]>(`${API_BASE}/products?sort=deals&limit=${limit}`),
  getById: <T>(id: string) => fetchJSON<T>(`${API_BASE}/products/${id}`),
};

export const categoriesApi = {
  getTop: <T>(limit = 8) => fetchJSON<T[]>(`${API_BASE}/categories?limit=${limit}`),
  getTree: <T>() => fetchJSON<T[]>(`${API_BASE}/categories/tree`),
};
