'use client';

import React, { useState } from 'react';
import { 
  Swords, 
  Bot, 
  Key, 
  Sparkles, 
  Zap, 
  Users, 
  Globe, 
  Brain, 
  ShieldCheck, 
  Loader2, 
  Send,
  Flame,
  CheckCircle2,
  RefreshCw
} from 'lucide-react';
import { QuestionCategory, PlayerProfile, Friend } from '@/lib/types';
import { soundManager } from '@/lib/audio';

interface MatchmakingLobbyProps {
  profile: PlayerProfile;
  friends: Friend[];
  onStartMatch: (config: {
    mode: 'quick' | 'bot' | 'private' | 'ai_custom';
    category: QuestionCategory | string;
    botDifficulty?: 'Beginner' | 'Intermediate' | 'Master' | 'Genius';
    roomCode?: string;
    customTopic?: string;
    opponentInfo?: { id: string; username: string; avatarUrl: string; mmr: number };
  }) => void;
  onChallengeFriend: (friend: Friend) => void;
}

const CATEGORIES: { id: QuestionCategory; name: string; icon: string; desc: string; isNew?: boolean }[] = [
  { id: 'Mathematics', name: 'Mathematics', icon: '📐', desc: 'Algebra, Calculus, Geometry & Step Breakdown', isNew: true },
  { id: 'English Language', name: 'English Language', icon: '📖', desc: 'Grammar, Vocabulary, Syntax & Comprehension', isNew: true },
  { id: 'Logical Reasoning', name: 'Logical Reasoning', icon: '🧠', desc: 'Syllogisms, Patterns, Series & Deductive Logic', isNew: true },
  { id: 'Computer Science', name: 'Computer Science', icon: '💻', desc: 'Algorithms, Data Structures & Systems' },
  { id: 'Science & Physics', name: 'Science & Physics', icon: '🔬', desc: 'Quantum Mechanics, Energy & Nature' },
  { id: 'Mathematics & Logic', name: 'Mathematics & Logic', icon: '🧮', desc: 'Fast Mental Math & Mixed Quantitative' },
  { id: 'General Knowledge', name: 'General Knowledge', icon: '🌐', desc: 'Trivia, Pop Culture & Global Facts' },
  { id: 'World History', name: 'World History', icon: '📜', desc: 'Ancient Empires, Revolutions & Timelines' },
  { id: 'Medical & Biology', name: 'Medical & Biology', icon: '🩺', desc: 'Anatomy, Genetics & Medical Science' },
  { id: 'GRE / SAT Prep', name: 'GRE / SAT Prep', icon: '🎓', desc: 'Vocab, Logic Reasoning & SAT Quantitative' },
];

const MOCK_ONLINE_PLAYERS = [
  { id: 'op1', username: 'HyperMind_99', avatarUrl: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=250', mmr: 1420, status: 'online' as const, wins: 42 },
  { id: 'op2', username: 'AuraSpeed', avatarUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=250', mmr: 1380, status: 'online' as const, wins: 38 },
  { id: 'op3', username: 'CalculatedGenius', avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250', mmr: 1560, status: 'in_match' as const, wins: 89 },
  { id: 'op4', username: 'Zenith_Byte', avatarUrl: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&q=80&w=250', mmr: 1290, status: 'online' as const, wins: 22 },
];

export const MatchmakingLobby: React.FC<MatchmakingLobbyProps> = ({
  profile,
  friends,
  onStartMatch,
  onChallengeFriend,
}) => {
  const [selectedCategory, setSelectedCategory] = useState<QuestionCategory | string>('Computer Science');
  const [selectedMode, setSelectedMode] = useState<'quick' | 'bot' | 'private' | 'ai_custom'>('quick');
  const [botDifficulty, setBotDifficulty] = useState<'Beginner' | 'Intermediate' | 'Master' | 'Genius'>('Master');
  const [roomCodeInput, setRoomCodeInput] = useState('');
  const [customTopicInput, setCustomTopicInput] = useState('');
  const [isSearchingMatch, setIsSearchingMatch] = useState(false);
  const [searchTimerSeconds, setSearchTimerSeconds] = useState(0);

  const [generatedPrivateCode, setGeneratedPrivateCode] = useState<string>(() => {
    return 'DUAL-' + Math.floor(1000 + Math.random() * 9000);
  });

  const handleLaunchMatch = () => {
    soundManager.playMatchStart();

    if (selectedMode === 'quick') {
      setIsSearchingMatch(true);
      setSearchTimerSeconds(0);

      // Simulate matchmaking search
      const interval = setInterval(() => {
        setSearchTimerSeconds((prev) => prev + 1);
      }, 1000);

      setTimeout(() => {
        clearInterval(interval);
        setIsSearchingMatch(false);
        onStartMatch({
          mode: 'quick',
          category: selectedCategory,
        });
      }, 2500);
      return;
    }

    if (selectedMode === 'private') {
      const code = roomCodeInput.trim().toUpperCase() || generatedPrivateCode;
      onStartMatch({
        mode: 'private',
        category: selectedCategory,
        roomCode: code,
      });
      return;
    }

    if (selectedMode === 'ai_custom') {
      if (!customTopicInput.trim()) return;
      onStartMatch({
        mode: 'ai_custom',
        category: 'Custom AI Topic',
        customTopic: customTopicInput.trim(),
      });
      return;
    }

    onStartMatch({
      mode: 'bot',
      category: selectedCategory,
      botDifficulty,
    });
  };

  return (
    <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 space-y-8 animate-fadeIn">
      {/* Hero Banner Header */}
      <div className="relative overflow-hidden rounded-3xl border border-slate-800 bg-slate-900 p-6 sm:p-8 shadow-2xl">
        <div className="relative z-10 flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
          <div className="space-y-2">
            <div className="inline-flex items-center gap-2 rounded-full border border-indigo-500/30 bg-indigo-500/10 px-3 py-1 text-xs font-bold text-indigo-400">
              <Zap className="h-3.5 w-3.5 fill-indigo-400" />
              <span>REAL-TIME DUAL EXAM ARENA</span>
            </div>
            <h1 className="text-2xl font-bold text-white sm:text-4xl tracking-tight">
              Test Speed & Accuracy in <span className="text-indigo-400">1v1 Real-Time Exam Duels</span>
            </h1>
            <p className="max-w-2xl text-sm text-slate-300 leading-relaxed">
              Two players receive identical exam questions simultaneously. The faster and more accurate answer wins! Score higher, earn MMR rating, and climb the global leaderboards.
            </p>
          </div>

          {/* Player Banner Card */}
          <div className="flex items-center gap-4 rounded-2xl border border-slate-800 bg-slate-950 p-4 shrink-0 shadow-lg ring-1 ring-slate-800">
            <img 
              src={profile.avatarUrl} 
              alt={profile.username}
              className="h-14 w-14 rounded-2xl object-cover ring-2 ring-indigo-500/40"
            />
            <div>
              <div className="flex items-center gap-2">
                <span className="text-base font-bold text-white">{profile.username}</span>
                <span className="rounded bg-indigo-500/20 px-2 py-0.5 text-[10px] font-bold text-indigo-300 border border-indigo-500/30">
                  {profile.title}
                </span>
              </div>
              <div className="mt-1 flex items-center gap-3 text-xs text-slate-400">
                <span className="font-semibold text-emerald-400">🏆 {profile.wins} Wins</span>
                <span>•</span>
                <span className="font-semibold text-indigo-400">⚡ {profile.avgResponseTimeMs / 1000}s Avg</span>
                <span>•</span>
                <span className="font-semibold text-cyan-400">🎯 {profile.accuracyPercentage}%</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Grid: Match Setup (Left) + Online Players & Friends (Right) */}
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-12">
        {/* Left Column: Mode Selector & Category Selector */}
        <div className="space-y-6 lg:col-span-8">
          {/* Step 1: Mode Selector */}
          <div>
            <h2 className="text-sm font-bold text-slate-400 uppercase tracking-wider mb-3 flex items-center gap-2">
              <span>1. Choose Game Mode</span>
            </h2>

            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {/* Quick Match */}
              <div
                onClick={() => {
                  soundManager.playClick();
                  setSelectedMode('quick');
                }}
                className={`cursor-pointer rounded-2xl border p-4 transition-all ${
                  selectedMode === 'quick'
                    ? 'border-indigo-500 bg-indigo-950/20 ring-2 ring-indigo-500/40 shadow-xl'
                    : 'border-slate-800 bg-slate-900/60 hover:border-slate-700 hover:bg-slate-900'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-500/20 text-indigo-400">
                    <Zap className="h-5 w-5" />
                  </div>
                  {selectedMode === 'quick' && <CheckCircle2 className="h-5 w-5 text-indigo-400" />}
                </div>
                <h3 className="mt-3 text-base font-bold text-white">Quick Matchmaking</h3>
                <p className="mt-1 text-xs text-slate-400">
                  Instant 1v1 match with an online opponent or smart AI bot.
                </p>
              </div>

              {/* Play against AI Bot */}
              <div
                onClick={() => {
                  soundManager.playClick();
                  setSelectedMode('bot');
                }}
                className={`cursor-pointer rounded-2xl border p-4 transition-all ${
                  selectedMode === 'bot'
                    ? 'border-indigo-500 bg-indigo-950/20 ring-2 ring-indigo-500/40 shadow-xl'
                    : 'border-slate-800 bg-slate-900/60 hover:border-slate-700 hover:bg-slate-900'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-purple-500/20 text-purple-400">
                    <Bot className="h-5 w-5" />
                  </div>
                  {selectedMode === 'bot' && <CheckCircle2 className="h-5 w-5 text-indigo-400" />}
                </div>
                <h3 className="mt-3 text-base font-bold text-white">AI Expert Duel</h3>
                <p className="mt-1 text-xs text-slate-400">
                  Practice against simulated AI players with custom difficulty.
                </p>
              </div>

              {/* Private Room Code */}
              <div
                onClick={() => {
                  soundManager.playClick();
                  setSelectedMode('private');
                }}
                className={`cursor-pointer rounded-2xl border p-4 transition-all ${
                  selectedMode === 'private'
                    ? 'border-indigo-500 bg-indigo-950/20 ring-2 ring-indigo-500/40 shadow-xl'
                    : 'border-slate-800 bg-slate-900/60 hover:border-slate-700 hover:bg-slate-900'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/20 text-emerald-400">
                    <Key className="h-5 w-5" />
                  </div>
                  {selectedMode === 'private' && <CheckCircle2 className="h-5 w-5 text-indigo-400" />}
                </div>
                <h3 className="mt-3 text-base font-bold text-white">Private Dual Code</h3>
                <p className="mt-1 text-xs text-slate-400">
                  Host or join a custom room. Works cross-tab in real time!
                </p>
              </div>

              {/* Custom AI Topic Exam */}
              <div
                onClick={() => {
                  soundManager.playClick();
                  setSelectedMode('ai_custom');
                }}
                className={`cursor-pointer rounded-2xl border p-4 transition-all ${
                  selectedMode === 'ai_custom'
                    ? 'border-indigo-500 bg-indigo-950/20 ring-2 ring-indigo-500/40 shadow-xl'
                    : 'border-slate-800 bg-slate-900/60 hover:border-slate-700 hover:bg-slate-900'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-cyan-500/20 text-cyan-400">
                    <Brain className="h-5 w-5" />
                  </div>
                  {selectedMode === 'ai_custom' && <CheckCircle2 className="h-5 w-5 text-indigo-400" />}
                </div>
                <h3 className="mt-3 text-base font-bold text-white">Custom AI Exam</h3>
                <p className="mt-1 text-xs text-slate-400">
                  Generate dynamic exam questions on ANY topic via Gemini AI.
                </p>
              </div>
            </div>
          </div>

          {/* Mode Configuration Sub-Panels */}
          {selectedMode === 'bot' && (
            <div className="rounded-2xl border border-purple-500/30 bg-purple-950/10 p-4 space-y-3">
              <label className="text-xs font-bold text-purple-300 uppercase tracking-wider block">
                Select AI Bot Difficulty
              </label>
              <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                {(['Beginner', 'Intermediate', 'Master', 'Genius'] as const).map((diff) => (
                  <button
                    key={diff}
                    onClick={() => {
                      soundManager.playClick();
                      setBotDifficulty(diff);
                    }}
                    className={`rounded-xl border py-2 text-xs font-bold transition-all ${
                      botDifficulty === diff
                        ? 'border-purple-500 bg-purple-600 text-white shadow-md'
                        : 'border-slate-800 bg-slate-900 text-slate-400 hover:text-white'
                    }`}
                  >
                    {diff}
                  </button>
                ))}
              </div>
            </div>
          )}

          {selectedMode === 'private' && (
            <div className="rounded-2xl border border-emerald-500/30 bg-emerald-950/10 p-4 space-y-3">
              <label className="text-xs font-bold text-emerald-300 uppercase tracking-wider block">
                Private Room Code
              </label>
              <div className="flex flex-col gap-2 sm:flex-row">
                <input
                  type="text"
                  value={roomCodeInput}
                  onChange={(e) => setRoomCodeInput(e.target.value)}
                  placeholder={`Enter code (or use host code: ${generatedPrivateCode})`}
                  className="flex-1 rounded-xl border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white focus:border-emerald-500 focus:outline-none"
                />
                <button
                  onClick={() => {
                    setRoomCodeInput(generatedPrivateCode);
                    navigator.clipboard?.writeText(generatedPrivateCode);
                  }}
                  className="flex items-center gap-1.5 rounded-xl border border-emerald-500/40 bg-emerald-500/20 px-3 py-2 text-xs font-bold text-emerald-300 hover:bg-emerald-500/30"
                >
                  <RefreshCw className="h-3.5 w-3.5" /> Copy Host Code ({generatedPrivateCode})
                </button>
              </div>
            </div>
          )}

          {selectedMode === 'ai_custom' && (
            <div className="rounded-2xl border border-cyan-500/30 bg-cyan-950/10 p-4 space-y-3">
              <label className="text-xs font-bold text-cyan-300 uppercase tracking-wider block">
                Type Custom Topic for Gemini AI Exam
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={customTopicInput}
                  onChange={(e) => setCustomTopicInput(e.target.value)}
                  placeholder="e.g. Astrophysics, Organic Chemistry, World War I, Python Programming"
                  className="flex-1 rounded-xl border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-white focus:border-cyan-500 focus:outline-none"
                />
              </div>
            </div>
          )}

          {/* Step 2: Subject Category Selection */}
          {selectedMode !== 'ai_custom' && (
            <div>
              <h2 className="text-sm font-bold text-slate-400 uppercase tracking-wider mb-3">
                2. Select Exam Category
              </h2>

              <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
                {CATEGORIES.map((cat) => (
                  <div
                    key={cat.id}
                    onClick={() => {
                      soundManager.playClick();
                      setSelectedCategory(cat.id);
                    }}
                    className={`flex cursor-pointer items-center gap-3 rounded-2xl border p-3.5 transition-all ${
                      selectedCategory === cat.id
                        ? 'border-indigo-500 bg-slate-900 ring-2 ring-indigo-500/30 text-white shadow-lg'
                        : 'border-slate-800 bg-slate-900/60 text-slate-300 hover:border-slate-700 hover:bg-slate-900'
                    }`}
                  >
                    <span className="text-2xl">{cat.icon}</span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-1">
                        <div className="flex items-center gap-1.5 min-w-0">
                          <span className="text-sm font-bold text-white truncate">{cat.name}</span>
                          {cat.isNew && (
                            <span className="rounded-md bg-indigo-500/20 px-1.5 py-0.5 text-[9px] font-extrabold text-indigo-300 border border-indigo-500/30 shrink-0">
                              STEP ANALYSIS
                            </span>
                          )}
                        </div>
                        {selectedCategory === cat.id && (
                          <span className="h-2 w-2 rounded-full bg-indigo-400 animate-pulse shrink-0" />
                        )}
                      </div>
                      <p className="text-[11px] text-slate-400 truncate">{cat.desc}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Big Launch Dual Exam Button */}
          <div className="pt-2">
            <button
              id="btn-launch-dual-exam"
              disabled={isSearchingMatch}
              onClick={handleLaunchMatch}
              className="relative w-full overflow-hidden rounded-2xl bg-indigo-600 p-4 text-center text-base font-bold tracking-wider text-white shadow-xl shadow-indigo-600/30 transition-all hover:bg-indigo-500 hover:scale-[1.01] active:scale-[0.99] disabled:opacity-50"
            >
              {isSearchingMatch ? (
                <div className="flex items-center justify-center gap-2">
                  <Loader2 className="h-5 w-5 animate-spin text-white" />
                  <span>MATCHMAKING IN PROGRESS... ({searchTimerSeconds}s)</span>
                </div>
              ) : (
                <div className="flex items-center justify-center gap-2">
                  <Swords className="h-5 w-5" />
                  <span>START 1v1 DUAL EXAM ARENA</span>
                </div>
              )}
            </button>
          </div>
        </div>

        {/* Right Column: Online Players & Friend Challenges */}
        <div className="space-y-6 lg:col-span-4">
          {/* Online Players Hub */}
          <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-5 space-y-4 shadow-xl">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4 text-emerald-400" />
                <h3 className="text-sm font-bold text-white">Online Duelists Hub</h3>
              </div>
              <span className="rounded-full bg-emerald-500/20 px-2 py-0.5 text-[10px] font-bold text-emerald-400 border border-emerald-500/30">
                ● 4 Active
              </span>
            </div>

            <div className="space-y-3">
              {MOCK_ONLINE_PLAYERS.map((p) => (
                <div
                  key={p.id}
                  className="flex items-center justify-between rounded-2xl border border-slate-800 bg-slate-950/60 p-3 hover:border-slate-700 transition-colors"
                >
                  <div className="flex items-center gap-2.5 min-w-0">
                    <div className="relative">
                      <img
                        src={p.avatarUrl}
                        alt={p.username}
                        className="h-9 w-9 rounded-xl object-cover"
                      />
                      <span className="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border-2 border-slate-950 bg-emerald-400" />
                    </div>
                    <div className="min-w-0">
                      <p className="text-xs font-bold text-white truncate">{p.username}</p>
                      <p className="text-[10px] text-slate-400">{p.mmr} MMR • {p.wins} Wins</p>
                    </div>
                  </div>

                  <button
                    onClick={() => {
                      soundManager.playClick();
                      onStartMatch({
                        mode: 'quick',
                        category: selectedCategory,
                        opponentInfo: p,
                      });
                    }}
                    className="flex items-center gap-1 rounded-xl bg-indigo-600 px-2.5 py-1.5 text-[11px] font-bold text-white hover:bg-indigo-500 transition-colors shadow-sm"
                  >
                    <Zap className="h-3 w-3" /> Challenge
                  </button>
                </div>
              ))}
            </div>
          </div>

          {/* Quick Friends Challenge */}
          <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-5 space-y-4 shadow-xl">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4 text-cyan-400" />
                <h3 className="text-sm font-bold text-white">Friend List</h3>
              </div>
              <span className="text-[10px] text-slate-400">{friends.length} Friends</span>
            </div>

            <div className="space-y-2.5">
              {friends.slice(0, 3).map((f) => (
                <div
                  key={f.id}
                  className="flex items-center justify-between rounded-2xl border border-slate-800/80 bg-slate-950/40 p-2.5"
                >
                  <div className="flex items-center gap-2 min-w-0">
                    <img
                      src={f.avatarUrl}
                      alt={f.username}
                      className="h-8 w-8 rounded-lg object-cover"
                    />
                    <div className="min-w-0">
                      <p className="text-xs font-semibold text-white truncate">{f.username}</p>
                      <span className="text-[9px] text-slate-400">{f.status}</span>
                    </div>
                  </div>

                  <button
                    onClick={() => {
                      soundManager.playClick();
                      onChallengeFriend(f);
                    }}
                    className="rounded-lg bg-slate-800 px-2.5 py-1 text-[11px] font-semibold text-cyan-300 hover:bg-slate-700 hover:text-white"
                  >
                    Invite
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
