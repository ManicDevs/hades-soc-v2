// ============================================================================
// Hades-V2 API Configuration & Utilities
// ============================================================================

import type { ApiEndpointConfig, ApiError } from './index';

// ============================================================================
// Request Options Interface
// ============================================================================

export interface RequestOptions extends RequestInit {
  timeout?: number;
  retry?: boolean;
  retryAttempts?: number;
  retryDelay?: number;
  signal?: AbortSignal;
}

export interface RequestConfig extends RequestInit {
  baseURL?: string;
  timeout?: number;
  headers?: Record<string, string>;
  params?: Record<string, string | number | boolean>;
}

// ============================================================================
// API Endpoints Configuration
// ============================================================================

export const API_ENDPOINTS = {
  // Auth
  auth: {
    login: '/api/v2/auth/login',
    logout: '/api/v2/auth/logout',
    refresh: '/api/v2/auth/refresh',
    register: '/api/v2/auth/register',
    forgotPassword: '/api/v2/auth/forgot-password',
    resetPassword: '/api/v2/auth/reset-password',
    verifyMfa: '/api/v2/auth/verify-mfa',
    changePassword: '/api/v2/auth/change-password',
  },

  // Dashboard
  dashboard: {
    metrics: '/api/v2/dashboard/metrics',
    activity: '/api/v2/dashboard/activity',
    status: '/api/v2/dashboard/status',
    security: '/api/v2/dashboard/security',
  },

  // Users
  users: {
    list: '/api/v2/users',
    get: (id: string) => `/api/v2/users/${id}`,
    create: '/api/v2/users',
    update: (id: string) => `/api/v2/users/${id}`,
    delete: (id: string) => `/api/v2/users/${id}`,
    roles: '/api/v2/users/roles',
    permissions: '/api/v2/users/permissions',
    profile: '/api/v2/users/profile',
    updateProfile: '/api/v2/users/profile',
  },

  // Threats
  threats: {
    list: '/api/v2/threats',
    get: (id: string) => `/api/v2/threats/${id}`,
    create: '/api/v2/threats',
    update: (id: string) => `/api/v2/threats/${id}`,
    delete: (id: string) => `/api/v2/threats/${id}`,
    block: (id: string) => `/api/v2/threats/${id}/block`,
    unblock: (id: string) => `/api/v2/threats/${id}/unblock`,
    stats: '/api/v2/threats/stats',
    history: (id: string) => `/api/v2/threats/${id}/history`,
  },

  // Security
  security: {
    policies: '/api/v2/security/policies',
    getPolicy: (id: string) => `/api/v2/security/policies/${id}`,
    createPolicy: '/api/v2/security/policies',
    updatePolicy: (id: string) => `/api/v2/security/policies/${id}`,
    deletePolicy: (id: string) => `/api/v2/security/policies/${id}`,
    vulnerabilities: '/api/v2/security/vulnerabilities',
    getVulnerability: (id: string) => `/api/v2/security/vulnerabilities/${id}`,
    score: '/api/v2/security/score',
  },

  // Kubernetes
  kubernetes: {
    clusters: '/api/v2/kubernetes/clusters',
    getCluster: (id: string) => `/api/v2/kubernetes/clusters/${id}`,
    deployments: '/api/v2/kubernetes/deployments',
    getDeployment: (id: string) => `/api/v2/kubernetes/deployments/${id}`,
    services: '/api/v2/kubernetes/services',
    getService: (id: string) => `/api/v2/kubernetes/services/${id}`,
    pods: '/api/v2/kubernetes/pods',
    getPod: (id: string) => `/api/v2/kubernetes/pods/${id}`,
    autoscalers: '/api/v2/kubernetes/autoscalers',
  },

  // SIEM
  siem: {
    events: '/api/v2/siem/events',
    getEvent: (id: string) => `/api/v2/siem/events/${id}`,
    alerts: '/api/v2/siem/alerts',
    getAlert: (id: string) => `/api/v2/siem/alerts/${id}`,
    correlations: '/api/v2/siem/correlations',
    getCorrelation: (id: string) => `/api/v2/siem/correlations/${id}`,
    threatFeeds: '/api/v2/siem/threat-feeds',
    rules: '/api/v2/siem/rules',
  },

  // Incident Response
  incident: {
    incidents: '/api/v2/incident/incidents',
    getIncident: (id: string) => `/api/v2/incident/incidents/${id}`,
    createIncident: '/api/v2/incident/incidents',
    updateIncident: (id: string) => `/api/v2/incident/incidents/${id}`,
    playbooks: '/api/v2/incident/playbooks',
    getPlaybook: (id: string) => `/api/v2/incident/playbooks/${id}`,
    responseActions: '/api/v2/incident/response-actions',
    activeResponses: '/api/v2/incident/active-responses',
  },

  // Analytics
  analytics: {
    overview: '/api/v2/analytics/overview',
    mlInsights: '/api/v2/analytics/ml-insights',
    predictions: '/api/v2/analytics/predictions',
    metrics: '/api/v2/analytics/metrics',
    trends: '/api/v2/analytics/trends',
    reports: '/api/v2/analytics/reports',
    export: '/api/v2/analytics/export',
  },

  // Threat Intelligence
  threatIntelligence: {
    indicators: '/api/v2/threat-intelligence/indicators',
    feeds: '/api/v2/threat-intelligence/feeds',
    actors: '/api/v2/threat-intelligence/actors',
    campaigns: '/api/v2/threat-intelligence/campaigns',
  },

  // Threat Hunting
  threatHunting: {
    hunts: '/api/v2/threat-hunting/hunts',
    getHunt: (id: string) => `/api/v2/threat-hunting/hunts/${id}`,
    createHunt: '/api/v2/threat-hunting/hunts',
    startHunt: (id: string) => `/api/v2/threat-hunting/${id}/start`,
    stopHunt: (id: string) => `/api/v2/threat-hunting/${id}/stop`,
    results: (id: string) => `/api/v2/threat-hunting/${id}/results`,
    indicators: '/api/v2/threat-hunting/indicators',
  },

  // Blockchain
  blockchain: {
    auditLogs: '/api/v2/blockchain/audit-logs',
    blocks: '/api/v2/blockchain/blocks',
    getBlock: (id: string) => `/api/v2/blockchain/blocks/${id}`,
    transactions: '/api/v2/blockchain/transactions',
    getTransaction: (id: string) => `/api/v2/blockchain/transactions/${id}`,
    integrity: '/api/v2/blockchain/integrity',
    verify: '/api/v2/blockchain/verify',
    addLog: '/api/v2/blockchain/audit-log',
  },

  // Zero Trust
  zeroTrust: {
    policies: '/api/v2/zerotrust/policies',
    getPolicy: (id: string) => `/api/v2/zerotrust/policies/${id}`,
    createPolicy: '/api/v2/zerotrust/policies',
    updatePolicy: (id: string) => `/api/v2/zerotrust/policies/${id}`,
    accessRequests: '/api/v2/zerotrust/access-requests',
    trustScores: '/api/v2/zerotrust/trust-scores',
    networkSegments: '/api/v2/zerotrust/network-segments',
  },

  // Quantum
  quantum: {
    algorithms: '/api/v2/quantum/algorithms',
    getAlgorithm: (id: string) => `/api/v2/quantum/algorithms/${id}`,
    keys: '/api/v2/quantum/keys',
    getKey: (id: string) => `/api/v2/quantum/keys/${id}`,
    generateKey: '/api/v2/quantum/generate-key',
    rotateKeys: '/api/v2/quantum/rotate-keys',
    certificates: '/api/v2/quantum/certificates',
    metrics: '/api/v2/quantum/metrics',
    verify: '/api/v2/quantum/verify',
  },

  // Threat Modeling
  threatModeling: {
    models: '/api/v2/threat-models',
    getModel: (id: string) => `/api/v2/threat-models/${id}`,
    scenarios: '/api/v2/threat-models/scenarios',
    getScenario: (id: string) => `/api/v2/threat-models/scenarios/${id}`,
    simulations: '/api/v2/threat-models/simulations',
    getSimulation: (id: string) => `/api/v2/threat-models/simulations/${id}`,
    mitigations: '/api/v2/threat-models/mitigations',
  },

  // System
  system: {
    health: '/api/v2/system/health',
    config: '/api/v2/system/config',
    logs: '/api/v2/system/logs',
    version: '/api/v2/system/version',
  },
} as const;

// ============================================================================
// Default API Configuration
// ============================================================================

export const DEFAULT_API_CONFIG: ApiEndpointConfig = {
  baseURL: (import.meta as any).env.VITE_API_BASE_URL || '/api/v2',
  timeout: 30000,
  retryAttempts: 3,
  retryDelay: 1000,
};

// ============================================================================
// CSRF Token Handling Utilities
// ============================================================================

const CSRF_TOKEN_KEY = 'hades_csrf_token';
const CSRF_HEADER_NAME = 'X-CSRF-Token';

export const csrfUtils = {
  /**
   * Get CSRF token from storage
   */
  getToken: (): string | null => {
    try {
      return sessionStorage.getItem(CSRF_TOKEN_KEY);
    } catch {
      return null;
    }
  },

  /**
   * Set CSRF token in storage
   */
  setToken: (token: string): void => {
    try {
      sessionStorage.setItem(CSRF_TOKEN_KEY, token);
    } catch (error) {
      console.error('Failed to set CSRF token:', error);
    }
  },

  /**
   * Remove CSRF token from storage
   */
  removeToken: (): void => {
    try {
      sessionStorage.removeItem(CSRF_TOKEN_KEY);
    } catch (error) {
      console.error('Failed to remove CSRF token:', error);
    }
  },

  /**
   * Get CSRF header for requests
   */
  getHeader: (): Record<string, string> => {
    const token = csrfUtils.getToken();
    return token ? { [CSRF_HEADER_NAME]: token } : {};
  },

  /**
   * Check if token exists
   */
  hasToken: (): boolean => {
    return csrfUtils.getToken() !== null;
  },
};

// ============================================================================
// API Error Class
// ============================================================================

export class APIError extends Error implements ApiError {
  code: string;
  details?: Record<string, unknown>;
  status_code: number;

  constructor(
    message: string,
    statusCode: number,
    code?: string,
    details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'APIError';
    this.code = code || `HTTP_${statusCode}`;
    this.status_code = statusCode;
    this.details = details || {};
  }

  static fromResponse(response: Response, data?: unknown): APIError {
    const message = data && typeof data === 'object' && 'message' in data
      ? (data as { message: string }).message
      : response.statusText || 'An error occurred';

    return new APIError(
      message,
      response.status,
      data && typeof data === 'object' && 'code' in data
        ? (data as { code: string }).code
        : undefined,
      data && typeof data === 'object' && 'details' in data
        ? (data as { details: Record<string, unknown> }).details
        : undefined
    );
  }

  static isApiError(error: unknown): error is APIError {
    return error instanceof APIError;
  }
}

// ============================================================================
// Response Handlers
// ============================================================================

export const responseHandlers = {
  /**
   * Handle JSON response
   */
  json: async <T>(response: Response): Promise<T> => {
    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw APIError.fromResponse(response, data);
    }

    const contentType = response.headers.get('content-type');
    if (contentType?.includes('application/json')) {
      const data = await response.json();
      return data.data ?? data;
    }

    return response.text() as Promise<T>;
  },

  /**
   * Handle blob response (for file downloads)
   */
  blob: async (response: Response): Promise<Blob> => {
    if (!response.ok) {
      throw APIError.fromResponse(response);
    }
    return response.blob();
  },

  /**
   * Handle text response
   */
  text: async (response: Response): Promise<string> => {
    if (!response.ok) {
      throw APIError.fromResponse(response);
    }
    return response.text();
  },
};

// ============================================================================
// Request Builder
// ============================================================================

export class RequestBuilder {
  private url: string;
  private config: RequestConfig;

  constructor(baseURL: string = DEFAULT_API_CONFIG.baseURL) {
    this.url = baseURL;
    this.config = {
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    };
  }

  setEndpoint(endpoint: string): this {
    this.url = `${this.url}${endpoint}`;
    return this;
  }

  setMethod(method: string): this {
    this.config.method = method;
    return this;
  }

  setParams(params: Record<string, string | number | boolean>): this {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, String(value));
      }
    });
    const queryString = searchParams.toString();
    if (queryString) {
      this.url = `${this.url}?${queryString}`;
    }
    return this;
  }

  setBody(body: unknown): this {
    this.config.body = JSON.stringify(body);
    return this;
  }

  setTimeout(timeout: number): this {
    this.config.timeout = timeout;
    return this;
  }

  addHeader(key: string, value: string): this {
    this.config.headers = {
      ...this.config.headers,
      [key]: value,
    };
    return this;
  }

  addCsrfToken(): this {
    const csrfHeader = csrfUtils.getHeader();
    if (Object.keys(csrfHeader).length > 0) {
      this.config.headers = {
        ...this.config.headers,
        ...csrfHeader,
      };
    }
    return this;
  }

  addAuthToken(token: string): this {
    this.config.headers = {
      ...this.config.headers,
      'Authorization': `Bearer ${token}`,
    };
    return this;
  }

  build(): { url: string; config: RequestConfig } {
    return { url: this.url, config: this.config };
  }
}

// ============================================================================
// Type Guards
// ============================================================================

export const isApiResponse = <T>(obj: unknown): obj is { data: T } => {
  return obj !== null && typeof obj === 'object' && 'data' in obj;
};

export const isPaginatedResponse = <T>(obj: unknown): obj is { data: T[]; pagination: { page: number } } => {
  return obj !== null && typeof obj === 'object' && 'data' in obj && 'pagination' in obj;
};

export const isApiErrorResponse = (obj: unknown): obj is { error: ApiError } => {
  return obj !== null && typeof obj === 'object' && 'error' in obj;
};

// ============================================================================
// Default Exports
// ============================================================================

export type { ApiEndpointConfig, ApiError };

export default {
  API_ENDPOINTS,
  DEFAULT_API_CONFIG,
  csrfUtils,
  APIError,
  responseHandlers,
  RequestBuilder,
  isApiResponse,
  isPaginatedResponse,
  isApiErrorResponse,
};