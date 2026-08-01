'use client';

/**
 * AuthPage.tsx — UI/UX Pro Max Enhanced Auth System
 *
 * Features:
 *  • Dual Theme Support: Dark Mode (#080b1a / Navy Glass) & Light Mode (#f8fafc / Quartz Glass)
 *  • Complete Auth Flow: Login, Register, & Interactive Forgot Password Reset
 *  • Live Password Strength Meter & Confirm Password Match indicator
 *  • Micro-animations: Floating 3D player cards, pulsing ambient orbs, tab glide, tactile buttons
 *  • Accessibility & Responsive: High contrast in both themes, clear focus rings, keyboard accessible
 */

import React, { useState } from 'react';
import {
  Mail,
  Lock,
  User,
  Eye,
  EyeOff,
  Loader2,
  PenLine,
  Zap,
  AlertCircle,
  CheckCircle2,
  Sun,
  Moon,
  ArrowLeft,
  KeyRound,
  ShieldCheck,
  Sparkles,
  Check,
  X,
} from 'lucide-react';
import { sendResetPasswordEmail } from '@/lib/firestoreService';

// ── Types ──────────────────────────────────────────────────────────────────

interface AuthPageProps {
  onLoginWithPassword: (login: string, password: string) => Promise<void>;
  onRegister: (username: string, email: string, password: string) => Promise<void>;
  onLoginWithGoogle: () => Promise<void>;
  isLoading: boolean;
  error: string | null;
  onClearError: () => void;
}

type AuthView = 'login' | 'register' | 'forgot';
type ThemeMode = 'dark' | 'light';

// ── Google SVG Icon ────────────────────────────────────────────────────────

const GoogleIcon = () => (
  <svg viewBox="0 0 24 24" className="h-[18px] w-[18px] shrink-0" aria-hidden="true">
    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
  </svg>
);

// ── Shield Logo SVG ────────────────────────────────────────────────────────

const ShieldLogo: React.FC<{ isDark: boolean }> = ({ isDark }) => (
  <svg viewBox="0 0 32 38" className="h-8 w-7 shrink-0" fill="none" aria-hidden="true">
    <path
      d="M16 0L0 7v12c0 10.5 6.8 20.3 16 23 9.2-2.7 16-12.5 16-23V7L16 0z"
      fill={`url(#sg_${isDark ? 'dark' : 'light'})`}
    />
    <path
      d="M10 18l4 4 8-8"
      stroke="white"
      strokeWidth="2.2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <defs>
      <linearGradient id={`sg_${isDark ? 'dark' : 'light'}`} x1="16" y1="0" x2="16" y2="38" gradientUnits="userSpaceOnUse">
        <stop stopColor={isDark ? '#818cf8' : '#6366f1'} />
        <stop offset="1" stopColor={isDark ? '#6366f1' : '#4f46e5'} />
      </linearGradient>
    </defs>
  </svg>
);

// ── Stat Badge ─────────────────────────────────────────────────────────────

interface StatBadgeProps {
  icon: React.ReactNode;
  value: string;
  sub?: string;
  isDark: boolean;
}

const StatBadge: React.FC<StatBadgeProps> = ({ icon, value, sub, isDark }) => (
  <div
    className="flex items-center gap-3 rounded-xl px-4 py-3 transition-all duration-300 hover:scale-[1.02]"
    style={{
      background: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(255,255,255,0.85)',
      border: isDark ? '1px solid rgba(255,255,255,0.08)' : '1px solid rgba(226,232,240,0.8)',
      boxShadow: isDark ? 'none' : '0 4px 12px rgba(99,102,241,0.05)',
    }}
  >
    <span className={isDark ? 'text-indigo-400' : 'text-indigo-600'}>{icon}</span>
    <div>
      <div className={`text-sm font-bold leading-none ${isDark ? 'text-white' : 'text-slate-900'}`}>{value}</div>
      {sub && <div className={`mt-1 text-[11px] font-medium ${isDark ? 'text-slate-400' : 'text-slate-500'}`}>{sub}</div>}
    </div>
  </div>
);

// ── Floating 3-D Player Card ────────────────────────────────────────────────

interface PlayerCardProps {
  name: string;
  mmr: number;
  avatar: string;
  className?: string;
  isDark: boolean;
}

const PlayerCard: React.FC<PlayerCardProps> = ({ name, mmr, avatar, className, isDark }) => (
  <div
    className={`absolute rounded-2xl p-3.5 backdrop-blur-md transition-all duration-500 hover:scale-110 ${className}`}
    style={{
      background: isDark ? 'rgba(18, 21, 48, 0.88)' : 'rgba(255, 255, 255, 0.95)',
      border: isDark ? '1px solid rgba(255,255,255,0.15)' : '1px solid rgba(226,232,240,0.9)',
      width: 136,
      boxShadow: isDark ? '0 20px 50px rgba(0,0,0,0.65)' : '0 20px 40px rgba(99,102,241,0.15)',
    }}
  >
    <div className="mx-auto mb-2 h-14 w-14 overflow-hidden rounded-full ring-2 ring-indigo-500/50 shadow-md">
      <img src={avatar} alt={name} className="h-full w-full object-cover" />
    </div>
    <div className="mb-1 text-center text-base leading-none">🦅</div>
    <div className={`text-center text-xs font-bold truncate ${isDark ? 'text-white' : 'text-slate-900'}`}>{name}</div>
    <div className={`mt-0.5 text-center text-[10px] font-bold tracking-wider ${isDark ? 'text-indigo-300' : 'text-indigo-600'}`}>
      MMR {mmr}
    </div>
  </div>
);

// ── Input Field ────────────────────────────────────────────────────────────

interface InputFieldProps {
  id: string;
  type: string;
  placeholder: string;
  value: string;
  onChange: (v: string) => void;
  icon: React.ReactNode;
  rightSlot?: React.ReactNode;
  autoComplete?: string;
  isDark: boolean;
}

const InputField: React.FC<InputFieldProps> = ({
  id, type, placeholder, value, onChange, icon, rightSlot, autoComplete, isDark,
}) => (
  <div className="relative group">
    <span className={`pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 transition-colors ${
      isDark ? 'text-indigo-400/60 group-focus-within:text-indigo-400' : 'text-indigo-500/60 group-focus-within:text-indigo-600'
    }`}>
      {icon}
    </span>
    <input
      id={id}
      type={type}
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      autoComplete={autoComplete}
      style={{
        background: isDark ? 'rgba(255,255,255,0.05)' : 'rgba(241,245,249,0.8)',
        border: isDark ? '1px solid rgba(255,255,255,0.1)' : '1px solid rgba(203,213,225,0.8)',
      }}
      className={`w-full rounded-xl py-3.5 pl-11 pr-11 text-sm outline-none transition-all ${
        isDark
          ? 'text-white placeholder-slate-500 focus:border-indigo-500/80 focus:bg-slate-900/60 focus:ring-2 focus:ring-indigo-500/20'
          : 'text-slate-900 placeholder-slate-400 focus:border-indigo-600 focus:bg-white focus:ring-2 focus:ring-indigo-500/20'
      }`}
    />
    {rightSlot && (
      <span className="absolute right-4 top-1/2 -translate-y-1/2">{rightSlot}</span>
    )}
  </div>
);

// ── Password Strength Calculator ──────────────────────────────────────────

function getPasswordStrength(pw: string): { score: number; label: string; color: string } {
  if (!pw) return { score: 0, label: '', color: '' };
  let score = 0;
  if (pw.length >= 8) score++;
  if (/[A-Z]/.test(pw) && /[a-z]/.test(pw)) score++;
  if (/[0-9]/.test(pw)) score++;
  if (/[^A-Za-z0-9]/.test(pw)) score++;

  if (score <= 1) return { score: 1, label: 'Weak', color: 'bg-rose-500' };
  if (score === 2 || score === 3) return { score: 2, label: 'Medium', color: 'bg-amber-500' };
  return { score: 3, label: 'Strong', color: 'bg-emerald-500' };
}

// ── Main Component ─────────────────────────────────────────────────────────

export const AuthPage: React.FC<AuthPageProps> = ({
  onLoginWithPassword,
  onRegister,
  onLoginWithGoogle,
  isLoading,
  error,
  onClearError,
}) => {
  const [view, setView] = useState<AuthView>('login');
  const [theme, setTheme] = useState<ThemeMode>('dark');

  // Login form state
  const [loginField, setLoginField] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [showLoginPw, setShowLoginPw] = useState(false);

  // Register form state
  const [regUsername, setRegUsername] = useState('');
  const [regEmail, setRegEmail] = useState('');
  const [regPassword, setRegPassword] = useState('');
  const [regConfirm, setRegConfirm] = useState('');
  const [showRegPw, setShowRegPw] = useState(false);
  const [regSuccess, setRegSuccess] = useState(false);

  // Forgot password form state
  const [forgotEmail, setForgotEmail] = useState('');
  const [forgotSent, setForgotSent] = useState(false);
  const [forgotLoading, setForgotLoading] = useState(false);

  const [fieldError, setFieldError] = useState<string | null>(null);

  const isDark = theme === 'dark';

  const toggleTheme = () => {
    setTheme((t) => (t === 'dark' ? 'light' : 'dark'));
  };

  const clearErrors = () => {
    setFieldError(null);
    onClearError();
  };

  const switchView = (v: AuthView) => {
    setView(v);
    clearErrors();
    setRegSuccess(false);
    setForgotSent(false);
  };

  // ── Form Submissions ──────────────────────────────────────────────────────

  const handleLoginSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearErrors();
    if (!loginField.trim() || !loginPassword) {
      setFieldError('Please enter your email/username and password.');
      return;
    }
    try {
      await onLoginWithPassword(loginField.trim(), loginPassword);
    } catch {
      // Handled by hook
    }
  };

  const handleRegisterSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearErrors();
    if (regUsername.trim().length < 3) {
      setFieldError('Username must be at least 3 characters.');
      return;
    }
    if (!/\S+@\S+\.\S+/.test(regEmail.trim())) {
      setFieldError('Please enter a valid email address.');
      return;
    }
    if (regPassword.length < 8) {
      setFieldError('Password must be at least 8 characters.');
      return;
    }
    if (regPassword !== regConfirm) {
      setFieldError('Passwords do not match.');
      return;
    }

    try {
      await onRegister(regUsername.trim(), regEmail.trim().toLowerCase(), regPassword);
      setRegSuccess(true);
    } catch {
      // Handled by hook
    }
  };

  const handleForgotSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearErrors();
    if (!/\S+@\S+\.\S+/.test(forgotEmail.trim())) {
      setFieldError('Please enter a valid email address.');
      return;
    }
    setForgotLoading(true);
    try {
      await sendResetPasswordEmail(forgotEmail.trim());
      setForgotSent(true);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to send password reset email.';
      setFieldError(msg);
    } finally {
      setForgotLoading(false);
    }
  };

  const handleGoogle = async () => {
    clearErrors();
    try {
      await onLoginWithGoogle();
    } catch {
      // Handled by hook
    }
  };

  const displayError = fieldError || error;
  const pwStrength = getPasswordStrength(regPassword);

  return (
    <div
      className={`relative flex min-h-screen w-full flex-col items-center justify-between overflow-hidden px-4 py-6 transition-colors duration-500 ${
        isDark ? 'bg-[#080b1a] text-slate-100' : 'bg-[#f8fafc] text-slate-900'
      }`}
    >
      {/* ── Background Glows & Ambience ──────────────────────────────────── */}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        {/* Stadium horizon glow */}
        <div
          className="absolute bottom-0 left-1/2 -translate-x-1/2 transition-opacity duration-700"
          style={{
            width: '140%',
            height: '60%',
            background: isDark
              ? 'radial-gradient(ellipse 65% 45% at 50% 100%, rgba(79,70,229,0.22) 0%, transparent 70%)'
              : 'radial-gradient(ellipse 65% 45% at 50% 100%, rgba(99,102,241,0.12) 0%, transparent 70%)',
          }}
        />

        {/* Pulsing Ambient Orbs */}
        <div
          className="animate-pulse-glow absolute -left-20 top-[15%] h-80 w-80 rounded-full blur-[100px]"
          style={{
            background: isDark
              ? 'radial-gradient(circle, #6366f1 0%, transparent 70%)'
              : 'radial-gradient(circle, #a5b4fc 0%, transparent 70%)',
            opacity: isDark ? 0.25 : 0.35,
          }}
        />
        <div
          className="animate-pulse-glow absolute -right-20 bottom-[15%] h-96 w-96 rounded-full blur-[120px]"
          style={{
            background: isDark
              ? 'radial-gradient(circle, #8b5cf6 0%, transparent 70%)'
              : 'radial-gradient(circle, #c4b5fd 0%, transparent 70%)',
            opacity: isDark ? 0.2 : 0.3,
            animationDelay: '2s',
          }}
        />

        {/* Glowing dot orbs (Dark mode) */}
        {isDark && (
          <>
            <div className="absolute h-3 w-3 rounded-full" style={{ top: '20%', left: '10%', background: '#818cf8', boxShadow: '0 0 14px 4px rgba(129,140,248,0.7)' }} />
            <div className="absolute h-2 w-2 rounded-full" style={{ top: '58%', left: '6%', background: '#a78bfa', boxShadow: '0 0 10px 3px rgba(167,139,250,0.6)' }} />
            <div className="absolute h-4 w-4 rounded-full" style={{ top: '75%', left: '42%', background: '#6366f1', boxShadow: '0 0 18px 5px rgba(99,102,241,0.6)' }} />
            <div className="absolute h-2.5 w-2.5 rounded-full" style={{ top: '12%', right: '28%', background: '#818cf8', boxShadow: '0 0 12px 3px rgba(129,140,248,0.5)' }} />
          </>
        )}

        {/* Star sparkles */}
        <div className={`absolute top-8 right-16 text-lg pointer-events-none ${isDark ? 'text-indigo-400/40' : 'text-indigo-500/30'}`}>✦</div>
        <div className={`absolute top-20 right-28 text-xs pointer-events-none ${isDark ? 'text-indigo-400/30' : 'text-indigo-500/20'}`}>✦</div>
        <div className={`absolute bottom-32 left-12 text-sm pointer-events-none ${isDark ? 'text-violet-400/30' : 'text-violet-500/20'}`}>✦</div>
      </div>

      {/* ── Top Header Controls ──────────────────────────────────────────── */}
      <header className="relative z-20 flex w-full max-w-6xl items-center justify-between py-2">
        <div className="flex items-center gap-2.5">
          <ShieldLogo isDark={isDark} />
          <div className="flex items-baseline gap-1.5">
            <span className={`text-xl font-black tracking-tight ${isDark ? 'text-white' : 'text-slate-900'}`}>DUAL EXAM</span>
            <span className={`text-xl font-black ${isDark ? 'text-indigo-400' : 'text-indigo-600'}`}>PRO</span>
          </div>
        </div>

        {/* Theme Toggle Button */}
        <button
          id="btn-theme-toggle"
          type="button"
          onClick={toggleTheme}
          title={`Switch to ${isDark ? 'Light' : 'Dark'} Mode`}
          className={`flex h-10 w-10 items-center justify-center rounded-xl border transition-all duration-300 hover:scale-105 active:scale-95 ${
            isDark
              ? 'border-slate-800 bg-slate-900/80 text-amber-400 hover:bg-slate-800 hover:border-slate-700'
              : 'border-slate-200 bg-white/90 text-indigo-600 shadow-sm hover:bg-slate-100 hover:border-slate-300'
          }`}
        >
          {isDark ? <Sun className="h-5 w-5 fill-amber-400/20" /> : <Moon className="h-5 w-5 fill-indigo-600/20" />}
        </button>
      </header>

      {/* ── Center Content: Hero Left + Glass Auth Card Right ───────────── */}
      <main className="relative z-10 my-auto flex w-full max-w-6xl items-center justify-between gap-12 py-6">

        {/* ════════════════════════════════════════════════════════════════
            LEFT HERO PANEL (Desktop)
        ════════════════════════════════════════════════════════════════ */}
        <div className="hidden lg:flex flex-col flex-1 max-w-[500px]">
          <div className="inline-flex items-center gap-2 rounded-full px-3.5 py-1.5 text-xs font-bold mb-6 w-fit border shadow-sm"
            style={{
              background: isDark ? 'rgba(99,102,241,0.12)' : 'rgba(99,102,241,0.08)',
              borderColor: isDark ? 'rgba(99,102,241,0.3)' : 'rgba(99,102,241,0.25)',
              color: isDark ? '#a5b4fc' : '#4f46e5',
            }}
          >
            <Sparkles className="h-3.5 w-3.5" />
            <span>1v1 Real-Time Speed &amp; Accuracy Arena</span>
          </div>

          <h1 className={`mb-6 font-black leading-[1.1] ${isDark ? 'text-white' : 'text-slate-900'}`} style={{ fontSize: 'clamp(2.4rem, 3.4vw, 3.4rem)' }}>
            Challenge the World&apos;s Best in Real-Time Exam Duels
          </h1>

          <p className={`mb-8 text-base leading-relaxed ${isDark ? 'text-slate-400' : 'text-slate-600'}`}>
            Battle live opponents, race against the clock, earn MMR points, and dominate the global leaderboard across 10+ exam categories.
          </p>

          {/* 2-Column Stat Badges */}
          <div className="grid grid-cols-2 gap-3 mb-3">
            <StatBadge
              isDark={isDark}
              icon={
                <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              }
              value="1,248"
              sub="Online Players"
            />
            <StatBadge
              isDark={isDark}
              icon={
                <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 3H5a2 2 0 00-2 2v4m6-6h10a2 2 0 012 2v4M9 3v18m0 0h10a2 2 0 002-2V9M9 21H5a2 2 0 01-2-2V9m0 0h18" />
                </svg>
              }
              value="4,891"
              sub="Duels Today"
            />
          </div>

          <StatBadge
            isDark={isDark}
            icon={
              <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
              </svg>
            }
            value="#1 Grandmaster"
            sub="ApexPredator (MMR 2,450)"
          />

          {/* Animated 3-D Floating Player Cards */}
          <div className="relative mt-8 h-48 w-full hidden xl:block">
            <PlayerCard
              isDark={isDark}
              name="Prodigy"
              mmr={2145}
              avatar="https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=250"
              className="animate-float-slow left-4 top-2"
            />
            <PlayerCard
              isDark={isDark}
              name="Exam Slayer"
              mmr={2088}
              avatar="https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=250"
              className="animate-float-reverse left-36 top-10"
            />
          </div>
        </div>

        {/* ════════════════════════════════════════════════════════════════
            RIGHT PANEL — GLASS AUTH CARD
        ════════════════════════════════════════════════════════════════ */}
        <div className="w-full max-w-[420px] shrink-0 mx-auto">
          <div
            className="rounded-2xl p-7 shadow-2xl transition-all duration-300"
            style={{
              background: isDark ? 'rgba(12, 14, 35, 0.92)' : 'rgba(255, 255, 255, 0.95)',
              border: isDark ? '1px solid rgba(255,255,255,0.09)' : '1px solid rgba(226,232,240,0.9)',
              boxShadow: isDark
                ? '0 25px 50px -12px rgba(0, 0, 0, 0.7)'
                : '0 20px 45px -10px rgba(99, 102, 241, 0.12)',
              backdropFilter: 'blur(20px)',
            }}
          >
            {/* Mobile Brand */}
            <div className="mb-5 flex items-center gap-2.5 lg:hidden">
              <ShieldLogo isDark={isDark} />
              <span className={`text-base font-black tracking-tight ${isDark ? 'text-white' : 'text-slate-900'}`}>
                DUAL EXAM <span className={isDark ? 'text-indigo-400' : 'text-indigo-600'}>PRO</span>
              </span>
            </div>

            {/* ── Tab Row (Login / Register) — Only when not in Forgot view ── */}
            {view !== 'forgot' && (
              <div
                className="mb-6 flex items-center gap-1 rounded-xl p-1 transition-all"
                style={{
                  background: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(241,245,249,0.9)',
                  border: isDark ? '1px solid rgba(255,255,255,0.07)' : '1px solid rgba(203,213,225,0.6)',
                }}
              >
                <button
                  id="auth-tab-login"
                  onClick={() => switchView('login')}
                  className="flex-1 rounded-lg py-2.5 text-sm font-bold transition-all duration-200"
                  style={
                    view === 'login'
                      ? {
                          background: isDark ? '#5b5ef4' : '#4f46e5',
                          color: '#fff',
                          boxShadow: isDark ? '0 4px 16px rgba(91,94,244,0.4)' : '0 4px 12px rgba(79,70,229,0.3)',
                        }
                      : { color: isDark ? 'rgba(255,255,255,0.45)' : 'rgba(71,85,105,0.7)' }
                  }
                >
                  Sign In
                </button>
                <button
                  id="auth-tab-register"
                  onClick={() => switchView('register')}
                  className="flex-1 rounded-lg py-2.5 text-sm font-bold transition-all duration-200"
                  style={
                    view === 'register'
                      ? {
                          background: isDark ? '#5b5ef4' : '#4f46e5',
                          color: '#fff',
                          boxShadow: isDark ? '0 4px 16px rgba(91,94,244,0.4)' : '0 4px 12px rgba(79,70,229,0.3)',
                        }
                      : { color: isDark ? 'rgba(255,255,255,0.45)' : 'rgba(71,85,105,0.7)' }
                  }
                >
                  Create Account
                </button>
              </div>
            )}

            {/* ── Error Banner ──────────────────────────────────────────── */}
            {displayError && (
              <div
                className="mb-4 flex items-start gap-2.5 rounded-xl px-3.5 py-3 text-sm transition-all"
                style={{
                  background: isDark ? 'rgba(239,68,68,0.12)' : 'rgba(254,242,242,0.9)',
                  border: isDark ? '1px solid rgba(239,68,68,0.3)' : '1px solid rgba(252,165,165,0.8)',
                  color: isDark ? '#fca5a5' : '#b91c1c',
                }}
              >
                <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                <span>{displayError}</span>
              </div>
            )}

            {/* ── Register Success Banner ───────────────────────────────── */}
            {regSuccess && (
              <div
                className="mb-4 flex items-start gap-2.5 rounded-xl px-3.5 py-3 text-sm transition-all"
                style={{
                  background: isDark ? 'rgba(16,185,129,0.12)' : 'rgba(236,253,245,0.9)',
                  border: isDark ? '1px solid rgba(16,185,129,0.3)' : '1px solid rgba(110,231,183,0.8)',
                  color: isDark ? '#6ee7b7' : '#047857',
                }}
              >
                <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0" />
                <span>Welcome to the Arena, <strong>{regUsername}</strong>! Account ready. ⚡</span>
              </div>
            )}

            {/* ════════════════════════════════════════════════════════════
                VIEW 1: SIGN IN
            ════════════════════════════════════════════════════════════ */}
            {view === 'login' && (
              <form onSubmit={handleLoginSubmit} className="space-y-3.5" noValidate>
                <InputField
                  isDark={isDark}
                  id="login-email"
                  type="text"
                  placeholder="Email Address or Username"
                  value={loginField}
                  onChange={setLoginField}
                  autoComplete="username email"
                  icon={<Mail className="h-4 w-4" />}
                />
                <InputField
                  isDark={isDark}
                  id="login-password"
                  type={showLoginPw ? 'text' : 'password'}
                  placeholder="Password"
                  value={loginPassword}
                  onChange={setLoginPassword}
                  autoComplete="current-password"
                  icon={<Lock className="h-4 w-4" />}
                  rightSlot={
                    <button
                      type="button"
                      onClick={() => setShowLoginPw((p) => !p)}
                      className="transition-colors hover:opacity-100"
                      style={{ color: isDark ? 'rgba(255,255,255,0.4)' : 'rgba(100,116,139,0.7)' }}
                      tabIndex={-1}
                    >
                      {showLoginPw ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                    </button>
                  }
                />

                <button
                  id="btn-login-submit"
                  type="submit"
                  disabled={isLoading}
                  className="relative mt-2 flex w-full items-center justify-center rounded-xl py-3.5 text-sm font-bold text-white transition-all duration-200 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed hover:shadow-lg"
                  style={{
                    background: isDark ? '#5b5ef4' : '#4f46e5',
                    boxShadow: isDark ? '0 4px 20px rgba(91,94,244,0.45)' : '0 4px 18px rgba(79,70,229,0.35)',
                  }}
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Signing In…
                    </>
                  ) : (
                    <>
                      <span className="flex-1 text-center">Enter the Arena</span>
                      <PenLine className="absolute right-4 h-4 w-4 opacity-80" />
                    </>
                  )}
                </button>
              </form>
            )}

            {/* ════════════════════════════════════════════════════════════
                VIEW 2: CREATE ACCOUNT (REGISTER)
            ════════════════════════════════════════════════════════════ */}
            {view === 'register' && (
              <form onSubmit={handleRegisterSubmit} className="space-y-3" noValidate>
                <InputField
                  isDark={isDark}
                  id="reg-username"
                  type="text"
                  placeholder="Username (e.g. SpeedDuellist)"
                  value={regUsername}
                  onChange={setRegUsername}
                  autoComplete="username"
                  icon={<User className="h-4 w-4" />}
                />
                <InputField
                  isDark={isDark}
                  id="reg-email"
                  type="email"
                  placeholder="Email Address"
                  value={regEmail}
                  onChange={setRegEmail}
                  autoComplete="email"
                  icon={<Mail className="h-4 w-4" />}
                />
                <InputField
                  isDark={isDark}
                  id="reg-password"
                  type={showRegPw ? 'text' : 'password'}
                  placeholder="Password (min 8 chars)"
                  value={regPassword}
                  onChange={setRegPassword}
                  autoComplete="new-password"
                  icon={<Lock className="h-4 w-4" />}
                  rightSlot={
                    <button
                      type="button"
                      onClick={() => setShowRegPw((p) => !p)}
                      className="transition-colors"
                      style={{ color: isDark ? 'rgba(255,255,255,0.4)' : 'rgba(100,116,139,0.7)' }}
                      tabIndex={-1}
                    >
                      {showRegPw ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                    </button>
                  }
                />

                {/* Password Strength Meter */}
                {regPassword && (
                  <div className="px-1 pt-0.5 space-y-1">
                    <div className="flex items-center justify-between text-[11px] font-semibold">
                      <span className={isDark ? 'text-slate-400' : 'text-slate-500'}>Password Strength:</span>
                      <span className={
                        pwStrength.score === 1 ? 'text-rose-400' :
                        pwStrength.score === 2 ? 'text-amber-400' : 'text-emerald-400'
                      }>
                        {pwStrength.label}
                      </span>
                    </div>
                    <div className="flex h-1.5 w-full gap-1.5 rounded-full overflow-hidden bg-slate-700/30">
                      <div className={`h-full flex-1 transition-all duration-300 ${pwStrength.score >= 1 ? pwStrength.color : 'opacity-20'}`} />
                      <div className={`h-full flex-1 transition-all duration-300 ${pwStrength.score >= 2 ? pwStrength.color : 'opacity-20'}`} />
                      <div className={`h-full flex-1 transition-all duration-300 ${pwStrength.score >= 3 ? pwStrength.color : 'opacity-20'}`} />
                    </div>
                  </div>
                )}

                <div>
                  <InputField
                    isDark={isDark}
                    id="reg-confirm"
                    type={showRegPw ? 'text' : 'password'}
                    placeholder="Confirm Password"
                    value={regConfirm}
                    onChange={setRegConfirm}
                    autoComplete="new-password"
                    icon={<Lock className="h-4 w-4" />}
                  />
                  {regConfirm && (
                    <p className={`mt-1 text-[11px] font-semibold px-1 flex items-center gap-1 ${
                      regPassword === regConfirm ? 'text-emerald-400' : 'text-rose-400'
                    }`}>
                      {regPassword === regConfirm ? (
                        <><Check className="h-3 w-3" /> Passwords match</>
                      ) : (
                        <><X className="h-3 w-3" /> Passwords do not match</>
                      )}
                    </p>
                  )}
                </div>

                <button
                  id="btn-register-submit"
                  type="submit"
                  disabled={isLoading}
                  className="relative mt-2 flex w-full items-center justify-center rounded-xl py-3.5 text-sm font-bold text-white transition-all duration-200 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed hover:shadow-lg"
                  style={{
                    background: isDark ? '#5b5ef4' : '#4f46e5',
                    boxShadow: isDark ? '0 4px 20px rgba(91,94,244,0.45)' : '0 4px 18px rgba(79,70,229,0.35)',
                  }}
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Creating Account…
                    </>
                  ) : (
                    <>
                      <span className="flex-1 text-center">Join the Arena</span>
                      <Zap className="absolute right-4 h-4 w-4 fill-white opacity-80" />
                    </>
                  )}
                </button>
              </form>
            )}

            {/* ════════════════════════════════════════════════════════════
                VIEW 3: FORGOT PASSWORD RESET
            ════════════════════════════════════════════════════════════ */}
            {view === 'forgot' && (
              <div className="space-y-4">
                <button
                  type="button"
                  onClick={() => switchView('login')}
                  className={`inline-flex items-center gap-1.5 text-xs font-semibold transition-colors ${
                    isDark ? 'text-indigo-400 hover:text-indigo-300' : 'text-indigo-600 hover:text-indigo-700'
                  }`}
                >
                  <ArrowLeft className="h-3.5 w-3.5" /> Back to Sign In
                </button>

                <div className="text-center py-2">
                  <div className={`mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-2xl ${
                    isDark ? 'bg-indigo-600/20 text-indigo-400 border border-indigo-500/30' : 'bg-indigo-100 text-indigo-600'
                  }`}>
                    <KeyRound className="h-6 w-6" />
                  </div>
                  <h2 className={`text-lg font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>Reset Password</h2>
                  <p className={`mt-1 text-xs ${isDark ? 'text-slate-400' : 'text-slate-500'}`}>
                    Enter your account email and we&apos;ll send you a password reset link.
                  </p>
                </div>

                {forgotSent ? (
                  <div
                    className="rounded-xl p-4 text-center space-y-2 animate-fadeIn"
                    style={{
                      background: isDark ? 'rgba(16,185,129,0.12)' : 'rgba(236,253,245,0.9)',
                      border: isDark ? '1px solid rgba(16,185,129,0.3)' : '1px solid rgba(110,231,183,0.8)',
                    }}
                  >
                    <ShieldCheck className={`mx-auto h-8 w-8 ${isDark ? 'text-emerald-400' : 'text-emerald-600'}`} />
                    <p className={`text-sm font-bold ${isDark ? 'text-emerald-300' : 'text-emerald-900'}`}>
                      Reset Link Sent!
                    </p>
                    <p className={`text-xs ${isDark ? 'text-emerald-400/80' : 'text-emerald-700'}`}>
                      We&apos;ve sent instructions to <strong>{forgotEmail}</strong>. Check your inbox and spam folder.
                    </p>
                    <button
                      type="button"
                      onClick={() => switchView('login')}
                      className={`mt-2 w-full rounded-xl py-2.5 text-xs font-bold transition-all ${
                        isDark ? 'bg-slate-800 text-white hover:bg-slate-700' : 'bg-white text-slate-800 hover:bg-slate-100 shadow-sm'
                      }`}
                    >
                      Return to Sign In
                    </button>
                  </div>
                ) : (
                  <form onSubmit={handleForgotSubmit} className="space-y-3.5">
                    <InputField
                      isDark={isDark}
                      id="forgot-email"
                      type="email"
                      placeholder="Your Email Address"
                      value={forgotEmail}
                      onChange={setForgotEmail}
                      autoComplete="email"
                      icon={<Mail className="h-4 w-4" />}
                    />
                    <button
                      type="submit"
                      disabled={forgotLoading}
                      className="w-full rounded-xl py-3 text-sm font-bold text-white transition-all active:scale-[0.98] disabled:opacity-60"
                      style={{
                        background: isDark ? '#5b5ef4' : '#4f46e5',
                        boxShadow: isDark ? '0 4px 20px rgba(91,94,244,0.4)' : '0 4px 18px rgba(79,70,229,0.3)',
                      }}
                    >
                      {forgotLoading ? (
                        <span className="flex items-center justify-center gap-2">
                          <Loader2 className="h-4 w-4 animate-spin" /> Sending Link…
                        </span>
                      ) : (
                        'Send Reset Link'
                      )}
                    </button>
                  </form>
                )}
              </div>
            )}

            {/* ── Social Divider (Only when not in Forgot view) ───────────── */}
            {view !== 'forgot' && (
              <>
                <div className="my-5 flex items-center gap-3">
                  <div className="flex-1 border-t" style={{ borderColor: isDark ? 'rgba(255,255,255,0.08)' : 'rgba(203,213,225,0.8)' }} />
                  <span className="text-xs font-medium" style={{ color: isDark ? 'rgba(255,255,255,0.35)' : 'rgba(100,116,139,0.8)' }}>
                    or continue with
                  </span>
                  <div className="flex-1 border-t" style={{ borderColor: isDark ? 'rgba(255,255,255,0.08)' : 'rgba(203,213,225,0.8)' }} />
                </div>

                {/* ── Google Button ───────────────────────────────────────── */}
                <button
                  id="btn-google-signin"
                  type="button"
                  onClick={handleGoogle}
                  disabled={isLoading}
                  className="flex w-full items-center justify-center gap-2.5 rounded-xl py-3.5 text-sm font-semibold transition-all duration-200 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
                  style={{
                    background: isDark ? 'rgba(255,255,255,0.06)' : 'rgba(241,245,249,0.9)',
                    border: isDark ? '1px solid rgba(255,255,255,0.1)' : '1px solid rgba(203,213,225,0.8)',
                    color: isDark ? '#ffffff' : '#0f172a',
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = isDark ? 'rgba(255,255,255,0.1)' : 'rgba(226,232,240,0.9)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = isDark ? 'rgba(255,255,255,0.06)' : 'rgba(241,245,249,0.9)';
                  }}
                >
                  <GoogleIcon />
                  Continue with Google
                </button>
              </>
            )}

            {/* ── Footer Navigation & Links ───────────────────────────────── */}
            <div className="mt-5 space-y-1.5 text-center">
              {view === 'login' && (
                <p className="text-xs">
                  <button
                    type="button"
                    onClick={() => switchView('forgot')}
                    className={`font-semibold transition-colors hover:underline ${
                      isDark ? 'text-slate-400 hover:text-slate-200' : 'text-slate-500 hover:text-slate-700'
                    }`}
                  >
                    Forgot Password?
                  </button>
                </p>
              )}

              {view !== 'forgot' && (
                <p className="text-xs" style={{ color: isDark ? 'rgba(255,255,255,0.4)' : 'rgba(100,116,139,0.8)' }}>
                  {view === 'login' ? (
                    <>
                      Don&apos;t have an account?{' '}
                      <button
                        type="button"
                        onClick={() => switchView('register')}
                        className={`font-bold transition-colors hover:underline ${
                          isDark ? 'text-indigo-400' : 'text-indigo-600'
                        }`}
                      >
                        Sign Up
                      </button>
                    </>
                  ) : (
                    <>
                      Already have an account?{' '}
                      <button
                        type="button"
                        onClick={() => switchView('login')}
                        className={`font-bold transition-colors hover:underline ${
                          isDark ? 'text-indigo-400' : 'text-indigo-600'
                        }`}
                      >
                        Sign In
                      </button>
                    </>
                  )}
                </p>
              )}

              <p className={`text-[11px] pt-1 ${isDark ? 'text-slate-600' : 'text-slate-400'}`}>
                <button type="button" className="hover:underline mr-3">Terms</button>
                <button type="button" className="hover:underline">Privacy</button>
              </p>
            </div>
          </div>
        </div>
      </main>

      {/* ── Footer Branding Bar ─────────────────────────────────────────── */}
      <footer className="relative z-20 w-full max-w-6xl text-center py-2">
        <p className={`text-xs ${isDark ? 'text-slate-500' : 'text-slate-400'}`}>
          &copy; {new Date().getFullYear()} Dual Exam Pro. All rights reserved. Real-time competitive exam platform.
        </p>
      </footer>
    </div>
  );
};
