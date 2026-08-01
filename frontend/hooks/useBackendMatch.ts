/**
 * useBackendMatch.ts
 *
 * WebSocket-first match hook.
 *
 * Flow:
 *  1. User clicks "Quick Match" → apiJoinQueue()
 *  2. Backend emits { type: "match_found" } over WS
 *  3. Hook calls onMatchFound() with the backend matchId + opponent info
 *  4. If the backend is offline (apiHealthCheck failed), onBackendOffline()
 *     is called so page.tsx can fall back to local simulation.
 *
 * The hook also handles:
 *  - Sending answers via WS: { type: "submit_answer", payload: {...} }
 *  - Receiving score_update, match_complete messages
 *  - Clean disconnect on unmount
 */

'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import {
  connectBackendWS,
  wsSend,
  apiJoinQueue,
  apiLeaveQueue,
  apiCreateFriendMatch,
  apiJoinFriendMatch,
  apiHealthCheck,
  MatchFoundPayload,
  ScoreUpdatePayload,
  MatchCompletePayload,
  WSMessage,
} from '@/lib/api';

// ── Types ──────────────────────────────────────────────────────────────────

export interface BackendMatchHookOptions {
  token: string | null;
  /** Called when a match is found via ranked queue */
  onMatchFound: (payload: MatchFoundPayload) => void;
  /** Called when match officially starts with question payload */
  onMatchStart?: (payload: any) => void;
  /** Called when live scores update from the server */
  onScoreUpdate: (payload: ScoreUpdatePayload) => void;
  /** Called when server broadcasts time update tick */
  onTimeUpdate?: (payload: any) => void;
  /** Called when the backend declares the match over */
  onMatchComplete: (payload: MatchCompletePayload | any) => void;
  /** Called when the backend is unreachable — use local simulation instead */
  onBackendOffline: () => void;
}

export interface UseBackendMatchResult {
  /** Whether the WS connection to the backend is live */
  isConnected: boolean;
  /** Join the ranked matchmaking queue for a category */
  joinQueue: (examCategoryId: string) => Promise<void>;
  /** Leave the queue */
  leaveQueue: () => Promise<void>;
  /** Create a private friend room; returns room code */
  createFriendRoom: (examCategoryId: string) => Promise<string>;
  /** Join a private friend room by code */
  joinFriendRoom: (roomCode: string) => Promise<{ match_id: string }>;
  /** Submit an answer for the active backend match */
  submitAnswer: (matchId: string, questionId: string, optionId: string, timeTakenMs: number) => void;
  /** Disconnect the WebSocket (e.g., match over, user quits) */
  disconnect: () => void;
}

// ── Hook ───────────────────────────────────────────────────────────────────

export function useBackendMatch(opts: BackendMatchHookOptions): UseBackendMatchResult {
  const {
    token,
    onMatchFound,
    onMatchStart,
    onScoreUpdate,
    onTimeUpdate,
    onMatchComplete,
    onBackendOffline,
  } = opts;

  const wsRef = useRef<{ send: (msg: object) => void; close: () => void } | null>(null);
  const [isConnected, setIsConnected] = useState(false);

  // ── Open WebSocket when token is available ─────────────────────────────
  useEffect(() => {
    if (!token) return;

    let mounted = true;
    let pingInterval: NodeJS.Timeout | null = null;

    async function connect() {
      const isUp = await apiHealthCheck();
      if (!isUp) {
        if (mounted) onBackendOffline();
        return;
      }

      const conn = connectBackendWS(
        token!,
        (msg: WSMessage) => {
          if (!mounted) return;

          switch (msg.type) {
            case 'connected':
              setIsConnected(true);
              break;

            case 'match_found':
              onMatchFound(msg.payload as MatchFoundPayload);
              break;

            case 'match_start':
              if (onMatchStart) {
                onMatchStart(msg.payload);
              } else {
                onMatchFound(msg.payload as MatchFoundPayload);
              }
              break;

            case 'score_update':
              onScoreUpdate(msg.payload as ScoreUpdatePayload);
              break;

            case 'time_update':
              if (onTimeUpdate) onTimeUpdate(msg.payload);
              break;

            case 'match_end':
            case 'match_complete':
              onMatchComplete(msg.payload as MatchCompletePayload);
              break;

            case 'pong':
              break;

            case 'error':
              console.warn('[WS] backend error:', msg.payload);
              break;

            default:
              break;
          }
        },
        () => {
          if (mounted) setIsConnected(false);
        },
      );

      wsRef.current = conn;

      // Heartbeat ping every 20 seconds
      pingInterval = setInterval(() => {
        if (wsRef.current) {
          wsSend(wsRef.current, 'ping', {});
        }
      }, 20000);
    }

    connect();

    return () => {
      mounted = false;
      if (pingInterval) clearInterval(pingInterval);
      wsRef.current?.close();
      wsRef.current = null;
      setIsConnected(false);
    };
  }, [token]); // eslint-disable-line react-hooks/exhaustive-deps

  // ── Public actions ─────────────────────────────────────────────────────

  const joinQueue = useCallback(async (examCategoryId: string) => {
    if (!token) throw new Error('Not authenticated');
    const isUp = await apiHealthCheck();
    if (!isUp) { onBackendOffline(); return; }
    await apiJoinQueue(token, examCategoryId);
  }, [token, onBackendOffline]);

  const leaveQueue = useCallback(async () => {
    if (!token) return;
    try { await apiLeaveQueue(token); } catch { /* ignore */ }
  }, [token]);

  const createFriendRoom = useCallback(async (examCategoryId: string): Promise<string> => {
    if (!token) throw new Error('Not authenticated');
    const isUp = await apiHealthCheck();
    if (!isUp) throw new Error('Backend offline');
    const resp = await apiCreateFriendMatch(token, examCategoryId);
    return resp.room_code;
  }, [token]);

  const joinFriendRoom = useCallback(async (roomCode: string) => {
    if (!token) throw new Error('Not authenticated');
    return apiJoinFriendMatch(token, roomCode);
  }, [token]);

  const submitAnswer = useCallback(
    (matchId: string, questionId: string, optionId: string, timeTakenMs: number) => {
      if (!wsRef.current || !isConnected) return;
      wsSend(wsRef.current, 'submit_answer', {
        match_id: matchId,
        question_id: questionId,
        option_id: optionId,
        time_taken_ms: timeTakenMs,
      });
    },
    [isConnected],
  );

  const disconnect = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
    setIsConnected(false);
  }, []);

  return {
    isConnected,
    joinQueue,
    leaveQueue,
    createFriendRoom,
    joinFriendRoom,
    submitAnswer,
    disconnect,
  };
}
