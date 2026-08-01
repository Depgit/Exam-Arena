'use client';

import React, { useEffect, useState } from 'react';
import confetti from 'canvas-confetti';
import { 
  Trophy, 
  Flame, 
  Zap, 
  RotateCcw, 
  Home, 
  Sparkles, 
  CheckCircle, 
  XCircle, 
  Clock, 
  Brain, 
  Share2,
  BookOpen,
  ChevronDown,
  ChevronUp,
  ListChecks,
  Lightbulb
} from 'lucide-react';
import { MatchState, PlayerProfile } from '@/lib/types';
import { soundManager } from '@/lib/audio';

interface MatchSummaryModalProps {
  match: MatchState;
  userProfile: PlayerProfile;
  onPlayAgain: () => void;
  onReturnToLobby: () => void;
}

export const MatchSummaryModal: React.FC<MatchSummaryModalProps> = ({
  match,
  userProfile,
  onPlayAgain,
  onReturnToLobby,
}) => {
  const userPlayer = match.player1;
  const opponentPlayer = match.player2;

  const isWinner = match.winnerId === userPlayer.id;
  const isDraw = match.winnerId === 'draw';

  const [aiCoachAnalysis, setAiCoachAnalysis] = useState<string>('');
  const [isLoadingCoach, setIsLoadingCoach] = useState<boolean>(true);
  const [expandedQuestionId, setExpandedQuestionId] = useState<string | null>(null);

  // Trigger Victory Confetti
  useEffect(() => {
    if (isWinner) {
      soundManager.playVictory();
      confetti({
        particleCount: 100,
        spread: 70,
        origin: { y: 0.6 },
      });
    }

    // Generate AI Coach Insights client-side
    const userAvgMs = userPlayer.answers.reduce((acc, a) => acc + (a.responseTimeMs || 0), 0) / Math.max(1, userPlayer.answers.length);
    const userCorrect = userPlayer.answers.filter((a) => a.isCorrect).length;
    const totalQ = match.questions.length || 1;
    const accuracyPct = Math.round((userCorrect / totalQ) * 100);

    let insight = 'Great speed and accuracy! Keep practicing.';
    if (accuracyPct >= 80 && userAvgMs < 3000) {
      insight = '⚡ Speed Demon! Outstanding accuracy combined with lightning fast response times.';
    } else if (accuracyPct >= 80) {
      insight = '🎯 High Accuracy! Strong subject knowledge. Focus on building faster response times for speed bonuses.';
    } else if (userAvgMs < 2500) {
      insight = '🔥 Fast Reactions! Try taking an extra second to read each option carefully to boost accuracy.';
    } else {
      insight = '💪 Keep practicing! Regular duels will sharpen your concept recall and competitive stamina.';
    }

    setAiCoachAnalysis(insight);
    setIsLoadingCoach(false);
  }, []);

  // Compute total average response times
  const userAvgSpeed = (
    userPlayer.answers.reduce((acc, a) => acc + (a.responseTimeMs || 0), 0) /
    Math.max(1, userPlayer.answers.length) /
    1000
  ).toFixed(2);

  const opponentAvgSpeed = (
    opponentPlayer.answers.reduce((acc, a) => acc + (a.responseTimeMs || 0), 0) /
    Math.max(1, opponentPlayer.answers.length) /
    1000
  ).toFixed(2);

  // Compute accuracy
  const userCorrectCount = userPlayer.answers.filter((a) => a.isCorrect).length;
  const userAccuracy = ((userCorrectCount / Math.max(1, match.questions.length)) * 100).toFixed(0);

  const opponentCorrectCount = opponentPlayer.answers.filter((a) => a.isCorrect).length;
  const opponentAccuracy = ((opponentCorrectCount / Math.max(1, match.questions.length)) * 100).toFixed(0);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 p-4 backdrop-blur-md overflow-y-auto animate-fadeIn">
      <div className="my-8 w-full max-w-2xl rounded-3xl border border-slate-800 bg-slate-900 p-6 sm:p-8 shadow-2xl space-y-6 text-white">
        {/* Victory / Defeat Header Banner */}
        <div className="text-center space-y-2">
          {isWinner ? (
            <div className="inline-flex items-center gap-2 rounded-full border border-amber-500/40 bg-amber-500/10 px-4 py-1 text-sm font-black text-amber-400">
              <Trophy className="h-4 w-4 fill-amber-400" />
              <span>VICTORY! +25 MMR</span>
            </div>
          ) : isDraw ? (
            <div className="inline-flex items-center gap-2 rounded-full border border-slate-700 bg-slate-800 px-4 py-1 text-sm font-black text-slate-300">
              <span>DRAW MATCH</span>
            </div>
          ) : (
            <div className="inline-flex items-center gap-2 rounded-full border border-rose-500/40 bg-rose-500/10 px-4 py-1 text-sm font-black text-rose-400">
              <span>DEFEAT (-10 MMR)</span>
            </div>
          )}

          <h1 className="text-3xl font-black tracking-tight sm:text-4xl">
            {isWinner ? 'Match Won!' : isDraw ? 'Tie Game!' : 'Better Luck Next Round!'}
          </h1>
          <p className="text-xs text-slate-400">
            Category: <span className="font-bold text-slate-200">{match.category}</span>
          </p>
        </div>

        {/* 1v1 Scoreboard Comparison */}
        <div className="grid grid-cols-2 gap-4 rounded-3xl border border-slate-800 bg-slate-950/80 p-5">
          {/* You */}
          <div className="flex flex-col items-center text-center space-y-2 border-r border-slate-800 pr-2">
            <img
              src={userPlayer.avatarUrl}
              alt={userPlayer.username}
              className="h-16 w-16 rounded-2xl object-cover ring-2 ring-amber-500"
            />
            <span className="text-sm font-bold text-white">{userPlayer.username}</span>
            <span className="text-2xl font-black text-amber-400">{userPlayer.score} PTS</span>
            <div className="text-[11px] text-slate-400 space-y-0.5">
              <p>⚡ Speed: <strong className="text-slate-200">{userAvgSpeed}s avg</strong></p>
              <p>🎯 Accuracy: <strong className="text-slate-200">{userAccuracy}%</strong></p>
            </div>
          </div>

          {/* Opponent */}
          <div className="flex flex-col items-center text-center space-y-2 pl-2">
            <img
              src={opponentPlayer.avatarUrl}
              alt={opponentPlayer.username}
              className="h-16 w-16 rounded-2xl object-cover ring-2 ring-cyan-500"
            />
            <span className="text-sm font-bold text-white">{opponentPlayer.username}</span>
            <span className="text-2xl font-black text-cyan-400">{opponentPlayer.score} PTS</span>
            <div className="text-[11px] text-slate-400 space-y-0.5">
              <p>⚡ Speed: <strong className="text-slate-200">{opponentAvgSpeed}s avg</strong></p>
              <p>🎯 Accuracy: <strong className="text-slate-200">{opponentAccuracy}%</strong></p>
            </div>
          </div>
        </div>

        {/* Gemini AI Performance Coach Insights */}
        <div className="rounded-2xl border border-cyan-500/30 bg-gradient-to-br from-cyan-950/30 via-slate-900 to-slate-900 p-4 space-y-2">
          <div className="flex items-center gap-2 text-xs font-bold text-cyan-300 uppercase tracking-wider">
            <Brain className="h-4 w-4 text-cyan-400" />
            <span>Gemini AI Tactical Coach</span>
          </div>
          {isLoadingCoach ? (
            <p className="text-xs text-slate-400 animate-pulse">Analyzing reaction speeds and weak spots...</p>
          ) : (
            <div className="text-xs text-slate-300 leading-relaxed space-y-1.5 whitespace-pre-line">
              {aiCoachAnalysis}
            </div>
          )}
        </div>

        {/* Question-by-Question Breakdown with Deep Study Accordion */}
        <div className="space-y-2 max-h-64 overflow-y-auto pr-1">
          <div className="flex items-center justify-between">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Question Summary & Deep Step Breakdown:
            </h3>
            <span className="text-[10px] text-indigo-400 font-semibold">Click question to inspect steps</span>
          </div>

          {match.questions.map((q, idx) => {
            const uAns = userPlayer.answers[idx];
            const oAns = opponentPlayer.answers[idx];
            const isExpanded = expandedQuestionId === q.id;

            return (
              <div
                key={q.id}
                className="rounded-2xl border border-slate-800 bg-slate-950/90 overflow-hidden transition-all"
              >
                {/* Header Row */}
                <div
                  onClick={() => {
                    soundManager.playClick();
                    setExpandedQuestionId(isExpanded ? null : q.id);
                  }}
                  className="flex cursor-pointer items-center justify-between p-3 text-xs hover:bg-slate-900/60 transition-colors"
                >
                  <div className="flex items-center gap-2 min-w-0 pr-2">
                    <span className="font-black text-indigo-400">Q{idx + 1}.</span>
                    <span className="text-slate-200 font-medium truncate">{q.question}</span>
                  </div>

                  <div className="flex items-center gap-3 shrink-0">
                    <span
                      className={`flex items-center gap-1 font-bold ${
                        uAns?.isCorrect ? 'text-emerald-400' : 'text-rose-400'
                      }`}
                    >
                      You: {uAns?.isCorrect ? `${(uAns.responseTimeMs / 1000).toFixed(1)}s` : 'X'}
                    </span>
                    <span className="text-slate-600">vs</span>
                    <span
                      className={`flex items-center gap-1 font-bold ${
                        oAns?.isCorrect ? 'text-emerald-400' : 'text-rose-400'
                      }`}
                    >
                      Opp: {oAns?.isCorrect ? `${(oAns.responseTimeMs / 1000).toFixed(1)}s` : 'X'}
                    </span>
                    {isExpanded ? (
                      <ChevronUp className="h-4 w-4 text-indigo-400 shrink-0" />
                    ) : (
                      <ChevronDown className="h-4 w-4 text-slate-500 shrink-0" />
                    )}
                  </div>
                </div>

                {/* Expanded Deep Study Details */}
                {isExpanded && (
                  <div className="border-t border-slate-800/80 bg-slate-900/40 p-3.5 space-y-3 text-xs text-slate-300 animate-fadeIn">
                    {/* Concept & Formula */}
                    <div className="flex flex-wrap items-center gap-2">
                      {q.learningConcept && (
                        <span className="inline-flex items-center gap-1 rounded-md border border-indigo-500/30 bg-indigo-500/10 px-2 py-0.5 text-[10px] font-bold text-indigo-300">
                          <BookOpen className="h-3 w-3 text-indigo-400" />
                          <span>{q.learningConcept}</span>
                        </span>
                      )}
                      {q.formulaOrRule && (
                        <span className="inline-flex items-center gap-1 rounded-md border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 text-[10px] font-bold text-amber-300">
                          <Lightbulb className="h-3 w-3 text-amber-400" />
                          <span>{q.formulaOrRule}</span>
                        </span>
                      )}
                    </div>

                    {/* Summary Explanation */}
                    <p className="text-xs leading-relaxed text-slate-200">
                      <strong className="text-indigo-400">Correct Answer:</strong> {q.options[q.correctAnswerIndex]}
                      <br />
                      <span className="text-slate-400">{q.explanation}</span>
                    </p>

                    {/* Step-by-Step Resolution */}
                    {q.stepByStepAnalysis && q.stepByStepAnalysis.length > 0 && (
                      <div className="space-y-1.5 pt-1">
                        <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">
                          Step-by-Step Breakdown:
                        </span>
                        {q.stepByStepAnalysis.map((step, sIdx) => (
                          <div
                            key={sIdx}
                            className="flex items-start gap-2 rounded-lg bg-slate-950 p-2 text-[11px] text-slate-300 border border-slate-800"
                          >
                            <span className="flex h-4 w-4 shrink-0 items-center justify-center rounded bg-indigo-600/80 text-[9px] font-bold text-white">
                              {sIdx + 1}
                            </span>
                            <span className="leading-relaxed">{step}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* Action Buttons */}
        <div className="flex flex-col gap-3 sm:flex-row pt-2">
          <button
            onClick={() => {
              soundManager.playClick();
              onPlayAgain();
            }}
            className="flex flex-1 items-center justify-center gap-2 rounded-2xl bg-gradient-to-r from-orange-500 to-amber-500 py-3 text-sm font-bold text-slate-950 shadow-lg transition-transform hover:scale-105 active:scale-95"
          >
            <RotateCcw className="h-4 w-4" />
            <span>Rematch / Play Again</span>
          </button>

          <button
            onClick={() => {
              soundManager.playClick();
              onReturnToLobby();
            }}
            className="flex items-center justify-center gap-2 rounded-2xl border border-slate-800 bg-slate-950 py-3 px-6 text-sm font-bold text-slate-300 hover:bg-slate-800 hover:text-white transition-colors"
          >
            <Home className="h-4 w-4" />
            <span>Return to Lobby</span>
          </button>
        </div>
      </div>
    </div>
  );
};
