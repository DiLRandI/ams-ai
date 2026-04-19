import type {
  Asset,
  AssetDocument,
  Category,
  Dashboard,
  FuelLog,
  LoginResponse,
  Reminder,
  ServiceRecord,
  User,
  VehicleEmissionRecord,
  VehicleInsuranceRecord,
  VehicleLicenseRecord,
  VehicleProfile
} from './types';

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '';
const TOKEN_KEY = 'ams_token';

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null) {
  if (token) {
    localStorage.setItem(TOKEN_KEY, token);
  } else {
    localStorage.removeItem(TOKEN_KEY);
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  const token = getToken();
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  if (options.body && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  const response = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (!response.ok) {
    let message = response.statusText;
    let code = 'request_failed';
    try {
      const payload = await response.json();
      message = payload.error?.message ?? message;
      code = payload.error?.code ?? code;
    } catch {
      // Keep status text when response is not JSON.
    }
    throw new ApiError(response.status, code, message);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return response.json() as Promise<T>;
}

export const api = {
  login(email: string, password: string) {
    return request<LoginResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });
  },
  me() {
    return request<User>('/api/auth/me');
  },
  users() {
    return request<User[]>('/api/users');
  },
  categories() {
    return request<Category[]>('/api/categories');
  },
  createCategory(payload: Pick<Category, 'name' | 'description'>) {
    return request<Category>('/api/categories', { method: 'POST', body: JSON.stringify(payload) });
  },
  updateCategory(id: number, payload: Pick<Category, 'name' | 'description'>) {
    return request<Category>(`/api/categories/${id}`, { method: 'PUT', body: JSON.stringify(payload) });
  },
  assets(params: URLSearchParams) {
    const qs = params.toString();
    return request<Asset[]>(`/api/assets${qs ? `?${qs}` : ''}`);
  },
  asset(id: number) {
    return request<Asset>(`/api/assets/${id}`);
  },
  createAsset(payload: Partial<Asset>) {
    return request<Asset>('/api/assets', { method: 'POST', body: JSON.stringify(payload) });
  },
  updateAsset(id: number, payload: Partial<Asset>) {
    return request<Asset>(`/api/assets/${id}`, { method: 'PUT', body: JSON.stringify(payload) });
  },
  archiveAsset(id: number) {
    return request<void>(`/api/assets/${id}/archive`, { method: 'POST' });
  },
  restoreAsset(id: number) {
    return request<void>(`/api/assets/${id}/restore`, { method: 'POST' });
  },
  documents(assetId: number) {
    return request<AssetDocument[]>(`/api/assets/${assetId}/documents`);
  },
  uploadDocument(assetId: number, form: FormData) {
    return request<AssetDocument>(`/api/assets/${assetId}/documents`, { method: 'POST', body: form });
  },
  deleteDocument(id: number) {
    return request<void>(`/api/documents/${id}`, { method: 'DELETE' });
  },
  downloadURL(id: number) {
    return `${API_BASE}/api/documents/${id}/download`;
  },
  vehicleProfile(assetId: number) {
    return request<VehicleProfile>(`/api/assets/${assetId}/vehicle`);
  },
  saveVehicleProfile(assetId: number, payload: Partial<VehicleProfile>) {
    return request<VehicleProfile>(`/api/assets/${assetId}/vehicle`, {
      method: 'PUT',
      body: JSON.stringify(payload)
    });
  },
  insurance(assetId: number) {
    return request<VehicleInsuranceRecord[]>(`/api/assets/${assetId}/vehicle/insurance`);
  },
  createInsurance(assetId: number, payload: Partial<VehicleInsuranceRecord>) {
    return request<VehicleInsuranceRecord>(`/api/assets/${assetId}/vehicle/insurance`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  },
  licenses(assetId: number) {
    return request<VehicleLicenseRecord[]>(`/api/assets/${assetId}/vehicle/licenses`);
  },
  createLicense(assetId: number, payload: Partial<VehicleLicenseRecord>) {
    return request<VehicleLicenseRecord>(`/api/assets/${assetId}/vehicle/licenses`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  },
  emissions(assetId: number) {
    return request<VehicleEmissionRecord[]>(`/api/assets/${assetId}/vehicle/emissions`);
  },
  createEmission(assetId: number, payload: Partial<VehicleEmissionRecord>) {
    return request<VehicleEmissionRecord>(`/api/assets/${assetId}/vehicle/emissions`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  },
  services(assetId: number) {
    return request<ServiceRecord[]>(`/api/assets/${assetId}/services`);
  },
  createService(assetId: number, payload: Partial<ServiceRecord>) {
    return request<ServiceRecord>(`/api/assets/${assetId}/services`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  },
  fuelLogs(assetId: number) {
    return request<FuelLog[]>(`/api/assets/${assetId}/fuel-logs`);
  },
  createFuelLog(assetId: number, payload: Partial<FuelLog>) {
    return request<FuelLog>(`/api/assets/${assetId}/fuel-logs`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  },
  dashboard() {
    return request<Dashboard>('/api/dashboard');
  },
  reminders() {
    return request<Reminder[]>('/api/reminders');
  }
};

export function authDownloadHeaders(): HeadersInit {
  const token = getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}
