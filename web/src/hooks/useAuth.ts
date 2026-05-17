import React, { useState, useContext } from "react";

import { authAPI } from "../api/auth";
import type { User, LoginCredentials, AuthResponse } from "../types/models";

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  login: (credentials: LoginCredentials) => Promise<AuthResponse | unknown>;
  logout: () => Promise<void>;
  refreshData: () => Promise<AuthResponse | unknown>;
  setUser: (user: User | null) => void;
  setIsAuthenticated: (auth: boolean) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

// Global window type declarations
declare global {
  interface Window {
    hadesToken: string | null;
    hadesUser: User | string | null;
    hadesRole: string | null;
    hadesEnvironment: string | null;
  }

  interface ImportMetaEnv {
    VITE_API_BASE_URL: string;
    VITE_WS_BASE_URL?: string;
  }

  interface ImportMeta {
    env: ImportMetaEnv;
  }
}

const AuthContext = React.createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }

  return context;
};

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Check for existing session on mount
  React.useEffect(() => {
    (async () => {
      const token = window.hadesToken || null;
      const rawUser = window.hadesUser || null;
      const environment = window.hadesEnvironment || null;

      // Check if we're in development environment
      const hostname = window.location.hostname || "";
      const isDevelopment =
        hostname === "localhost" ||
        hostname === "127.0.0.1" ||
        hostname === "192.168.0.2" ||
        hostname.includes("dev") ||
        hostname.includes("test") ||
        hostname.includes("qa") ||
        hostname.includes("staging");

      if (token && rawUser) {
        // We have a stored session, restore it
        const parsedUser =
          typeof rawUser === "string" ? JSON.parse(rawUser) : rawUser;
        setUser(parsedUser as User);
        setIsAuthenticated(true);

        if (environment) {
          // Session restored silently
        }
      } else if (isDevelopment) {
        // For development, get a real JWT token from backend
        const apiUrl =
          (import.meta as any).env.VITE_API_BASE_URL || "http://localhost:8080";
        try {
          const response = await fetch(`${apiUrl}/api/v1/auth/login`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              username: "admin",
              password: "admin123",
            }),
          });

          if (response.ok) {
            const data = await response.json();
            const defaultUser = data.data?.user;
            const realToken = data.data?.token;

            setUser(defaultUser);
            setIsAuthenticated(true);
            window.hadesToken = realToken;
            window.hadesUser = defaultUser;
            window.hadesRole = "Administrator";
            window.hadesEnvironment = "development";

            // Development session created with real JWT token
          } else {
            // Fallback to fake token if backend is not available
            const defaultUser: User = {
              id: 1,
              username: "admin",
              email: "admin@hades-toolkit.com",
              role: "Administrator",
              permissions: ["read", "write", "admin"],
            };

            setUser(defaultUser);
            setIsAuthenticated(true);
            window.hadesToken = "dev-token-" + Date.now();
            window.hadesUser = defaultUser;
            window.hadesRole = "Administrator";
            window.hadesEnvironment = "development";

            // Fallback development session created
          }
        } catch (err) {
          console.error("Failed to get dev token:", err);
          // Fallback to fake token
          const defaultUser: User = {
            id: 1,
            username: "admin",
            email: "admin@hades-toolkit.com",
            role: "Administrator",
            permissions: ["read", "write", "admin"],
          };

          setUser(defaultUser);
          setIsAuthenticated(true);
          window.hadesToken = "dev-token-" + Date.now();
          window.hadesUser = defaultUser;
          window.hadesRole = "Administrator";
          window.hadesEnvironment = "development";

          // Fallback development session created
        }
      }

      setLoading(false);
    })();
  }, []);

  type LoginCredentialsLocal = LoginCredentials;

  const login = async (
    credentials: LoginCredentialsLocal,
  ): Promise<AuthResponse | unknown> => {
    setLoading(true);
    setError(null);

    try {
      // Check if this is a development environment login with role-based data
      const hostname = window.location.hostname || "";
      const isDevelopment =
        hostname === "localhost" ||
        hostname === "127.0.0.1" ||
        hostname.includes("dev") ||
        hostname.includes("test") ||
        hostname.includes("qa") ||
        hostname.includes("staging");

      if (isDevelopment && credentials.user && credentials.token) {
        // This is a development login with realistic data
        const response = credentials as AuthResponse;

        // Store all session data in secure memory
        window.hadesToken = response.token as string;
        window.hadesUser = response.user as User;
        window.hadesRole = credentials.role || "user";
        window.hadesEnvironment =
          (response.user as any)?.environment || "development";

        setUser(response.user as User);
        setIsAuthenticated(true);

        return response;
      } else {
        // Production authentication (or fallback)
        const loginCredentials = {
          ...credentials,
          role: (credentials.role as string) || "user",
        };

        const response = (await authAPI.login(
          loginCredentials as { username: string; password: string },
        )) as AuthResponse;

        // Store token in secure memory
        window.hadesToken = response.token as string;
        window.hadesUser = response.user as User;
        window.hadesRole = credentials.role as string;

        setUser(response.user as User);
        setIsAuthenticated(true);

        return response;
      }
    } catch (err) {
      setError("Invalid credentials. Please try again.");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    setLoading(true);

    try {
      // Call logout API if available (only in production)
      const hostname = window.location.hostname || "";
      const isDevelopment =
        hostname === "localhost" ||
        hostname === "127.0.0.1" ||
        hostname.includes("dev") ||
        hostname.includes("test") ||
        hostname.includes("qa") ||
        hostname.includes("staging");

      if (!isDevelopment) {
        await authAPI.logout();
      }
    } catch (err) {
      console.error("Logout error:", err);
    } finally {
      // Clear secure memory and state
      window.hadesToken = null;
      window.hadesUser = null;
      window.hadesRole = null;
      window.hadesEnvironment = null;

      // Clear authentication state
      setUser(null);
      setIsAuthenticated(false);
      setLoading(false);

      // Redirect to login page
      window.location.replace("/login");
    }
  };

  const refreshToken = async (): Promise<AuthResponse | unknown> => {
    try {
      const response = (await authAPI.refreshToken()) as AuthResponse;

      // Update token in secure memory
      window.hadesToken = response.token as string;

      return response;
    } catch (err) {
      // Refresh failed, logout user
      await logout();
      throw err;
    }
  };

  const value: AuthContextType = {
    user,
    isAuthenticated,
    loading,
    error,
    login,
    logout,
    refreshData: refreshToken,
    setUser,
    setIsAuthenticated,
    setLoading,
    setError,
  };

  return React.createElement(AuthContext.Provider, { value: value }, children);
};

export default useAuth;
