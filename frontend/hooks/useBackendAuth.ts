/**
 * useBackendAuth.ts
 *
 * Dual authentication hook:
 *   1. Email + Password  →  Go backend JWT  (stored as 'ea_jwt' in localStorage)
 *   2. Google Sign-In    →  Firebase Auth   (existing Firebase flow)
 *
 * On mount the hook re-validates any stored JWT and keeps the
 * backend user object in sync with the React component tree.
 */

'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  apiLogin,
  apiRegister,
  apiLoginWithFirebase,
  apiGetMe,
  apiHealthCheck,
  BackendUser,
} from '@/lib/api';
import {
  signInWithGoogle,
  signInWithEmailAndPasswordUser,
  signUpWithEmailAndPassword,
  logOutUser,
} from '@/lib/firestoreService';
import { getAuthToken, saveAuthToken, clearAuthToken } from '@/lib/storage';

// ── Types ──────────────────────────────────────────────────────────────────

export type AuthMode = 'backend' | 'google' | 'none';

export interface UseBackendAuthResult {
  /** JWT token from Go backend; null if not authenticated via backend */
  token: string | null;
  /** User object from Go backend; null if using Google-only auth */
  backendUser: BackendUser | null;
  /** Which auth mode is active */
  authMode: AuthMode;
  loading: boolean;
  error: string | null;
  /** Email + password login via Go backend */
  loginWithPassword: (login: string, password: string) => Promise<void>;
  /** Register new account via Go backend */
  registerWithPassword: (
    username: string,
    email: string,
    password: string,
  ) => Promise<void>;
  /** Google OAuth via Firebase */
  loginWithGoogle: () => Promise<void>;
  /** Signs out of whichever method is active */
  logout: () => Promise<void>;
  /** Clear any auth error */
  clearError: () => void;
}

// ── Hook ───────────────────────────────────────────────────────────────────

export function useBackendAuth(): UseBackendAuthResult {
  const [token, setToken] = useState<string | null>(null);
  const [backendUser, setBackendUser] = useState<BackendUser | null>(null);
  const [authMode, setAuthMode] = useState<AuthMode>('none');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // ── On mount: restore JWT from storage and validate it ──────────────────
  useEffect(() => {
    async function restore() {
      const stored = getAuthToken();
      if (!stored) {
        setLoading(false);
        return;
      }
      try {
        const isUp = await apiHealthCheck();
        if (!isUp) {
          // Backend offline — keep token in storage for when it comes back
          setToken(stored);
          setAuthMode('backend');
          setLoading(false);
          return;
        }
        const user = await apiGetMe(stored);
        setToken(stored);
        setBackendUser(user);
        setAuthMode('backend');
      } catch {
        // Token expired or invalid — clear it
        clearAuthToken();
      } finally {
        setLoading(false);
      }
    }
    restore();
  }, []);

  // ── Email + Password login ───────────────────────────────────────────────
  const loginWithPassword = useCallback(
    async (login: string, password: string) => {
      setLoading(true);
      setError(null);
      try {
        const resp = await apiLogin(login, password);
        saveAuthToken(resp.token);
        setToken(resp.token);
        setBackendUser(resp.user);
        setAuthMode('backend');
      } catch (err: unknown) {
        // Fallback to Firebase Auth if backend API is offline or returns error
        try {
          await signInWithEmailAndPasswordUser(login, password);
          setAuthMode('google'); // Firebase Auth active session
        } catch (fbErr: unknown) {
          const msg = fbErr instanceof Error ? fbErr.message : (err instanceof Error ? err.message : 'Login failed');
          setError(msg);
          throw fbErr;
        }
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // ── Email + Password register ────────────────────────────────────────────
  const registerWithPassword = useCallback(
    async (username: string, email: string, password: string) => {
      setLoading(true);
      setError(null);
      try {
        const resp = await apiRegister(username, email, password);
        saveAuthToken(resp.token);
        setToken(resp.token);
        setBackendUser(resp.user);
        setAuthMode('backend');
      } catch (err: unknown) {
        // Fallback to Firebase Auth if backend API is offline or returns error
        try {
          await signUpWithEmailAndPassword(email, password, username);
          setAuthMode('google'); // Firebase Auth active session
        } catch (fbErr: unknown) {
          const msg = fbErr instanceof Error ? fbErr.message : (err instanceof Error ? err.message : 'Registration failed');
          setError(msg);
          throw fbErr;
        }
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // ── Google Sign-In (Firebase + Go Backend Sync) ─────────────────────────
  const loginWithGoogle = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const fbUser = await signInWithGoogle();
      if (fbUser) {
        try {
          const idToken = await fbUser.getIdToken();
          const resp = await apiLoginWithFirebase(
            idToken,
            fbUser.email || '',
            fbUser.displayName || undefined,
          );
          saveAuthToken(resp.token);
          setToken(resp.token);
          setBackendUser(resp.user);
          setAuthMode('backend');
        } catch {
          // Fallback to Google Auth mode if Go backend is offline
          setAuthMode('google');
        }
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Google sign-in failed';
      setError(msg);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // ── Logout ───────────────────────────────────────────────────────────────
  const logout = useCallback(async () => {
    setLoading(true);
    try {
      if (authMode === 'google') {
        await logOutUser();
      }
      clearAuthToken();
      setToken(null);
      setBackendUser(null);
      setAuthMode('none');
    } finally {
      setLoading(false);
    }
  }, [authMode]);

  const clearError = useCallback(() => setError(null), []);

  return {
    token,
    backendUser,
    authMode,
    loading,
    error,
    loginWithPassword,
    registerWithPassword,
    loginWithGoogle,
    logout,
    clearError,
  };
}
