import { api } from './client'

// ---- Auth ----
export const registerUser = (body) => api.post('/api/v1/auth/register', body)
export const loginUser = (body) => api.post('/api/v1/auth/login', body)
export const getMe = () => api.get('/api/v1/auth/me')

// ---- Users ----
export const getUserProfile = (id) => api.get(`/api/v1/users/${id}`)
export const getUserStats = (id) => api.get(`/api/v1/users/${id}/stats`)
export const getUserMatchHistory = (id) => api.get(`/api/v1/users/${id}/matches`)

// ---- Subjects ----
export const getSubjects = () => api.get('/api/v1/subjects')

// ---- Matchmaking ----
export const joinQueue = (body) => api.post('/api/v1/matches/queue', body)
export const leaveQueue = () => api.delete('/api/v1/matches/queue')
export const getQueueStats = () => api.get('/api/v1/matches/queue/stats')

// ---- Matches ----
export const getMatch = (id) => api.get(`/api/v1/matches/${id}`)
export const createFriendMatch = (body) => api.post('/api/v1/matches/friend', body)
export const joinFriendMatch = (body) => api.post('/api/v1/matches/friend/join', body)

// ---- Practice ----
export const startPractice = (body) => api.post('/api/v1/practice/start', body)
export const submitPracticeAnswer = (id, body) => api.post(`/api/v1/practice/${id}/answer`, body)
export const endPractice = (id) => api.post(`/api/v1/practice/${id}/end`)
export const getPractice = (id) => api.get(`/api/v1/practice/${id}`)

// ---- Leaderboard ----
export const getLeaderboard = (categoryCode, params) =>
  api.get(`/api/v1/leaderboard/${categoryCode}`, { params })

// ---- Admin ----
export const createQuestion = (body) => api.post('/api/v1/admin/questions', body)
export const publishQuestion = (id) => api.put(`/api/v1/admin/questions/${id}/publish`)
export const getAdminStats = () => api.get('/api/v1/admin/stats')
