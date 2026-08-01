'use client';

import React from 'react';
import { 
  Zap, 
  Trophy, 
  User, 
  Users, 
  Bell, 
  Volume2, 
  VolumeX, 
  Flame, 
  Swords, 
  LogOut
} from 'lucide-react';
import { PlayerProfile, AppNotification } from '@/lib/types';
import { soundManager } from '@/lib/audio';

interface NavbarProps {
  activeTab: 'lobby' | 'leaderboard' | 'profile' | 'social';
  setActiveTab: (tab: 'lobby' | 'leaderboard' | 'profile' | 'social') => void;
  profile: PlayerProfile;
  notifications: AppNotification[];
  onOpenNotifications: () => void;
  soundEnabled: boolean;
  setSoundEnabled: (enabled: boolean) => void;
  /** Optional logout handler from auth system */
  onLogout?: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  setActiveTab,
  profile,
  notifications,
  onOpenNotifications,
  soundEnabled,
  setSoundEnabled,
  onLogout,
}) => {
  const unreadCount = notifications.filter((n) => !n.read).length;

  const handleTabClick = (tab: 'lobby' | 'leaderboard' | 'profile' | 'social') => {
    soundManager.playClick();
    setActiveTab(tab);
  };

  const toggleAudio = () => {
    const newState = !soundEnabled;
    soundManager.toggleSound(newState);
    setSoundEnabled(newState);
  };

  return (
    <header className="sticky top-0 z-40 w-full border-b border-slate-800 bg-slate-950/80 backdrop-blur-md">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6">
        {/* Brand Logo */}
        <div 
          onClick={() => handleTabClick('lobby')}
          className="flex cursor-pointer items-center gap-3 transition-transform hover:scale-105"
        >
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-600 font-bold text-white text-sm shadow-md shadow-indigo-600/30">
            DX
          </div>
          <div>
            <div className="flex items-center gap-1.5">
              <span className="text-lg font-bold tracking-tight text-white">DUAL EXAM</span>
              <span className="text-lg font-bold text-indigo-400">PRO</span>
            </div>
            <p className="hidden text-[10px] font-medium text-slate-400 sm:block tracking-wide uppercase">
              1v1 Speed & Accuracy Battle
            </p>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav className="flex items-center gap-1 rounded-2xl border border-slate-800 bg-slate-900/80 p-1">
          <button
            id="nav-tab-lobby"
            onClick={() => handleTabClick('lobby')}
            className={`flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs sm:text-sm font-medium transition-all ${
              activeTab === 'lobby'
                ? 'bg-indigo-600 text-white font-bold shadow-md shadow-indigo-600/30'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            <Swords className="h-4 w-4" />
            <span className="hidden sm:inline">Arena Lobby</span>
          </button>

          <button
            id="nav-tab-leaderboard"
            onClick={() => handleTabClick('leaderboard')}
            className={`flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs sm:text-sm font-medium transition-all ${
              activeTab === 'leaderboard'
                ? 'bg-indigo-600 text-white font-bold shadow-md shadow-indigo-600/30'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            <Trophy className="h-4 w-4" />
            <span className="hidden sm:inline">Ranks</span>
          </button>

          <button
            id="nav-tab-social"
            onClick={() => handleTabClick('social')}
            className={`flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs sm:text-sm font-medium transition-all ${
              activeTab === 'social'
                ? 'bg-indigo-600 text-white font-bold shadow-md shadow-indigo-600/30'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            <Users className="h-4 w-4" />
            <span className="hidden sm:inline">Social</span>
          </button>

          <button
            id="nav-tab-profile"
            onClick={() => handleTabClick('profile')}
            className={`flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs sm:text-sm font-medium transition-all ${
              activeTab === 'profile'
                ? 'bg-indigo-600 text-white font-bold shadow-md shadow-indigo-600/30'
                : 'text-slate-400 hover:text-white'
            }`}
          >
            <User className="h-4 w-4" />
            <span className="hidden sm:inline">Profile</span>
          </button>
        </nav>

        {/* Right Status Actions & Controls */}
        <div className="flex items-center gap-3">
          {/* Online count badge */}
          <div className="hidden lg:flex items-center text-xs text-emerald-400 font-medium">
            <span className="w-2 h-2 bg-emerald-500 rounded-full mr-2 animate-pulse" /> 1,248 Online
          </div>

          {/* User Streak & MMR pill */}
          <div className="hidden items-center gap-2 rounded-xl border border-slate-800 bg-slate-900 px-2.5 py-1 md:flex">
            <div className="flex items-center gap-1 text-indigo-400">
              <Zap className="h-3.5 w-3.5 fill-indigo-400" />
              <span className="text-xs font-bold font-mono">{profile.mmr} MMR</span>
            </div>
            <div className="h-3 w-px bg-slate-800" />
            <div className="flex items-center gap-1 text-amber-400">
              <Flame className="h-3.5 w-3.5 fill-amber-400" />
              <span className="text-xs font-bold">{profile.currentStreak}x</span>
            </div>
          </div>

          {/* Sound Toggle */}
          <button
            id="btn-sound-toggle"
            onClick={toggleAudio}
            title={soundEnabled ? 'Mute Sound Effects' : 'Enable Sound Effects'}
            className="flex h-9 w-9 items-center justify-center rounded-xl border border-slate-800 bg-slate-900 text-slate-300 transition-colors hover:bg-slate-800 hover:text-white"
          >
            {soundEnabled ? (
              <Volume2 className="h-4 w-4 text-emerald-400" />
            ) : (
              <VolumeX className="h-4 w-4 text-rose-400" />
            )}
          </button>

          {/* Notifications Button */}
          <button
            id="btn-notification-bell"
            onClick={() => {
              soundManager.playClick();
              onOpenNotifications();
            }}
            className="relative flex h-9 w-9 items-center justify-center rounded-xl border border-slate-800 bg-slate-900 text-slate-300 transition-colors hover:bg-slate-800 hover:text-white"
          >
            <Bell className="h-4 w-4" />
            {unreadCount > 0 && (
              <span className="absolute -top-1 -right-1 flex h-4 min-w-[16px] items-center justify-center rounded-full bg-rose-500 px-1 text-[10px] font-bold text-white shadow-md animate-pulse">
                {unreadCount}
              </span>
            )}
          </button>

          {/* Logout button — only shown when auth handler is provided */}
          {onLogout && (
            <button
              id="btn-logout"
              onClick={onLogout}
              title="Sign Out"
              className="flex h-9 w-9 items-center justify-center rounded-xl border border-slate-800 bg-slate-900 text-slate-400 transition-colors hover:bg-rose-500/20 hover:text-rose-400 hover:border-rose-500/40"
            >
              <LogOut className="h-4 w-4" />
            </button>
          )}

          {/* User Avatar */}
          <div 
            onClick={() => handleTabClick('profile')}
            className="relative h-9 w-9 cursor-pointer overflow-hidden rounded-xl border border-amber-500/40 ring-2 ring-amber-500/20 transition-transform hover:scale-105"
          >
            <img 
              src={profile.avatarUrl} 
              alt={profile.username} 
              className="h-full w-full object-cover"
            />
          </div>
        </div>
      </div>
    </header>
  );
};

