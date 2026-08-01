'use client';

import React, { useState } from 'react';
import { 
  Trophy, 
  Zap, 
  Search, 
  Award, 
  Flame, 
  Target, 
  Crown, 
  Medal, 
  Users, 
  Globe 
} from 'lucide-react';
import { LeaderboardEntry, PlayerProfile } from '@/lib/types';
import { soundManager } from '@/lib/audio';

interface LeaderboardViewProps {
  leaderboard: LeaderboardEntry[];
  userProfile: PlayerProfile;
}

export const LeaderboardView: React.FC<LeaderboardViewProps> = ({
  leaderboard,
  userProfile,
}) => {
  const [filterMode, setFilterMode] = useState<'global' | 'speed' | 'accuracy'>('global');
  const [searchQuery, setSearchQuery] = useState('');

  // Sort entries according to active filter
  const sortedEntries = [...leaderboard].sort((a, b) => {
    if (filterMode === 'speed') {
      return a.avgResponseTimeMs - b.avgResponseTimeMs; // Lowest reaction time first
    }
    if (filterMode === 'accuracy') {
      return b.accuracyPercentage - a.accuracyPercentage;
    }
    return b.mmr - a.mmr; // Global MMR
  });

  const filteredEntries = sortedEntries.filter((e) =>
    e.username.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 space-y-8 animate-fadeIn">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full border border-indigo-500/30 bg-indigo-500/10 px-3 py-1 text-xs font-bold text-indigo-400">
            <Trophy className="h-3.5 w-3.5" />
            <span>GLOBAL HALL OF FAME</span>
          </div>
          <h1 className="text-2xl font-bold text-white sm:text-3xl tracking-tight mt-1">
            Dual Exam Leaderboards
          </h1>
          <p className="text-xs text-slate-400">
            Top participants ranked by MMR Rating, Reaction Speed, and Accuracy.
          </p>
        </div>

        {/* Filter buttons */}
        <div className="flex items-center gap-1.5 rounded-2xl border border-slate-800 bg-slate-900 p-1.5">
          <button
            onClick={() => {
              soundManager.playClick();
              setFilterMode('global');
            }}
            className={`flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs font-bold transition-all ${
              filterMode === 'global'
                ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/30'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            <Crown className="h-3.5 w-3.5" /> MMR Rank
          </button>

          <button
            onClick={() => {
              soundManager.playClick();
              setFilterMode('speed');
            }}
            className={`flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs font-bold transition-all ${
              filterMode === 'speed'
                ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/30'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            <Zap className="h-3.5 w-3.5" /> Speed Demons
          </button>

          <button
            onClick={() => {
              soundManager.playClick();
              setFilterMode('accuracy');
            }}
            className={`flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs font-bold transition-all ${
              filterMode === 'accuracy'
                ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/30'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            <Target className="h-3.5 w-3.5" /> Accuracy %
          </button>
        </div>
      </div>

      {/* Top 3 Champions Podium */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {sortedEntries.slice(0, 3).map((champion, idx) => {
          const podiumColors = [
            'border-amber-500/50 bg-gradient-to-b from-amber-500/10 via-slate-900 to-slate-900 text-amber-300',
            'border-slate-400/50 bg-gradient-to-b from-slate-400/10 via-slate-900 to-slate-900 text-slate-200',
            'border-orange-600/50 bg-gradient-to-b from-orange-600/10 via-slate-900 to-slate-900 text-orange-400',
          ];

          const badges = ['🥇 1st Place', '🥈 2nd Place', '🥉 3rd Place'];

          return (
            <div
              key={champion.id}
              className={`relative overflow-hidden rounded-3xl border p-6 text-center shadow-xl space-y-3 ${podiumColors[idx]}`}
            >
              <div className="inline-block rounded-full bg-slate-950 px-3 py-1 text-xs font-extrabold border border-current">
                {badges[idx]}
              </div>

              <div className="relative mx-auto h-20 w-20">
                <img
                  src={champion.avatarUrl}
                  alt={champion.username}
                  className="h-full w-full rounded-2xl object-cover ring-2 ring-current"
                />
              </div>

              <div>
                <h3 className="text-base font-extrabold text-white">{champion.username}</h3>
                <p className="text-xs text-slate-400">{champion.title}</p>
              </div>

              <div className="flex items-center justify-center gap-4 rounded-2xl bg-slate-950/80 p-3 text-xs">
                <div>
                  <span className="text-[10px] text-slate-400 block">MMR</span>
                  <strong className="text-amber-400 font-black">{champion.mmr}</strong>
                </div>
                <div className="h-6 w-px bg-slate-800" />
                <div>
                  <span className="text-[10px] text-slate-400 block">Avg Speed</span>
                  <strong className="text-cyan-400 font-black">{(champion.avgResponseTimeMs / 1000).toFixed(2)}s</strong>
                </div>
                <div className="h-6 w-px bg-slate-800" />
                <div>
                  <span className="text-[10px] text-slate-400 block">Accuracy</span>
                  <strong className="text-emerald-400 font-black">{champion.accuracyPercentage}%</strong>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Leaderboard Search & Table */}
      <div className="rounded-3xl border border-slate-800 bg-slate-900/90 p-6 shadow-2xl space-y-4">
        <div className="relative">
          <Search className="absolute left-4 top-3 h-4 w-4 text-slate-500" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search participant name..."
            className="w-full rounded-2xl border border-slate-800 bg-slate-950 pl-11 pr-4 py-2.5 text-sm text-white placeholder-slate-500 focus:border-amber-500 focus:outline-none"
          />
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="border-b border-slate-800 text-[11px] font-bold uppercase tracking-wider text-slate-400">
              <tr>
                <th className="pb-3 pl-2">Rank</th>
                <th className="pb-3">Participant</th>
                <th className="pb-3 text-center">MMR</th>
                <th className="pb-3 text-center">Wins / Matches</th>
                <th className="pb-3 text-center">Win Rate %</th>
                <th className="pb-3 text-center">Avg Speed</th>
                <th className="pb-3 text-center">Accuracy %</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {filteredEntries.map((entry, index) => {
                const isCurrentUser = entry.id === userProfile.id;

                return (
                  <tr
                    key={entry.id}
                    className={`transition-colors ${
                      isCurrentUser
                        ? 'bg-amber-500/10 font-bold text-white ring-1 ring-amber-500/30'
                        : 'hover:bg-slate-800/40 text-slate-300'
                    }`}
                  >
                    <td className="py-3.5 pl-2 font-black text-slate-400">
                      #{index + 1}
                    </td>
                    <td className="py-3.5">
                      <div className="flex items-center gap-3">
                        <img
                          src={entry.avatarUrl}
                          alt={entry.username}
                          className="h-8 w-8 rounded-xl object-cover"
                        />
                        <div>
                          <p className="font-bold text-white text-xs">{entry.username}</p>
                          <p className="text-[10px] text-slate-400">{entry.title}</p>
                        </div>
                      </div>
                    </td>
                    <td className="py-3.5 text-center font-extrabold text-amber-400">
                      {entry.mmr}
                    </td>
                    <td className="py-3.5 text-center text-slate-300">
                      {entry.wins} / {entry.totalMatches}
                    </td>
                    <td className="py-3.5 text-center font-bold text-cyan-400">
                      {entry.winRate}%
                    </td>
                    <td className="py-3.5 text-center font-bold text-orange-400">
                      {(entry.avgResponseTimeMs / 1000).toFixed(2)}s
                    </td>
                    <td className="py-3.5 text-center font-bold text-emerald-400">
                      {entry.accuracyPercentage}%
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
