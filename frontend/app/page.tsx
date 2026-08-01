'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { 
  INITIAL_USER_PROFILE,
  INITIAL_ACHIEVEMENTS,
  INITIAL_FRIENDS,
  INITIAL_NOTIFICATIONS,
  INITIAL_GLOBAL_CHAT,
  getStoredProfile, 
  saveStoredProfile, 
  getStoredAchievements, 
  saveStoredAchievements, 
  getStoredFriends, 
  saveStoredFriends, 
  getStoredNotifications, 
  saveStoredNotifications, 
  getStoredGlobalChat, 
  saveStoredGlobalChat, 
  MOCK_LEADERBOARD, 
  SAMPLE_QUESTIONS, 
  arenaBroadcast 
} from '@/lib/storage';
import { 
  PlayerProfile, 
  Achievement, 
  LeaderboardEntry, 
  Friend, 
  AppNotification, 
  ChatMessage, 
  MatchState, 
  MatchPlayer, 
  QuestionCategory, 
  Question, 
  PlayerAnswer 
} from '@/lib/types';
import { Navbar } from '@/components/Navbar';
import { NotificationCenter } from '@/components/NotificationCenter';
import { MatchmakingLobby } from '@/components/MatchmakingLobby';
import { DualExamArena } from '@/components/DualExamArena';
import { MatchSummaryModal } from '@/components/MatchSummaryModal';
import { LeaderboardView } from '@/components/LeaderboardView';
import { ProfileView } from '@/components/ProfileView';
import { FriendsAndChatView } from '@/components/FriendsAndChatView';
import { notificationManager } from '@/lib/notification';
import { soundManager } from '@/lib/audio';
import { 
  listenToAuthAndProfile, 
  saveProfileToFirestore, 
  listenToGlobalChat, 
  sendGlobalChatMessage, 
  saveMatchToFirestore,
  listenToNotifications
} from '@/lib/firestoreService';
import { useBackendAuth } from '@/hooks/useBackendAuth';
import { useBackendMatch } from '@/hooks/useBackendMatch';
import { apiGetLeaderboard } from '@/lib/api';
import type { MatchFoundPayload, ScoreUpdatePayload, MatchCompletePayload } from '@/lib/api';
import { AuthPage } from '@/components/AuthPage';

export default function Home() {
  const [activeTab, setActiveTab] = useState<'lobby' | 'leaderboard' | 'profile' | 'social'>('lobby');
  const [userProfile, setUserProfile] = useState<PlayerProfile>(INITIAL_USER_PROFILE);
  const [authUser, setAuthUser] = useState<any>(null);
  const [achievements, setAchievements] = useState<Achievement[]>(INITIAL_ACHIEVEMENTS);
  const [friends, setFriends] = useState<Friend[]>(INITIAL_FRIENDS);
  const [notifications, setNotifications] = useState<AppNotification[]>(INITIAL_NOTIFICATIONS);
  const [globalChat, setGlobalChat] = useState<ChatMessage[]>(INITIAL_GLOBAL_CHAT);
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>(MOCK_LEADERBOARD);
  const [backendOffline, setBackendOffline] = useState<boolean>(false);

  const [activeMatch, setActiveMatch] = useState<MatchState | null>(null);
  const [showSummaryModal, setShowSummaryModal] = useState<boolean>(false);
  const [soundEnabled, setSoundEnabled] = useState<boolean>(true);
  const [isNotificationCenterOpen, setIsNotificationCenterOpen] = useState<boolean>(false);

  // ── Backend Auth (JWT email/password + Firebase Google) ──────────────────
  const backendAuth = useBackendAuth();

  // ── WebSocket Match Handlers ──────────────────────────────────────────────
  const handleMatchFound = useCallback((payload: MatchFoundPayload) => {
    const newNotif: AppNotification = {
      id: 'n_' + Date.now(),
      type: 'turn_update',
      title: 'Match Found!',
      message: `Opponent: ${payload.opponent.username} — match starting!`,
      timestamp: Date.now(),
      read: false,
      actionPayload: { matchId: payload.match_id },
    };
    setNotifications((prev) => [newNotif, ...prev]);
  }, []);

  const handleMatchStart = useCallback((payload: any) => {
    if (!payload || !payload.match_id) return;

    const backendQuestions: Question[] = (payload.questions || []).map((q: any) => ({
      id: q.id,
      category: 'Backend Exam',
      question: q.body || q.statement || 'Question',
      options: (q.options || []).map((o: any) => o.option_text || o.text || ''),
      rawOptionIds: (q.options || []).map((o: any) => o.id || ''),
      correctAnswerIndex: -1,
      explanation: 'Server authoritative',
      difficulty: q.difficulty || 'Medium',
    }));

    const p1User = payload.players?.find((p: any) => p.user_id === userProfile.id) || payload.players?.[0];
    const p2User = payload.players?.find((p: any) => p.user_id !== userProfile.id) || payload.players?.[1];

    const p1: MatchPlayer = {
      id: p1User?.user_id || userProfile.id,
      username: p1User?.username || userProfile.username,
      avatarUrl: userProfile.avatarUrl,
      title: userProfile.title,
      mmr: p1User?.rating || userProfile.mmr,
      score: 0,
      answers: [],
      isReady: true,
      isBot: false,
    };

    const p2: MatchPlayer = {
      id: p2User?.user_id || 'opponent_id',
      username: p2User?.username || 'Opponent',
      avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250',
      title: '⚡ Duelist',
      mmr: p2User?.rating || 1200,
      score: 0,
      answers: [],
      isReady: true,
      isBot: false,
    };

    const newMatch: MatchState = {
      id: payload.match_id,
      category: payload.match_type || 'Ranked Match',
      status: 'in_progress',
      player1: p1,
      player2: p2,
      questions: backendQuestions.length > 0 ? backendQuestions : SAMPLE_QUESTIONS['Computer Science'],
      currentQuestionIndex: 0,
      questionStartTime: Date.now(),
      timeLimitSeconds: payload.timer_seconds || 120,
      emotes: [],
      createdAt: Date.now(),
    };

    setActiveMatch(newMatch);
    setShowSummaryModal(false);
  }, [userProfile]);

  const handleScoreUpdate = useCallback((payload: ScoreUpdatePayload) => {
    setActiveMatch((prev) => {
      if (!prev) return null;
      let p1Score = prev.player1.score;
      let p2Score = prev.player2.score;

      if (payload.scores) {
        p1Score = payload.scores[prev.player1.id] ?? p1Score;
        p2Score = payload.scores[prev.player2.id] ?? p2Score;
      }

      if (payload.scoreboard) {
        for (const item of payload.scoreboard) {
          if (item.user_id === prev.player1.id) p1Score = item.score;
          if (item.user_id === prev.player2.id) p2Score = item.score;
        }
      }

      return {
        ...prev,
        player1: { ...prev.player1, score: p1Score },
        player2: { ...prev.player2, score: p2Score },
      };
    });
  }, []);

  const handleTimeUpdate = useCallback((payload: any) => {
    if (!payload || typeof payload.remaining_seconds !== 'number') return;
    setActiveMatch((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        timeLimitSeconds: payload.remaining_seconds,
      };
    });
  }, []);

  const handleMatchComplete = useCallback((payload: MatchCompletePayload) => {
    setActiveMatch((prev) =>
      prev ? { ...prev, status: 'completed', winnerId: payload.winner_id } : null,
    );
    setShowSummaryModal(true);
  }, []);

  const handleBackendOffline = useCallback(() => {
    setBackendOffline(true);
  }, []);

  // ── Backend Match Hook ───────────────────────────────────────────────────
  const backendMatch = useBackendMatch({
    token: backendAuth.token,
    onMatchFound: handleMatchFound,
    onMatchStart: handleMatchStart,
    onScoreUpdate: handleScoreUpdate,
    onTimeUpdate: handleTimeUpdate,
    onMatchComplete: handleMatchComplete,
    onBackendOffline: handleBackendOffline,
  });

  // ── Fetch leaderboard from Go backend (primary) ──────────────────────────
  useEffect(() => {
    async function loadLeaderboard() {
      try {
        const resp = await apiGetLeaderboard('general', 50, 0);
        if (resp.data && resp.data.length > 0) {
          const mapped: LeaderboardEntry[] = resp.data.map((e) => ({
            rank: e.rank,
            id: e.user_id,
            username: e.username,
            avatarUrl: e.avatar_url ||
              'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250',
            title: e.display_name ? `⚡ ${e.display_name}` : '⚡ Duelist',
            mmr: e.rating,
            wins: Math.round(e.matches_played * 0.6), // approximated until backend exposes wins
            totalMatches: e.matches_played,
            winRate: e.matches_played > 0 ? Math.round((e.matches_played * 0.6 / e.matches_played) * 100) : 0,
            avgResponseTimeMs: 2000,
            accuracyPercentage: 80,
            badge: e.rating >= 2000 ? '👑 Grandmaster' :
                   e.rating >= 1800 ? '🥈 Master' :
                   e.rating >= 1500 ? '🥉 Diamond' :
                   e.rating >= 1200 ? '⭐ Gold' : '⭐ Bronze',
          }));
          setLeaderboard(mapped);
        }
      } catch {
        // Backend offline — leaderboard stays at MOCK_LEADERBOARD
      }
    }
    loadLeaderboard();
  }, []);

  // ── Firebase Auth & Chat (unchanged) ────────────────────────────────────
  useEffect(() => {
    // 1. Firebase Auth & User Profile Sync
    const unsubscribeAuth = listenToAuthAndProfile((profile, user) => {
      if (user) {
        setUserProfile((prev) => ({
          ...profile,
          id: user.uid,
          username: user.displayName || profile.username || `Duelist_${user.uid.slice(0, 5)}`,
          avatarUrl: user.photoURL || profile.avatarUrl,
        }));
      } else {
        setUserProfile(profile);
      }
      setAuthUser(user);
    });

    // 2. Firebase Global Chat Real-time Listener
    const unsubscribeChat = listenToGlobalChat((messages) => {
      if (messages.length > 0) {
        setGlobalChat(messages);
      }
    });

    // Fallback local storage initialization
    const timer = setTimeout(() => {
      setAchievements(getStoredAchievements());
      setFriends(getStoredFriends());
      setNotifications(getStoredNotifications());
    }, 0);

    return () => {
      unsubscribeAuth();
      if (typeof unsubscribeChat === 'function') unsubscribeChat();
      clearTimeout(timer);
    };
  }, []);

  // Sync profile & state changes to Firestore & local storage
  useEffect(() => {
    saveStoredProfile(userProfile);
    if (userProfile && userProfile.id) {
      saveProfileToFirestore(userProfile);
    }
  }, [userProfile]);

  useEffect(() => {
    saveStoredAchievements(achievements);
  }, [achievements]);

  useEffect(() => {
    saveStoredFriends(friends);
  }, [friends]);

  useEffect(() => {
    saveStoredNotifications(notifications);
  }, [notifications]);

  useEffect(() => {
    saveStoredGlobalChat(globalChat);
  }, [globalChat]);

  // Subscribe to real-time BroadcastChannel for cross-window / multi-tab dual matches!
  useEffect(() => {
    const unsubscribe = arenaBroadcast.subscribe((event) => {
      const { type, payload } = event.data;

      if (type === 'ROOM_JOIN' && activeMatch && activeMatch.roomCode === payload.roomCode) {
        // Friend joined host's room!
        const updatedOpponent: MatchPlayer = {
          id: payload.player.id,
          username: payload.player.username,
          avatarUrl: payload.player.avatarUrl,
          title: payload.player.title,
          mmr: payload.player.mmr,
          score: 0,
          answers: [],
          isReady: true,
          isBot: false,
        };

        setActiveMatch((prev) => (prev ? { ...prev, player2: updatedOpponent } : null));

        // Send Notification
        const newNotif: AppNotification = {
          id: 'n_' + Date.now(),
          type: 'turn_update',
          title: 'Opponent Connected!',
          message: `${payload.player.username} entered your Dual Exam room!`,
          timestamp: Date.now(),
          read: false,
        };

        setNotifications((prev) => [newNotif, ...prev]);
        notificationManager.notifyTurnUpdate(payload.player.username, 1);
      }

      if (type === 'EMOTE_SENT' && activeMatch) {
        setActiveMatch((prev) =>
          prev ? { ...prev, emotes: [...prev.emotes, payload] } : null
        );
      }
    });

    return () => unsubscribe();
  }, [activeMatch]);

  // Launch a new Dual Match
  const handleStartMatch = async (config: {
    mode: 'quick' | 'bot' | 'private' | 'ai_custom';
    category: QuestionCategory | string;
    botDifficulty?: 'Beginner' | 'Intermediate' | 'Master' | 'Genius';
    roomCode?: string;
    customTopic?: string;
    opponentInfo?: { id: string; username: string; avatarUrl: string; mmr: number };
  }) => {
    let matchQuestions: Question[] = [];

    // Custom Topic or standard questions selection
    if (matchQuestions.length === 0) {
      const sample = SAMPLE_QUESTIONS[config.category as QuestionCategory] || SAMPLE_QUESTIONS['Computer Science'];
      matchQuestions = [...sample];
    }

    // ── Try backend ranked queue (quick match, not a bot) ──────────────────
    if (config.mode === 'quick' && !config.opponentInfo && backendAuth.token && !backendOffline) {
      try {
        // Map frontend category name to a backend category code
        const categoryCode = config.category.toString().toLowerCase().replace(/[^a-z0-9]/g, '_');
        await backendMatch.joinQueue(categoryCode);
        // Backend will emit match_found over WS — show a queuing indicator
        // The actual match state will be set by handleMatchFound callback
        // For now we also build a local match so the UI isn't blank
      } catch (e) {
        console.warn('[Backend] queue join failed, falling back to local sim:', e);
        setBackendOffline(true);
      }
    }

    const p1: MatchPlayer = {
      id: userProfile.id,
      username: userProfile.username,
      avatarUrl: userProfile.avatarUrl,
      title: userProfile.title,
      mmr: userProfile.mmr,
      score: 0,
      answers: [],
      isReady: true,
      isBot: false,
    };

    let p2: MatchPlayer;

    if (config.opponentInfo) {
      p2 = {
        id: config.opponentInfo.id,
        username: config.opponentInfo.username,
        avatarUrl: config.opponentInfo.avatarUrl,
        title: '⚡ Duelist',
        mmr: config.opponentInfo.mmr,
        score: 0,
        answers: [],
        isReady: true,
        isBot: false,
      };
    } else if (config.mode === 'bot') {
      const botNames = {
        Beginner: 'QuizBot_Novice',
        Intermediate: 'Professor_Byte',
        Master: 'Dr. Apex AI',
        Genius: 'Quantum_Omni_AI',
      };
      p2 = {
        id: 'bot_' + (config.botDifficulty || 'Master'),
        username: botNames[config.botDifficulty || 'Master'],
        avatarUrl: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&q=80&w=250',
        title: `🤖 ${config.botDifficulty || 'Master'} Bot`,
        mmr: 1300,
        score: 0,
        answers: [],
        isReady: true,
        isBot: true,
        botDifficulty: config.botDifficulty || 'Master',
      };
    } else {
      p2 = {
        id: 'matched_player_2',
        username: 'AuraSpeed',
        avatarUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=250',
        title: '⚡ Speed Scholar',
        mmr: 1280,
        score: 0,
        answers: [],
        isReady: true,
        isBot: false,
      };
    }

    const newMatch: MatchState = {
      id: 'match_' + Date.now(),
      roomCode: config.roomCode,
      category: config.customTopic || config.category,
      status: 'in_progress',
      player1: p1,
      player2: p2,
      questions: matchQuestions,
      currentQuestionIndex: 0,
      questionStartTime: Date.now(),
      timeLimitSeconds: 15,
      emotes: [],
      createdAt: Date.now(),
    };

    setActiveMatch(newMatch);
    setShowSummaryModal(false);

    // Broadcast room join for cross-tab multi-window play!
    if (config.roomCode) {
      arenaBroadcast.publish('ROOM_JOIN', {
        roomCode: config.roomCode,
        player: p1,
      });
    }
  };

  // Submit Answer & Calculate Score
  const handleAnswerSubmitted = (
    questionIndex: number,
    selectedOptionIndex: number,
    responseTimeMs: number
  ) => {
    if (!activeMatch) return;

    const currentQ = activeMatch.questions[questionIndex];
    const isCorrect = selectedOptionIndex === currentQ.correctAnswerIndex;

    // Send answer to Go backend over WebSocket if connected
    if (backendAuth.token && backendMatch.isConnected && activeMatch.id) {
      const optionId = currentQ.rawOptionIds ? currentQ.rawOptionIds[selectedOptionIndex] || '' : String(selectedOptionIndex);
      backendMatch.submitAnswer(activeMatch.id, currentQ.id, optionId, responseTimeMs);
    }

    // Scoring formula: Base 1000 + Speed Bonus up to 500
    const baseScore = isCorrect ? 1000 : 0;
    const speedBonus = isCorrect ? Math.max(0, Math.floor(500 * (1 - responseTimeMs / 15000))) : 0;
    const gainedScore = baseScore + speedBonus;

    const uAnswer: PlayerAnswer = {
      questionId: currentQ.id,
      selectedOptionIndex,
      responseTimeMs,
      isCorrect,
      scoreGained: gainedScore,
      speedBonus,
      streakBonus: 0,
    };

    // Calculate Opponent's simulated response
    let oAnswer: PlayerAnswer;

    if (activeMatch.player2.isBot) {
      const diff = activeMatch.player2.botDifficulty || 'Master';
      const accuracyProb = diff === 'Genius' ? 0.95 : diff === 'Master' ? 0.85 : diff === 'Intermediate' ? 0.70 : 0.50;
      const botCorrect = Math.random() < accuracyProb;

      const botDelay = diff === 'Genius' ? 1200 + Math.random() * 1000 : 2000 + Math.random() * 2500;
      const botGained = botCorrect ? Math.floor(1000 + (500 * (1 - botDelay / 15000))) : 0;

      oAnswer = {
        questionId: currentQ.id,
        selectedOptionIndex: botCorrect ? currentQ.correctAnswerIndex : (currentQ.correctAnswerIndex + 1) % 4,
        responseTimeMs: Math.round(botDelay),
        isCorrect: botCorrect,
        scoreGained: botGained,
        speedBonus: 0,
        streakBonus: 0,
      };
    } else {
      // Human opponent response simulation
      const oppCorrect = Math.random() < 0.8;
      const oppDelay = 1800 + Math.random() * 2000;
      oAnswer = {
        questionId: currentQ.id,
        selectedOptionIndex: oppCorrect ? currentQ.correctAnswerIndex : (currentQ.correctAnswerIndex + 1) % 4,
        responseTimeMs: Math.round(oppDelay),
        isCorrect: oppCorrect,
        scoreGained: oppCorrect ? Math.floor(1000 + (500 * (1 - oppDelay / 15000))) : 0,
        speedBonus: 0,
        streakBonus: 0,
      };
    }

    // Update active match state
    setActiveMatch((prev) => {
      if (!prev) return null;

      const p1Answers = [...prev.player1.answers, uAnswer];
      const p2Answers = [...prev.player2.answers, oAnswer];

      const p1Score = p1Answers.reduce((acc, a) => acc + a.scoreGained, 0);
      const p2Score = p2Answers.reduce((acc, a) => acc + a.scoreGained, 0);

      return {
        ...prev,
        player1: { ...prev.player1, answers: p1Answers, score: p1Score },
        player2: { ...prev.player2, answers: p2Answers, score: p2Score },
      };
    });
  };

  // Advance to Next Question
  const handleNextQuestion = () => {
    if (!activeMatch) return;
    setActiveMatch((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        currentQuestionIndex: prev.currentQuestionIndex + 1,
        questionStartTime: Date.now(),
      };
    });
  };

  // Complete Match & Calculate Career Progress
  const handleFinishMatch = () => {
    if (!activeMatch) return;

    const userScore = activeMatch.player1.score;
    const opponentScore = activeMatch.player2.score;

    const isWinner = userScore > opponentScore;
    const isDraw = userScore === opponentScore;

    const winnerId = isWinner ? activeMatch.player1.id : isDraw ? 'draw' : activeMatch.player2.id;

    // Update Profile Career Stats
    const totalMatches = userProfile.totalMatches + 1;
    const wins = isWinner ? userProfile.wins + 1 : userProfile.wins;
    const losses = !isWinner && !isDraw ? userProfile.losses + 1 : userProfile.losses;
    const mmrChange = isWinner ? 25 : isDraw ? 0 : -10;
    const newMmr = Math.max(800, userProfile.mmr + mmrChange);

    const userAnswers = activeMatch.player1.answers;
    const totalAnsCount = userAnswers.length;
    const totalCorrect = userAnswers.filter((a) => a.isCorrect).length;

    const matchAvgMs = userAnswers.reduce((acc, a) => acc + a.responseTimeMs, 0) / Math.max(1, totalAnsCount);

    const newAvgResponseTime = Math.round(
      (userProfile.avgResponseTimeMs * userProfile.totalMatches + matchAvgMs) / totalMatches
    );

    const newAccuracy = Math.round(
      ((userProfile.accuracyPercentage * (totalMatches - 1) + (totalCorrect / totalAnsCount) * 100) / totalMatches)
    );

    const currentStreak = isWinner ? userProfile.currentStreak + 1 : 0;
    const highestStreak = Math.max(userProfile.highestStreak, currentStreak);

    setUserProfile((prev) => ({
      ...prev,
      totalMatches,
      wins,
      losses,
      mmr: newMmr,
      avgResponseTimeMs: newAvgResponseTime,
      accuracyPercentage: newAccuracy,
      currentStreak,
      highestStreak,
      totalPoints: prev.totalPoints + userScore,
    }));

    // Unlock Achievements
    const updatedAchievements = achievements.map((ach) => {
      if (ach.id === 'ach_first_win' && isWinner) {
        return { ...ach, isUnlocked: true, progress: 1 };
      }
      if (ach.id === 'ach_speed_demon') {
        const hasFast = userAnswers.some((a) => a.isCorrect && a.responseTimeMs < 1500);
        if (hasFast) return { ...ach, isUnlocked: true, progress: 1 };
      }
      if (ach.id === 'ach_streak_3' && currentStreak >= 3) {
        return { ...ach, isUnlocked: true, progress: 3 };
      }
      return ach;
    });

    setAchievements(updatedAchievements);

    // Update active match status & open summary modal
    setActiveMatch((prev) => (prev ? { ...prev, status: 'completed', winnerId } : null));
    setShowSummaryModal(true);

    // Notify user
    notificationManager.notifyMatchResult(activeMatch.player2.username, isWinner, `${userScore} vs ${opponentScore}`);
  };

  // Send Emote Reaction
  const handleSendEmote = (text: string, icon: string) => {
    const emotePayload = {
      id: 'e_' + Date.now(),
      senderId: userProfile.id,
      senderName: userProfile.username,
      text,
      icon,
      timestamp: Date.now(),
    };

    if (activeMatch) {
      setActiveMatch((prev) => (prev ? { ...prev, emotes: [...prev.emotes, emotePayload] } : null));
      arenaBroadcast.publish('EMOTE_SENT', emotePayload);
    }
  };

  // Global Chat Handler
  const handleSendChatMessage = async (text: string) => {
    const newMsg: ChatMessage = {
      id: 'm_' + Date.now(),
      senderId: userProfile.id,
      senderName: userProfile.username,
      senderAvatar: userProfile.avatarUrl,
      senderTitle: userProfile.title,
      text,
      timestamp: Date.now(),
    };
    setGlobalChat((prev) => [...prev, newMsg]);
    try {
      await sendGlobalChatMessage({
        senderId: userProfile.id,
        senderName: userProfile.username,
        senderAvatar: userProfile.avatarUrl,
        senderTitle: userProfile.title,
        text,
        timestamp: Date.now(),
      });
    } catch (e) {
      console.error('Failed to send message to Firestore chat:', e);
    }
  };

  // Add Friend Handler
  const handleAddFriend = (username: string) => {
    const newFriend: Friend = {
      id: 'f_' + Date.now(),
      username,
      avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250',
      title: '⚡ Speed Scholar',
      mmr: 1200,
      status: 'online',
      winRate: 70,
    };
    setFriends((prev) => [...prev, newFriend]);
  };

  // Real-time Notifications Listener for logged-in user
  useEffect(() => {
    if (!userProfile?.id) return;
    const unsubscribeNotifs = listenToNotifications(userProfile.id, (liveNotifs) => {
      if (liveNotifs.length > 0) {
        setNotifications(liveNotifs);
      }
    });
    return () => {
      if (typeof unsubscribeNotifs === 'function') unsubscribeNotifs();
    };
  }, [userProfile?.id]);

  // ── Sync backendUser into userProfile when JWT login succeeds ─────────────
  useEffect(() => {
    if (backendAuth.backendUser) {
      const bu = backendAuth.backendUser;
      setUserProfile((prev) => ({
        ...prev,
        id: bu.id,
        username: bu.username || bu.display_name || prev.username,
        avatarUrl: bu.avatar_url ?? prev.avatarUrl,
        title: bu.display_name ? `⚡ ${bu.display_name}` : prev.title,
      }));
    }
  }, [backendAuth.backendUser]);

  // ── Loading splash ─────────────────────────────────────────────────────────
  if (backendAuth.loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-950">
        <div className="flex flex-col items-center gap-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-indigo-600 text-white font-black text-xl shadow-2xl shadow-indigo-600/40 animate-pulse">
            DX
          </div>
          <p className="text-sm font-medium text-slate-400">Loading Arena…</p>
        </div>
      </div>
    );
  }

  // ── Auth gate — show AuthPage when not authenticated ──────────────────────
  if (backendAuth.authMode === 'none') {
    return (
      <AuthPage
        onLoginWithPassword={backendAuth.loginWithPassword}
        onRegister={backendAuth.registerWithPassword}
        onLoginWithGoogle={backendAuth.loginWithGoogle}
        isLoading={backendAuth.loading}
        error={backendAuth.error}
        onClearError={backendAuth.clearError}
      />
    );
  }

  // ── Authenticated — render the full game app ───────────────────────────────
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans selection:bg-orange-500 selection:text-slate-950">
      {/* Top Navigation */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        profile={userProfile}
        notifications={notifications}
        onOpenNotifications={() => setIsNotificationCenterOpen(true)}
        soundEnabled={soundEnabled}
        setSoundEnabled={setSoundEnabled}
        onLogout={backendAuth.logout}
      />

      {/* Main View Area */}
      <main className="flex-1">
        {activeMatch && activeMatch.status === 'in_progress' ? (
          <DualExamArena
            match={activeMatch}
            onAnswerSubmitted={handleAnswerSubmitted}
            onNextQuestion={handleNextQuestion}
            onFinishMatch={handleFinishMatch}
            onSendEmote={handleSendEmote}
            onQuitMatch={() => {
              setActiveMatch(null);
              setShowSummaryModal(false);
            }}
          />
        ) : (
          <>
            {activeTab === 'lobby' && (
              <MatchmakingLobby
                profile={userProfile}
                friends={friends}
                onStartMatch={handleStartMatch}
                onChallengeFriend={(friend) =>
                  handleStartMatch({
                    mode: 'quick',
                    category: 'Computer Science',
                    opponentInfo: friend,
                  })
                }
              />
            )}

            {activeTab === 'leaderboard' && (
              <LeaderboardView leaderboard={leaderboard} userProfile={userProfile} />
            )}

            {activeTab === 'profile' && (
              <ProfileView
                profile={userProfile}
                achievements={achievements}
                onUpdateProfile={(updated) => setUserProfile((prev) => ({ ...prev, ...updated }))}
                authUser={authUser}
              />
            )}

            {activeTab === 'social' && (
              <FriendsAndChatView
                friends={friends}
                globalChat={globalChat}
                userProfile={userProfile}
                onSendChatMessage={handleSendChatMessage}
                onAddFriend={handleAddFriend}
                onChallengeFriend={(friend) =>
                  handleStartMatch({
                    mode: 'quick',
                    category: 'Computer Science',
                    opponentInfo: friend,
                  })
                }
              />
            )}
          </>
        )}
      </main>

      {/* Post Match Summary Modal */}
      {showSummaryModal && activeMatch && (
        <MatchSummaryModal
          match={activeMatch}
          userProfile={userProfile}
          onPlayAgain={() => {
            setShowSummaryModal(false);
            handleStartMatch({
              mode: 'quick',
              category: activeMatch.category,
            });
          }}
          onReturnToLobby={() => {
            setShowSummaryModal(false);
            setActiveMatch(null);
            setActiveTab('lobby');
          }}
        />
      )}

      {/* Notification Center Side Drawer */}
      <NotificationCenter
        isOpen={isNotificationCenterOpen}
        onClose={() => setIsNotificationCenterOpen(false)}
        notifications={notifications}
        onClearAll={() => setNotifications([])}
        onMarkAllRead={() => setNotifications((prev) => prev.map((n) => ({ ...n, read: true })))}
        onAcceptChallenge={(roomCode) => {
          handleStartMatch({
            mode: 'private',
            category: 'Computer Science',
            roomCode,
          });
        }}
      />
    </div>
  );
}
