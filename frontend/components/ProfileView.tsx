'use client';

import React, { useState, useEffect } from 'react';
import { 
  User as UserIcon, 
  Award, 
  Flame, 
  Zap, 
  Target, 
  Trophy, 
  ShieldCheck, 
  CheckCircle2, 
  Lock, 
  Edit3, 
  Sparkles, 
  Save,
  LogIn,
  LogOut,
  Globe
} from 'lucide-react';
import { PlayerProfile, Achievement } from '@/lib/types';
import { soundManager } from '@/lib/audio';
import { signInWithGoogle, signInAnonymouslyUser, logOutUser } from '@/lib/firestoreService';

interface ProfileViewProps {
  profile: PlayerProfile;
  achievements: Achievement[];
  onUpdateProfile: (updated: Partial<PlayerProfile>) => void;
  authUser?: any;
}

const AVATAR_OPTIONS = [
  'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250',
  'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=250',
  'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=250',
  'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&q=80&w=250',
  'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&q=80&w=250',
];

const TITLE_OPTIONS = [
  '⚡ Speed Scholar',
  '👑 Grandmaster Scholar',
  '⚡ Lightning Reflexes',
  '🧮 Logic Master',
  '🧠 Neural Titan',
  '🎯 Precision Duelist',
];

export const ProfileView: React.FC<ProfileViewProps> = ({
  profile,
  achievements,
  onUpdateProfile,
  authUser,
}) => {
  const [isEditing, setIsEditing] = useState(false);
  const [usernameInput, setUsernameInput] = useState(profile.username);
  const [titleInput, setTitleInput] = useState(profile.title);
  const [avatarInput, setAvatarInput] = useState(profile.avatarUrl);
  const [bioInput, setBioInput] = useState(profile.bio || '');
  const [authError, setAuthError] = useState('');

  useEffect(() => {
    setUsernameInput(profile.username);
    setTitleInput(profile.title);
    setAvatarInput(profile.avatarUrl);
    setBioInput(profile.bio || '');
  }, [profile]);

  const handleGoogleLogin = async () => {
    soundManager.playClick();
    setAuthError('');
    try {
      await signInWithGoogle();
    } catch (e: any) {
      setAuthError(e.message || 'Google Sign-In failed');
    }
  };

  const handleAnonLogin = async () => {
    soundManager.playClick();
    setAuthError('');
    try {
      await signInAnonymouslyUser();
    } catch (e: any) {
      setAuthError(e.message || 'Anonymous Sign-In failed');
    }
  };

  const handleLogout = async () => {
    soundManager.playClick();
    try {
      await logOutUser();
    } catch (e: any) {
      console.error(e);
    }
  };

  const handleSave = () => {
    soundManager.playClick();
    onUpdateProfile({
      username: usernameInput,
      title: titleInput,
      avatarUrl: avatarInput,
      bio: bioInput,
    });
    setIsEditing(false);
  };

  const unlockedCount = achievements.filter((a) => a.isUnlocked).length;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 space-y-8 animate-fadeIn">
      {/* Firebase Cloud Sync Banner */}
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-indigo-500/30 bg-indigo-950/20 px-5 py-3 text-xs text-indigo-200">
        <div className="flex items-center gap-2">
          <Globe className="h-4 w-4 text-indigo-400" />
          <span><strong>Firebase Cloud Sync Active:</strong> Profiles, MMR ratings, 1v1 match rooms & global chat are persisted live.</span>
        </div>

        <div className="flex items-center gap-2">
          {authUser ? (
            <div className="flex items-center gap-2">
              <span className="text-emerald-400 font-bold flex items-center gap-1">
                <CheckCircle2 className="h-3.5 w-3.5" /> Authenticated ({authUser.isAnonymous ? 'Anonymous Duelist' : authUser.email || authUser.displayName})
              </span>
              <button
                onClick={handleLogout}
                className="flex items-center gap-1 rounded-xl border border-rose-500/30 bg-rose-500/10 px-3 py-1 text-rose-300 font-bold hover:bg-rose-500/20"
              >
                <LogOut className="h-3 w-3" /> Sign Out
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <button
                onClick={handleGoogleLogin}
                className="flex items-center gap-1.5 rounded-xl bg-indigo-600 px-3 py-1.5 text-xs font-bold text-white shadow-md hover:bg-indigo-500"
              >
                <LogIn className="h-3.5 w-3.5" /> Sign in with Google
              </button>
              <button
                onClick={handleAnonLogin}
                className="flex items-center gap-1.5 rounded-xl border border-slate-700 bg-slate-900 px-3 py-1.5 text-xs font-bold text-slate-300 hover:text-white"
              >
                Quick Guest Auth
              </button>
            </div>
          )}
        </div>
      </div>

      {authError && (
        <div className="rounded-xl border border-rose-500/40 bg-rose-950/30 px-4 py-2 text-xs font-bold text-rose-300">
          {authError}
        </div>
      )}

      {/* Profile Header Banner */}
      <div className="relative overflow-hidden rounded-3xl border border-slate-800 bg-slate-900 p-6 sm:p-8 shadow-2xl">
        <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
          <div className="flex items-center gap-5">
            <div className="relative">
              <img
                src={avatarInput}
                alt={profile.username}
                className="h-24 w-24 rounded-3xl object-cover ring-4 ring-indigo-500/40 shadow-2xl"
              />
              <span className="absolute -bottom-1 -right-1 flex h-7 w-7 items-center justify-center rounded-xl bg-indigo-600 text-white shadow-md">
                <Zap className="h-4 w-4 fill-white" />
              </span>
            </div>

            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold text-white">{profile.username}</h1>
                <span className="rounded-full bg-indigo-500/20 px-3 py-0.5 text-xs font-bold text-indigo-300 border border-indigo-500/30">
                  {profile.title}
                </span>
              </div>
              <p className="text-xs text-slate-400 max-w-md">{profile.bio || 'Dual Exam Arena Contestant'}</p>
              <p className="text-[11px] text-slate-500">Member since {profile.joinedDate}</p>
            </div>
          </div>

          <button
            onClick={() => {
              soundManager.playClick();
              setIsEditing(!isEditing);
            }}
            className="flex items-center gap-2 rounded-2xl border border-indigo-500/40 bg-indigo-600 px-4 py-2.5 text-xs font-bold text-white hover:bg-indigo-500 transition-colors self-start md:self-auto shadow-md shadow-indigo-600/30"
          >
            <Edit3 className="h-4 w-4" />
            <span>{isEditing ? 'Cancel Edit' : 'Edit Profile'}</span>
          </button>
        </div>

        {/* Customization Drawer */}
        {isEditing && (
          <div className="mt-6 rounded-2xl border border-slate-800 bg-slate-950 p-5 space-y-4 animate-fadeIn">
            <h3 className="text-sm font-bold text-amber-400 uppercase tracking-wider">Customize Profile</h3>
            
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label className="text-xs font-bold text-slate-400 block mb-1">Display Username</label>
                <input
                  type="text"
                  value={usernameInput}
                  onChange={(e) => setUsernameInput(e.target.value)}
                  className="w-full rounded-xl border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white focus:border-amber-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="text-xs font-bold text-slate-400 block mb-1">Select Custom Title</label>
                <select
                  value={titleInput}
                  onChange={(e) => setTitleInput(e.target.value)}
                  className="w-full rounded-xl border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white focus:border-amber-500 focus:outline-none"
                >
                  {TITLE_OPTIONS.map((t) => (
                    <option key={t} value={t}>{t}</option>
                  ))}
                </select>
              </div>
            </div>

            <div>
              <label className="text-xs font-bold text-slate-400 block mb-2">Choose Avatar</label>
              <div className="flex items-center gap-3">
                {AVATAR_OPTIONS.map((img, idx) => (
                  <img
                    key={idx}
                    src={img}
                    alt="avatar option"
                    onClick={() => setAvatarInput(img)}
                    className={`h-12 w-12 cursor-pointer rounded-2xl object-cover border-2 transition-all ${
                      avatarInput === img ? 'border-orange-500 scale-110' : 'border-slate-800 opacity-60'
                    }`}
                  />
                ))}
              </div>
            </div>

            <div>
              <label className="text-xs font-bold text-slate-400 block mb-1">Bio Statement</label>
              <input
                type="text"
                value={bioInput}
                onChange={(e) => setBioInput(e.target.value)}
                className="w-full rounded-xl border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white focus:border-amber-500 focus:outline-none"
              />
            </div>

            <button
              onClick={handleSave}
              className="flex items-center gap-2 rounded-xl bg-gradient-to-r from-orange-500 to-amber-500 px-5 py-2 text-xs font-bold text-slate-950 shadow-md"
            >
              <Save className="h-4 w-4" /> Save Profile
            </button>
          </div>
        )}
      </div>

      {/* Career Statistics Grid */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <div className="rounded-3xl border border-slate-800 bg-slate-900 p-5 text-center space-y-1">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">Rating MMR</span>
          <p className="text-3xl font-black text-amber-400">{profile.mmr}</p>
          <span className="text-[10px] text-emerald-400 font-semibold">
            Tier: {
              profile.mmr >= 2000 ? '👑 Grandmaster' :
              profile.mmr >= 1800 ? '🥈 Master' :
              profile.mmr >= 1500 ? '🥉 Diamond' :
              profile.mmr >= 1200 ? '⭐ Gold II' :
              profile.mmr >= 1000 ? '⭐ Gold I' : '⭐ Bronze'
            }
          </span>
        </div>

        <div className="rounded-3xl border border-slate-800 bg-slate-900 p-5 text-center space-y-1">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">Wins / Matches</span>
          <p className="text-3xl font-black text-white">{profile.wins} / {profile.totalMatches}</p>
          <span className="text-[10px] text-cyan-400 font-semibold">
            {profile.totalMatches > 0 ? `${((profile.wins / profile.totalMatches) * 100).toFixed(0)}% Win Rate` : '0% Win Rate'}
          </span>
        </div>

        <div className="rounded-3xl border border-slate-800 bg-slate-900 p-5 text-center space-y-1">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">Avg Reaction Speed</span>
          <p className="text-3xl font-black text-orange-400">
            {profile.avgResponseTimeMs > 0 ? `${(profile.avgResponseTimeMs / 1000).toFixed(2)}s` : 'N/A'}
          </p>
          <span className="text-[10px] text-orange-300 font-semibold">
            {profile.avgResponseTimeMs > 0 ? 'Top Speed Rank' : 'Play a duel to rank'}
          </span>
        </div>

        <div className="rounded-3xl border border-slate-800 bg-slate-900 p-5 text-center space-y-1">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">Accuracy %</span>
          <p className="text-3xl font-black text-emerald-400">{profile.accuracyPercentage}%</p>
          <span className="text-[10px] text-emerald-300 font-semibold">Highest Streak: {profile.highestStreak}</span>
        </div>
      </div>

      {/* Achievements & Badges Grid */}
      <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6 space-y-4 shadow-2xl">
        <div className="flex items-center justify-between border-b border-slate-800 pb-3">
          <div className="flex items-center gap-2">
            <Award className="h-5 w-5 text-amber-400" />
            <h2 className="text-base font-bold text-white">Career Achievements</h2>
          </div>
          <span className="text-xs font-bold text-amber-400">
            {unlockedCount} / {achievements.length} Unlocked
          </span>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {achievements.map((ach) => (
            <div
              key={ach.id}
              className={`rounded-2xl border p-4 transition-all ${
                ach.isUnlocked
                  ? 'border-amber-500/40 bg-gradient-to-br from-amber-500/10 via-slate-950 to-slate-950 text-white ring-1 ring-amber-500/20'
                  : 'border-slate-800 bg-slate-950/40 text-slate-500 opacity-60'
              }`}
            >
              <div className="flex items-center justify-between">
                <span className="text-2xl">{ach.icon}</span>
                {ach.isUnlocked ? (
                  <CheckCircle2 className="h-4 w-4 text-amber-400" />
                ) : (
                  <Lock className="h-4 w-4 text-slate-600" />
                )}
              </div>

              <h4 className="mt-2 text-xs font-extrabold text-white">{ach.title}</h4>
              <p className="mt-1 text-[11px] text-slate-400 leading-tight">{ach.description}</p>

              {/* Progress bar */}
              <div className="mt-3 space-y-1">
                <div className="h-1.5 w-full rounded-full bg-slate-800 overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-orange-500 to-amber-400"
                    style={{ width: `${(ach.progress / ach.maxProgress) * 100}%` }}
                  />
                </div>
                <div className="flex justify-between text-[10px] text-slate-500 font-bold">
                  <span>{ach.rarity}</span>
                  <span>{ach.progress}/{ach.maxProgress}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
