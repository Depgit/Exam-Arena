'use client';

import React, { useState, useEffect, useRef } from 'react';
import { 
  Zap, 
  Volume2, 
  Flame, 
  Clock, 
  CheckCircle, 
  XCircle, 
  Award, 
  MessageSquare, 
  Send, 
  Sparkles,
  ArrowRight,
  Shield,
  RotateCcw,
  BookOpen,
  ListChecks,
  Lightbulb
} from 'lucide-react';
import { MatchState, Question, PlayerAnswer, MatchEmote } from '@/lib/types';
import { soundManager } from '@/lib/audio';
import { arenaBroadcast } from '@/lib/storage';

interface DualExamArenaProps {
  match: MatchState;
  onAnswerSubmitted: (questionIndex: number, selectedOptionIndex: number, responseTimeMs: number) => void;
  onNextQuestion: () => void;
  onFinishMatch: () => void;
  onSendEmote: (text: string, icon: string) => void;
  onQuitMatch: () => void;
}

const getNow = () => Date.now();

export const DualExamArena: React.FC<DualExamArenaProps> = ({
  match,
  onAnswerSubmitted,
  onNextQuestion,
  onFinishMatch,
  onSendEmote,
  onQuitMatch,
}) => {
  const currentQIndex = match.currentQuestionIndex;
  const currentQuestion: Question | undefined = match.questions[currentQIndex];

  const userPlayer = match.player1;
  const opponentPlayer = match.player2;

  const userAnswer: PlayerAnswer | undefined = userPlayer.answers[currentQIndex];
  const opponentAnswer: PlayerAnswer | undefined = opponentPlayer.answers[currentQIndex];

  // Millisecond precision timer tracking
  const questionStartTimeRef = useRef<number>(getNow());
  const [timeLeft, setTimeLeft] = useState<number>(match.timeLimitSeconds);
  const [isTTSLoading, setIsTTSLoading] = useState<boolean>(false);
  const [customChatMessage, setCustomChatMessage] = useState<string>('');

  // Derived floating emotes from match prop
  const activeEmotes = match.emotes ? match.emotes.slice(-5) : [];

  // Handle Timer Countdown
  useEffect(() => {
    questionStartTimeRef.current = getNow();
    const timeLimit = match.timeLimitSeconds;
    
    // Defer resetting state out of synchronous effect body to adhere to React compiler rules
    const timeoutId = setTimeout(() => {
      setTimeLeft(timeLimit);
    }, 0);

    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          // Auto-submit if time expires and not answered yet
          if (!userPlayer.answers[currentQIndex]) {
            const timeTaken = match.timeLimitSeconds * 1000;
            onAnswerSubmitted(currentQIndex, -1, timeTaken);
          }
          return 0;
        }
        if (prev <= 4) {
          soundManager.playTimerTick(prev - 1);
        }
        return prev - 1;
      });
    }, 1000);

    return () => {
      clearTimeout(timeoutId);
      clearInterval(timer);
    };
  }, [currentQIndex]);

  // Answer handler
  const handleSelectOption = (optionIndex: number) => {
    if (userAnswer) return; // Already answered

    const responseTimeMs = Math.max(100, getNow() - questionStartTimeRef.current);
    const isCorrect = optionIndex === currentQuestion?.correctAnswerIndex;

    if (isCorrect) {
      soundManager.playCorrect();
      if (responseTimeMs < 2000) {
        soundManager.playSpeedBonus();
      }
    } else {
      soundManager.playIncorrect();
    }

    onAnswerSubmitted(currentQIndex, optionIndex, responseTimeMs);
  };

  // Text-To-Speech Reader using Web Speech API
  const handleReadQuestion = async () => {
    if (!currentQuestion || isTTSLoading) return;
    setIsTTSLoading(true);
    soundManager.playClick();

    try {
      if ('speechSynthesis' in window) {
        window.speechSynthesis.cancel(); // Stop any previous playback
        const textToRead = `${currentQuestion.question}. Options: ${currentQuestion.options.join(', ')}`;
        const utterance = new SpeechSynthesisUtterance(textToRead);
        utterance.rate = 1.0;
        utterance.pitch = 1.0;
        window.speechSynthesis.speak(utterance);
      }
    } catch {
      // ignore SpeechSynthesis error
    } finally {
      setIsTTSLoading(false);
    }
  };

  if (!currentQuestion) {
    return (
      <div className="flex h-96 flex-col items-center justify-center text-white">
        <Sparkles className="h-10 w-10 text-amber-400 animate-spin mb-4" />
        <p className="text-lg font-bold">Preparing Dual Exam Questions...</p>
      </div>
    );
  }

  // Calculate percentage of time remaining for radial SVG bar
  const timePercent = (timeLeft / match.timeLimitSeconds) * 100;
  const bothAnswered = userAnswer !== undefined && opponentAnswer !== undefined;

  return (
    <div className="relative mx-auto max-w-5xl px-4 py-6 sm:px-6 space-y-6 animate-fadeIn">
      {/* Floating Emote Overlay */}
      <div className="pointer-events-none fixed inset-0 z-50 flex items-center justify-center overflow-hidden">
        {activeEmotes.map((e) => (
          <div
            key={e.id}
            className="animate-bounce rounded-full bg-slate-900/90 border border-amber-500/50 px-4 py-2 text-sm font-bold text-amber-300 shadow-2xl backdrop-blur-md"
          >
            <span className="mr-1.5 text-lg">{e.icon}</span>
            <span>{e.senderName}:</span>
            <span className="ml-1 text-white">{e.text}</span>
          </div>
        ))}
      </div>

      {/* Top Bar: Match Topic & Quit Button */}
      <div className="flex items-center justify-between border-b border-slate-800 pb-4">
        <div className="flex items-center gap-2">
          <span className="rounded-xl bg-orange-500/20 px-3 py-1 text-xs font-bold text-orange-400 border border-orange-500/30">
            {match.category}
          </span>
          <span className="text-xs font-semibold text-slate-400">
            Question {currentQIndex + 1} of {match.questions.length}
          </span>
        </div>

        <button
          onClick={() => {
            soundManager.playClick();
            onQuitMatch();
          }}
          className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-3 py-1.5 text-xs font-bold text-rose-400 hover:bg-rose-500/20 transition-colors"
        >
          Forfeit / Exit
        </button>
      </div>

      {/* 1v1 Player Battle Cards Header */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {/* Player 1 (User) */}
        <div
          className={`relative overflow-hidden rounded-3xl border p-5 transition-all bg-gradient-to-br from-indigo-900/20 to-slate-950 ${
            userAnswer
              ? userAnswer.isCorrect
                ? 'border-emerald-500 bg-emerald-950/20 ring-2 ring-emerald-500/30'
                : 'border-rose-500 bg-rose-950/20 ring-2 ring-rose-500/30'
              : 'border-indigo-500/40 bg-slate-900/90'
          }`}
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3.5">
              <div className="relative">
                <img
                  src={userPlayer.avatarUrl}
                  alt={userPlayer.username}
                  className="h-14 w-14 rounded-full object-cover border-2 border-indigo-500 p-0.5"
                />
                <span className="absolute -bottom-1 -right-1 bg-indigo-500 text-white text-[9px] px-1.5 py-0.5 rounded font-bold">
                  YOU
                </span>
              </div>
              <div>
                <div className="flex items-center gap-1.5">
                  <span className="font-bold text-white text-base">{userPlayer.username}</span>
                </div>
                <div className="flex items-center gap-2 text-xs text-slate-400 mt-0.5">
                  <span className="text-indigo-400 font-mono text-base font-bold">Score: {userPlayer.score}</span>
                </div>
              </div>
            </div>

            {/* User Real-Time Status */}
            {userAnswer ? (
              <div className="text-right">
                <span
                  className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-bold ${
                    userAnswer.isCorrect
                      ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
                      : 'bg-rose-500/20 text-rose-400 border border-rose-500/40'
                  }`}
                >
                  {userAnswer.isCorrect ? <CheckCircle className="h-3.5 w-3.5" /> : <XCircle className="h-3.5 w-3.5" />}
                  {(userAnswer.responseTimeMs / 1000).toFixed(2)}s
                </span>
                <p className="text-[10px] text-emerald-300 font-semibold mt-1">
                  +{userAnswer.scoreGained} pts
                </p>
              </div>
            ) : (
              <div className="flex items-center gap-1 text-xs font-semibold text-indigo-400 animate-pulse">
                <Clock className="h-3.5 w-3.5" /> Answering...
              </div>
            )}
          </div>
        </div>

        {/* Player 2 (Opponent) */}
        <div
          className={`relative overflow-hidden rounded-3xl border p-5 transition-all bg-gradient-to-bl from-rose-900/20 to-slate-950 ${
            opponentAnswer
              ? opponentAnswer.isCorrect
                ? 'border-emerald-500 bg-emerald-950/20 ring-2 ring-emerald-500/30'
                : 'border-rose-500 bg-rose-950/20 ring-2 ring-rose-500/30'
              : 'border-rose-500/40 bg-slate-900/90'
          }`}
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3.5">
              <div className="relative">
                <img
                  src={opponentPlayer.avatarUrl}
                  alt={opponentPlayer.username}
                  className="h-14 w-14 rounded-full object-cover border-2 border-rose-500 p-0.5"
                />
                <span className="absolute -bottom-1 -left-1 bg-rose-500 text-white text-[9px] px-1.5 py-0.5 rounded font-bold">
                  OPPONENT
                </span>
              </div>
              <div>
                <div className="flex items-center gap-1.5">
                  <span className="font-bold text-white text-base">{opponentPlayer.username}</span>
                  {opponentPlayer.isBot && (
                    <span className="rounded bg-purple-500/20 px-1.5 py-0.5 text-[9px] font-bold text-purple-300 border border-purple-500/30">
                      BOT ({opponentPlayer.botDifficulty})
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-2 text-xs text-slate-400 mt-0.5">
                  <span className="text-rose-400 font-mono text-base font-bold">Score: {opponentPlayer.score}</span>
                </div>
              </div>
            </div>

            {/* Opponent Real-Time Status */}
            {opponentAnswer ? (
              <div className="text-right">
                <span
                  className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-bold ${
                    opponentAnswer.isCorrect
                      ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
                      : 'bg-rose-500/20 text-rose-400 border border-rose-500/40'
                  }`}
                >
                  {opponentAnswer.isCorrect ? <CheckCircle className="h-3.5 w-3.5" /> : <XCircle className="h-3.5 w-3.5" />}
                  {(opponentAnswer.responseTimeMs / 1000).toFixed(2)}s
                </span>
                <p className="text-[10px] text-emerald-300 font-semibold mt-1">
                  +{opponentAnswer.scoreGained} pts
                </p>
              </div>
            ) : (
              <div className="flex items-center gap-1 text-xs font-semibold text-rose-400 animate-pulse">
                <Clock className="h-3.5 w-3.5" /> Thinking...
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Timer Circle & Countdown */}
      <div className="flex flex-col items-center justify-center py-2">
        <div className="relative flex h-20 w-20 items-center justify-center">
          <svg className="h-full w-full -rotate-90" viewBox="0 0 36 36">
            <path
              className="text-slate-800"
              strokeWidth="3.5"
              stroke="currentColor"
              fill="none"
              d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
            />
            <path
              className={`transition-all duration-1000 ease-linear ${
                timeLeft <= 3 ? 'text-rose-500' : timeLeft <= 7 ? 'text-amber-400' : 'text-emerald-400'
              }`}
              strokeDasharray={`${timePercent}, 100`}
              strokeWidth="3.5"
              strokeLinecap="round"
              stroke="currentColor"
              fill="none"
              d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
            />
          </svg>
          <span
            className={`absolute text-2xl font-black ${
              timeLeft <= 3 ? 'text-rose-400 animate-ping' : 'text-white'
            }`}
          >
            {timeLeft}s
          </span>
        </div>
      </div>

      {/* Main Exam Question Card */}
      <div className="rounded-3xl border-2 border-slate-700 bg-slate-900 p-6 sm:p-8 shadow-2xl space-y-6">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <span className="text-xs font-bold uppercase tracking-widest text-indigo-400">
              Difficulty: {currentQuestion.difficulty}
            </span>
            <h2 className="text-lg font-bold text-white sm:text-2xl leading-snug">
              {currentQuestion.question}
            </h2>
          </div>

          <button
            onClick={handleReadQuestion}
            title="Read Question with Voice"
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border border-slate-700 bg-slate-800 text-indigo-400 hover:bg-slate-700 transition-transform active:scale-95"
          >
            <Volume2 className="h-5 w-5" />
          </button>
        </div>

        {/* 4 Choices Grid */}
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          {currentQuestion.options.map((option, idx) => {
            const isSelected = userAnswer?.selectedOptionIndex === idx;
            const isCorrectAnswer = currentQuestion.correctAnswerIndex === idx;
            const showAnswerResults = userAnswer !== undefined;

            let buttonStyle = 'border-slate-700 bg-slate-800 text-slate-200 hover:border-indigo-500 hover:bg-slate-800/80';

            if (showAnswerResults) {
              if (isCorrectAnswer) {
                buttonStyle = 'border-emerald-500 bg-emerald-950/80 text-emerald-200 font-bold ring-2 ring-emerald-500/40';
              } else if (isSelected && !isCorrectAnswer) {
                buttonStyle = 'border-rose-500 bg-rose-950/80 text-rose-200 font-bold ring-2 ring-rose-500/40';
              } else {
                buttonStyle = 'border-slate-800 bg-slate-950/40 text-slate-500 opacity-50';
              }
            }

            const optionLetters = ['A', 'B', 'C', 'D'];

            return (
              <button
                key={idx}
                disabled={userAnswer !== undefined}
                onClick={() => handleSelectOption(idx)}
                className={`group flex items-center gap-3.5 rounded-2xl border p-4 text-left text-sm font-semibold transition-all duration-200 ${buttonStyle} disabled:cursor-not-allowed`}
              >
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-slate-900 text-xs font-black text-indigo-400 group-hover:bg-indigo-600 group-hover:text-white transition-colors">
                  {optionLetters[idx]}
                </span>
                <span className="flex-1">{option}</span>
                {showAnswerResults && isCorrectAnswer && (
                  <CheckCircle className="h-5 w-5 text-emerald-400 shrink-0" />
                )}
                {showAnswerResults && isSelected && !isCorrectAnswer && (
                  <XCircle className="h-5 w-5 text-rose-400 shrink-0" />
                )}
              </button>
            );
          })}
        </div>

        {/* Round Explanation Reveal Panel */}
        {bothAnswered && (
          <div className="rounded-2xl border border-indigo-500/30 bg-indigo-950/30 p-4 sm:p-5 space-y-4 animate-fadeIn shadow-xl">
            <div className="flex flex-wrap items-center justify-between gap-2 border-b border-indigo-500/20 pb-3">
              <div className="flex items-center gap-2 text-xs font-bold text-indigo-400 uppercase tracking-wider">
                <Sparkles className="h-4 w-4 text-indigo-400" />
                <span>Deep Analysis & Explanation</span>
              </div>
              {currentQuestion.learningConcept && (
                <span className="inline-flex items-center gap-1.5 rounded-lg border border-indigo-500/40 bg-indigo-500/10 px-2.5 py-1 text-[11px] font-bold text-indigo-300">
                  <BookOpen className="h-3 w-3 text-indigo-400" />
                  <span>{currentQuestion.learningConcept}</span>
                </span>
              )}
            </div>

            {/* Quick Rule / Formula Callout */}
            {currentQuestion.formulaOrRule && (
              <div className="flex items-start gap-2.5 rounded-xl border border-amber-500/30 bg-amber-500/10 p-3 text-xs text-amber-200">
                <Lightbulb className="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
                <div>
                  <span className="font-bold uppercase text-[10px] tracking-wider text-amber-400 block">Key Formula / Rule:</span>
                  <span className="font-semibold">{currentQuestion.formulaOrRule}</span>
                </div>
              </div>
            )}

            {/* Concise Summary Explanation */}
            <p className="text-xs text-slate-200 leading-relaxed font-medium">
              {currentQuestion.explanation}
            </p>

            {/* Step-by-Step Problem Breakdown */}
            {currentQuestion.stepByStepAnalysis && currentQuestion.stepByStepAnalysis.length > 0 && (
              <div className="space-y-2 pt-1 border-t border-slate-800/80">
                <div className="flex items-center gap-1.5 text-[11px] font-bold uppercase tracking-wider text-slate-400">
                  <ListChecks className="h-3.5 w-3.5 text-indigo-400" />
                  <span>Step-by-Step Resolution:</span>
                </div>
                <div className="space-y-1.5">
                  {currentQuestion.stepByStepAnalysis.map((step, idx) => (
                    <div
                      key={idx}
                      className="flex items-start gap-2.5 rounded-xl border border-slate-800 bg-slate-900/80 p-2.5 text-xs text-slate-300"
                    >
                      <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-lg bg-indigo-600 text-[10px] font-bold text-white">
                        {idx + 1}
                      </span>
                      <span className="leading-relaxed flex-1">{step}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="pt-2 flex justify-end">
              {currentQIndex < match.questions.length - 1 ? (
                <button
                  onClick={() => {
                    soundManager.playClick();
                    onNextQuestion();
                  }}
                  className="flex items-center gap-2 rounded-xl bg-indigo-600 px-5 py-2.5 text-xs font-bold text-white hover:bg-indigo-500 hover:scale-105 transition-all shadow-md shadow-indigo-600/30"
                >
                  <span>Next Question</span>
                  <ArrowRight className="h-4 w-4" />
                </button>
              ) : (
                <button
                  onClick={() => {
                    soundManager.playVictory();
                    onFinishMatch();
                  }}
                  className="flex items-center gap-2 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-400 px-6 py-2.5 text-xs font-bold text-slate-950 hover:scale-105 transition-transform shadow-lg shadow-emerald-500/20"
                >
                  <span>View Match Results</span>
                  <Award className="h-4 w-4" />
                </button>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Quick Reaction Emotes Bar */}
      <div className="flex items-center justify-between rounded-2xl border border-slate-800 bg-slate-900 p-3">
        <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">Quick Reaction Emotes:</span>
        <div className="flex items-center gap-1.5">
          {[
            { icon: '⚡', text: 'Too Fast!' },
            { icon: '🧠', text: 'Calculated' },
            { icon: '🔥', text: 'Speed Demon!' },
            { icon: '👏', text: 'Good Game' },
            { icon: '😅', text: 'Oops!' },
          ].map((e, idx) => (
            <button
              key={idx}
              onClick={() => {
                soundManager.playEmotePop();
                onSendEmote(e.text, e.icon);
              }}
              className="flex items-center gap-1 rounded-xl border border-slate-800 bg-slate-950 px-2.5 py-1.5 text-xs font-bold text-slate-300 hover:border-amber-500/50 hover:bg-slate-800 hover:text-white transition-all active:scale-95"
            >
              <span>{e.icon}</span>
              <span className="hidden sm:inline">{e.text}</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
};
