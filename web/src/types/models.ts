// Shared frontend models

export interface User {
  id: number | string;
  username: string;
  email?: string;
  role?: string;
  permissions?: string[];
  status?: "active" | "inactive" | string;
  environment?: string;
  [key: string]: unknown;
}

export interface LoginCredentials {
  username?: string;
  password?: string;
  token?: string;
  user?: User | string;
  role?: string;
  [key: string]: unknown;
}

export interface AuthResponse {
  token?: string;
  user?: User;
  [key: string]: unknown;
}

export interface UserFilters {
  search?: string;
  role?: string;
  status?: string;
}

export interface UserStats {
  total_users: number;
  active_users: number;
  inactive_users: number;
  by_role: Record<string, number>;
  by_status: Record<string, number>;
}

// Threat types
export interface Threat {
  id: number | string;
  title?: string;
  description?: string;
  severity?: string;
  status?: string;
  type?: string;
  [key: string]: unknown;
}

export interface ThreatFilters {
  severity?: string;
  status?: string;
  type?: string;
}

export interface ThreatStats {
  total_threats: number;
  by_severity: Record<string, number>;
  by_status: Record<string, number>;
  by_type: Record<string, number>;
}

// Websocket messages
export type WSMessage = {
  type?: string;
  data?: unknown;
  [key: string]: unknown;
};

// Security types
export interface Policy {
  id: number | string;
  name?: string;
  description?: string;
  enabled?: boolean;
  [key: string]: unknown;
}

export interface Vulnerability {
  id: number | string;
  title?: string;
  severity?: string;
  status?: string;
  [key: string]: unknown;
}

export interface AuditLog {
  id: number | string;
  message?: string;
  timestamp?: string;
  [key: string]: unknown;
}

export type SecurityScore = number | { score: number; [key: string]: unknown };
