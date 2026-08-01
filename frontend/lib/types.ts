export type QuestionCategory = 
  | 'Mathematics'
  | 'English Language'
  | 'Logical Reasoning'
  | 'Mathematics & Logic'
  | 'General Knowledge'
  | 'Science & Physics'
  | 'Computer Science'
  | 'World History'
  | 'Medical & Biology'
  | 'GRE / SAT Prep'
  | 'Custom AI Topic';

export interface Question {
  id: string;
  category: QuestionCategory | string;
  question: string;
  options: string[];
  rawOptionIds?: string[];
  correctAnswerIndex: number;
  explanation: string;
  difficulty: 'Easy' | 'Medium' | 'Hard';
  stepByStepAnalysis?: string[];
  learningConcept?: string;
  formulaOrRule?: string;
}

export interface PlayerAnswer {
  questionId: string;
  selectedOptionIndex: number;
  responseTimeMs: number;
  isCorrect: boolean;
  scoreGained: number;
  speedBonus: number;
  streakBonus: number;
}

export interface PlayerProfile {
  id: string;
  username: string;
  avatarUrl: string;
  title: string;
  mmr: number;
  totalMatches: number;
  wins: number;
  losses: number;
  totalPoints: number;
  avgResponseTimeMs: number;
  accuracyPercentage: number;
  highestStreak: number;
  currentStreak: number;
  unlockedAchievements: string[];
  favoriteCategory: QuestionCategory;
  status: 'online' | 'in_match' | 'offline';
  bio?: string;
  joinedDate: string;
}

export interface MatchPlayer {
  id: string;
  username: string;
  avatarUrl: string;
  title: string;
  mmr: number;
  score: number;
  answers: PlayerAnswer[];
  isReady: boolean;
  isBot?: boolean;
  botDifficulty?: 'Beginner' | 'Intermediate' | 'Master' | 'Genius';
  currentAnswerStatus?: {
    answered: boolean;
    responseTimeMs?: number;
    isCorrect?: boolean;
  };
}

export const SEARCHING_MATCH_PLAYER: MatchPlayer = {
  id: 'waiting_opponent',
  username: 'Searching Opponent...',
  avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250',
  title: '⚡ Duelist',
  mmr: 1000,
  score: 0,
  answers: [],
  isReady: false,
};

export interface MatchEmote {
  id: string;
  senderId: string;
  senderName: string;
  text: string;
  icon: string;
  timestamp: number;
}

export interface MatchState {
  id: string;
  roomCode?: string;
  category: QuestionCategory | string;
  status: 'waiting' | 'starting' | 'in_progress' | 'round_ended' | 'completed';
  player1: MatchPlayer;
  player2: MatchPlayer;
  questions: Question[];
  currentQuestionIndex: number;
  questionStartTime: number | null;
  timeLimitSeconds: number;
  winnerId?: string | 'draw';
  emotes: MatchEmote[];
  createdAt: number;
}

export interface Achievement {
  id: string;
  title: string;
  description: string;
  icon: string;
  rarity: 'Common' | 'Rare' | 'Epic' | 'Legendary';
  category: 'Speed' | 'Wins' | 'Accuracy' | 'Social' | 'Mastery';
  progress: number;
  maxProgress: number;
  isUnlocked: boolean;
  unlockedAt?: string;
}

export interface LeaderboardEntry {
  rank: number;
  id: string;
  username: string;
  avatarUrl: string;
  title: string;
  mmr: number;
  wins: number;
  totalMatches: number;
  winRate: number;
  avgResponseTimeMs: number;
  accuracyPercentage: number;
  badge: string;
}

export interface Friend {
  id: string;
  username: string;
  avatarUrl: string;
  title: string;
  mmr: number;
  status: 'online' | 'in_match' | 'offline';
  winRate: number;
  isFavorite?: boolean;
}

export interface ChatMessage {
  id: string;
  senderId: string;
  senderName: string;
  senderAvatar: string;
  senderTitle?: string;
  text: string;
  timestamp: number;
  isSystem?: boolean;
}

export interface AppNotification {
  id: string;
  type: 'challenge' | 'turn_update' | 'friend_request' | 'achievement' | 'match_result';
  title: string;
  message: string;
  timestamp: number;
  read: boolean;
  actionPayload?: {
    matchId?: string;
    roomCode?: string;
    friendId?: string;
  };
}
