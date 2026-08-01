/**
 * api.ts — Typed client for the Exam Arena Go backend.
 *
 * All HTTP calls go to NEXT_PUBLIC_API_URL (http://localhost:8080).
 * WebSocket connects to NEXT_PUBLIC_WS_URL (ws://localhost:8080).
 *
 * Every function throws on network failure so callers can decide
 * whether to fall back to local/Firebase data.
 */

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const WS_URL  = process.env.NEXT_PUBLIC_WS_URL  || 'ws://localhost:8080';

// ── Helpers ────────────────────────────────────────────────────────────────

async function request<T>(
  path: string,
  options: RequestInit = {},
  token?: string,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, { ...options, headers });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body?.error || `HTTP ${res.status}`);
  }

  return res.json();
}

// ── Auth types ─────────────────────────────────────────────────────────────

export interface BackendUser {
  id: string;
  username: string;
  email: string;
  display_name: string | null;
  avatar_url: string | null;
  role: string;
  status: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: BackendUser;
}

// ── Leaderboard types ──────────────────────────────────────────────────────

export interface BackendLeaderboardEntry {
  rank: number;
  user_id: string;
  username: string;
  display_name: string | null;
  avatar_url: string | null;
  rating: number;
  matches_played: number;
}

export interface LeaderboardResponse {
  data: BackendLeaderboardEntry[];
  meta: { limit: number; offset: number; category: string };
}

// ── User Stats & Match History types ──────────────────────────────────────

export interface UserStats {
  user_id: string;
  rating: number;
  highest_rating: number;
  matches_played: number;
  wins: number;
  losses: number;
  draws: number;
  questions_answered: number;
  correct_answers: number;
  average_response_ms: number;
  win_rate: number;
  accuracy: number;
}

export interface PracticeSessionResponse {
  id: string;
  user_id: string;
  category_id: string;
  status: string;
  questions: Array<{
    id: string;
    statement: string;
    options: Array<{ id: string; option_text: string }>;
  }>;
}

// ── Auth API ───────────────────────────────────────────────────────────────

/**
 * Register with email, username, and password.
 * Backend: POST /api/v1/auth/register
 */
export async function apiRegister(
  username: string,
  email: string,
  password: string,
): Promise<AuthResponse> {
  return request<AuthResponse>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, email, password }),
  });
}

/**
 * Login / Register using Firebase ID Token.
 * Backend: POST /api/v1/auth/firebase
 */
export async function apiLoginWithFirebase(
  idToken: string,
  email: string,
  username?: string,
): Promise<AuthResponse> {
  return request<AuthResponse>('/api/v1/auth/firebase', {
    method: 'POST',
    body: JSON.stringify({ id_token: idToken, email, username }),
  });
}

/**
 * Login with username-or-email + password.
 * Backend: POST /api/v1/auth/login
 */
export async function apiLogin(
  login: string,
  password: string,
): Promise<AuthResponse> {
  return request<AuthResponse>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ login, password }),
  });
}

/**
 * Fetch the currently authenticated user.
 * Backend: GET /api/v1/auth/me
 */
export async function apiGetMe(token: string): Promise<BackendUser> {
  return request<BackendUser>('/api/v1/auth/me', {}, token);
}

// ── User API ───────────────────────────────────────────────────────────────

export async function apiGetUserProfile(id: string): Promise<BackendUser> {
  return request<BackendUser>(`/api/v1/users/${id}`);
}

export async function apiGetUserStats(id: string): Promise<UserStats> {
  return request<UserStats>(`/api/v1/users/${id}/stats`);
}

export async function apiGetUserMatches(id: string): Promise<any[]> {
  return request<any[]>(`/api/v1/users/${id}/matches`);
}

// ── Practice API ───────────────────────────────────────────────────────────

export async function apiStartPractice(
  token: string,
  categoryId: string,
  count: number = 10,
): Promise<PracticeSessionResponse> {
  return request<PracticeSessionResponse>(
    '/api/v1/practice/start',
    {
      method: 'POST',
      body: JSON.stringify({ category_id: categoryId, count }),
    },
    token,
  );
}

export async function apiSubmitPracticeAnswer(
  token: string,
  sessionId: string,
  questionId: string,
  optionId: string,
  timeTakenMs: number,
): Promise<{ is_correct: boolean; explanation?: string }> {
  return request(
    `/api/v1/practice/${sessionId}/answer`,
    {
      method: 'POST',
      body: JSON.stringify({
        question_id: questionId,
        option_id: optionId,
        time_taken_ms: timeTakenMs,
      }),
    },
    token,
  );
}

export async function apiEndPracticeSession(
  token: string,
  sessionId: string,
): Promise<{ session_id: string; total_score: number; accuracy: number }> {
  return request(
    `/api/v1/practice/${sessionId}/end`,
    { method: 'POST' },
    token,
  );
}

// ── Match API ──────────────────────────────────────────────────────────────

/**
 * Join the ranked matchmaking queue.
 * Backend: POST /api/v1/matches/queue
 */
export async function apiJoinQueue(
  token: string,
  examCategoryId: string,
): Promise<{ status: string; message: string }> {
  return request(
    '/api/v1/matches/queue',
    { method: 'POST', body: JSON.stringify({ exam_category_id: examCategoryId, match_type: 'ranked' }) },
    token,
  );
}

/**
 * Leave the ranked matchmaking queue.
 * Backend: DELETE /api/v1/matches/queue
 */
export async function apiLeaveQueue(token: string): Promise<void> {
  await request('/api/v1/matches/queue', { method: 'DELETE' }, token);
}

export async function apiGetQueueStats(token: string): Promise<{ active_players: number; total_queued: number }> {
  return request('/api/v1/matches/queue/stats', {}, token);
}

/**
 * Create a private friend match and receive a room code.
 * Backend: POST /api/v1/matches/friend
 */
export async function apiCreateFriendMatch(
  token: string,
  examCategoryId: string,
): Promise<{ match_id: string; room_code: string; status: string }> {
  return request(
    '/api/v1/matches/friend',
    { method: 'POST', body: JSON.stringify({ exam_category_id: examCategoryId }) },
    token,
  );
}

/**
 * Join a friend's private room by room code.
 * Backend: POST /api/v1/matches/friend/join
 */
export async function apiJoinFriendMatch(
  token: string,
  roomCode: string,
): Promise<{ match_id: string; status: string }> {
  return request(
    '/api/v1/matches/friend/join',
    { method: 'POST', body: JSON.stringify({ room_code: roomCode }) },
    token,
  );
}

// ── Leaderboard API ────────────────────────────────────────────────────────

/**
 * Fetch the leaderboard for a given category code.
 * Backend: GET /api/v1/leaderboard/{category}
 */
export async function apiGetLeaderboard(
  category: string,
  limit = 50,
  offset = 0,
): Promise<LeaderboardResponse> {
  return request<LeaderboardResponse>(
    `/api/v1/leaderboard/${encodeURIComponent(category)}?limit=${limit}&offset=${offset}`,
  );
}

// ── Health check ───────────────────────────────────────────────────────────

/**
 * Ping the backend health endpoint.
 * Returns true if the backend is reachable.
 */
export async function apiHealthCheck(): Promise<boolean> {
  try {
    await fetch(`${API_URL}/health`, { signal: AbortSignal.timeout(3000) });
    return true;
  } catch {
    return false;
  }
}

// ── WebSocket ──────────────────────────────────────────────────────────────

export type WSMessageType =
  | 'connected'
  | 'match_found'
  | 'match_start'
  | 'score_update'
  | 'time_update'
  | 'match_complete'
  | 'match_end'
  | 'match_failed'
  | 'submit_answer'
  | 'ping'
  | 'pong'
  | 'error'
  | string;

export interface WSMessage<T = any> {
  type: WSMessageType;
  payload: T;
  req_id?: string;
}

export interface MatchFoundPayload {
  match_id: string;
  room_code?: string;
  opponent: {
    user_id: string;
    username: string;
    avatar_url: string | null;
    rating: number;
  };
}

export interface MatchStartPayload {
  match_id: string;
  match_type: string;
  timer_seconds: number;
  questions: Array<{
    id: string;
    body: string;
    options: Array<{ id: string; option_text: string }>;
  }>;
  players: Array<{
    user_id: string;
    username: string;
    rating?: number;
  }>;
}

export interface ScoreUpdatePayload {
  match_id: string;
  user_id?: string;
  question_id?: string;
  is_correct?: boolean;
  points_earned?: number;
  scores: Record<string, number>; // user_id → score
  scoreboard?: Array<{
    user_id: string;
    username: string;
    score: number;
    correct: number;
  }>;
}

export interface TimeUpdatePayload {
  match_id: string;
  remaining_seconds: number;
}

export interface MatchEndPayload {
  match_id: string;
  results: Array<{
    user_id: string;
    username: string;
    score: number;
    rank: number;
    correct: number;
    total: number;
  }>;
}

export interface MatchCompletePayload {
  match_id: string;
  winner_id: string | 'draw';
  results: Array<{
    user_id: string;
    username: string;
    score: number;
    rank: number;
    rating_delta: number;
  }>;
}

/**
 * Open a WebSocket connection to the backend game hub.
 * Auth: JWT is passed as ?token= query parameter (browser WS limitation).
 *
 * @param token  - JWT from the Go backend auth flow
 * @param onMessage - callback for every incoming message
 * @param onClose   - called when the socket closes (incl. error)
 * @returns send(msg) function + close() function
 */
export function connectBackendWS(
  token: string,
  onMessage: (msg: WSMessage) => void,
  onClose?: () => void,
): { send: (msg: object) => void; close: () => void } {
  const url = `${WS_URL}/ws?token=${encodeURIComponent(token)}`;
  const ws = new WebSocket(url);

  ws.onmessage = (event) => {
    try {
      const msg: WSMessage = JSON.parse(event.data);
      onMessage(msg);
    } catch {
      // ignore malformed frames
    }
  };

  ws.onclose = () => onClose?.();
  ws.onerror = () => onClose?.();

  const send = (msg: object) => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(msg));
    }
  };

  const close = () => ws.close();

  return { send, close };
}

/**
 * Send a typed WebSocket message.
 */
export function wsSend(
  ws: { send: (msg: object) => void },
  type: string,
  payload: object,
) {
  ws.send({ type, payload });
}

