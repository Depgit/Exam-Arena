import { useEffect, useMemo, useRef, useState } from 'react'
import { useLocation, useParams, useNavigate } from 'react-router-dom'
import { getMatch } from '../../api/endpoints'
import { useAuth } from '../../context/AuthContext'
import { useWebSocket, useWSListener } from '../../context/WebSocketContext'

export default function LiveMatch() {
  const { matchId } = useParams()
  const location = useLocation()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { send } = useWebSocket()

  const [questions, setQuestions] = useState(location.state?.questions || [])
  const [players, setPlayers] = useState(location.state?.players || [])
  const [timerSeconds, setTimerSeconds] = useState(location.state?.timer_seconds ?? null)
  const [remaining, setRemaining] = useState(location.state?.timer_seconds ?? null)
  const [scoreboard, setScoreboard] = useState([])
  const [answeredIds, setAnsweredIds] = useState(new Set())
  const [current, setCurrent] = useState(0)
  const [selected, setSelected] = useState('')
  const [results, setResults] = useState(null)
  const [error, setError] = useState('')
  const questionStartRef = useRef(Date.now())

  // If we arrived here without router state (e.g. page refresh), hydrate
  // from the REST endpoint instead. Completed matches show final results.
  useEffect(() => {
    if (questions.length > 0) return
    getMatch(matchId)
      .then(({ data }) => {
        if (data.questions) setQuestions(data.questions)
        if (data.match?.timer_seconds) setTimerSeconds(data.match.timer_seconds)
        if (data.live_scores) setScoreboard(data.live_scores)
        if (data.match?.status === 'completed') {
          setResults(
            data.players.map((p) => ({
              user_id: p.user_id,
              username: p.username,
              score: p.score,
              rank: p.final_rank,
              correct: null,
              total: null,
              elo_delta: null,
              rating_before: null,
              rating_after: null,
            }))
          )
        }
      })
      .catch((err) => setError(err.message))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [matchId])

  useEffect(() => {
    questionStartRef.current = Date.now()
  }, [current])

  useWSListener('score_update', (payload) => {
    if (payload.match_id !== matchId) return
    setScoreboard(payload.scoreboard || [])
    if (payload.user_id === user.id) {
      setAnsweredIds((prev) => new Set(prev).add(payload.question_id))
    }
  })

  useWSListener('time_update', (payload) => {
    if (payload.match_id !== matchId) return
    setRemaining(payload.remaining_seconds)
  })

  useWSListener('match_end', (payload) => {
    if (payload.match_id !== matchId) return
    setResults(payload.results)
  })

  useWSListener('error', (payload) => {
    setError(payload.message)
  })

  const question = questions[current]
  const alreadyAnswered = question ? answeredIds.has(question.id) : false

  // True when this user has answered every question
  const allDone = questions.length > 0 && answeredIds.size >= questions.length

  const myScore = useMemo(
    () => scoreboard.find((s) => s.user_id === user.id),
    [scoreboard, user.id]
  )

  function submitAnswer(optionId) {
    if (!question || alreadyAnswered) return
    const timeTakenMs = Date.now() - questionStartRef.current
    setSelected(optionId)
    send('submit_answer', {
      match_id: matchId,
      question_id: question.id,
      option_id: optionId,
      time_taken_ms: timeTakenMs,
    })
  }

  function goNext() {
    setSelected('')
    setCurrent((c) => Math.min(c + 1, questions.length - 1))
  }

  // ── Results screen ──────────────────────────────────────────────────
  if (results) {
    const sorted = [...results].sort((a, b) => (a.rank ?? 99) - (b.rank ?? 99))
    const me = sorted.find((r) => r.user_id === user.id)

    return (
      <div className="page match-results-page">
        <div className="results-header">
          {me?.rank === 1 ? (
            <div className="results-trophy">🏆</div>
          ) : me?.rank === 2 ? (
            <div className="results-trophy">🥈</div>
          ) : (
            <div className="results-trophy">⚔️</div>
          )}
          <h1>{me?.rank === 1 ? 'You Won!' : me?.rank === 2 ? 'You Lost' : 'Match Over'}</h1>
          <p className="muted">{me?.correct != null ? `${me.correct}/${me.total} correct` : ''}</p>
        </div>

        <div className="results-list">
          {sorted.map((r) => {
            const isMe = r.user_id === user.id
            const delta = r.elo_delta ?? 0
            const deltaSign = delta > 0 ? '+' : ''
            const deltaClass = delta > 0 ? 'elo-gain' : delta < 0 ? 'elo-loss' : 'elo-draw'
            return (
              <div key={r.user_id} className={`result-row ${isMe ? 'me' : ''}`}>
                <span className="result-rank">#{r.rank}</span>
                <div className="result-info">
                  <span className="result-name">
                    {r.username}{isMe ? ' (you)' : ''}
                  </span>
                  {r.total != null && (
                    <span className="muted result-accuracy">{r.correct}/{r.total} correct</span>
                  )}
                </div>
                <span className="result-score">{r.score} pts</span>
                {r.elo_delta != null && (
                  <div className="result-elo">
                    <span className={`elo-delta ${deltaClass}`}>
                      {deltaSign}{Math.round(delta)} ELO
                    </span>
                    {r.rating_after != null && (
                      <span className="elo-new muted">→ {Math.round(r.rating_after)}</span>
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>

        <button className="btn-primary" onClick={() => navigate('/app/dashboard')}>
          Back to Dashboard
        </button>
      </div>
    )
  }

  if (error) {
    return (
      <div className="page">
        <div className="alert-error">{error}</div>
        <button className="btn-primary" onClick={() => navigate('/app/dashboard')}>Back to dashboard</button>
      </div>
    )
  }

  if (!question) {
    return <div className="page-center">Loading match…</div>
  }

  // ── All questions answered — waiting for opponent ────────────────────
  if (allDone && !results) {
    return (
      <div className="page match-page">
        <div className="match-header">
          <div>All questions answered ✅</div>
          {remaining != null && <div className="timer">⏱ {remaining}s</div>}
        </div>

        <div className="scoreboard">
          {(players.length > 0 ? players : scoreboard).map((p) => {
            const live = scoreboard.find((s) => s.user_id === (p.user_id || p.user_id))
            const isMe = (p.user_id) === user.id
            return (
              <div key={p.user_id} className={`score-pill ${isMe ? 'me' : ''}`}>
                {p.username}: {live?.score ?? p.score ?? 0}
              </div>
            )
          })}
        </div>

        <div className="waiting-all-done">
          <div className="spinner" />
          <p>Waiting for your opponent to finish…</p>
          <p className="muted">The match will end automatically when the timer runs out or both players finish.</p>
        </div>
      </div>
    )
  }

  // ── Active question ──────────────────────────────────────────────────
  return (
    <div className="page match-page">
      <div className="match-header">
        <div>Question {current + 1} / {questions.length}</div>
        {remaining != null && <div className="timer">⏱ {remaining}s</div>}
      </div>

      <div className="scoreboard">
        {players.length > 0
          ? players.map((p) => {
              const live = scoreboard.find((s) => s.user_id === p.user_id)
              return (
                <div key={p.user_id} className={`score-pill ${p.user_id === user.id ? 'me' : ''}`}>
                  {p.username}: {live?.score ?? 0}
                </div>
              )
            })
          : scoreboard.map((s) => (
              <div key={s.user_id} className={`score-pill ${s.user_id === user.id ? 'me' : ''}`}>
                {s.username}: {s.score}
              </div>
            ))}
      </div>

      <div className="question-card">
        <span className={`badge badge-${question.difficulty}`}>{question.difficulty}</span>
        <p className="question-body">{question.body}</p>
        <div className="options">
          {question.options
            .slice()
            .sort((a, b) => a.order_index - b.order_index)
            .map((opt) => (
              <button
                key={opt.id}
                className={`option-btn ${selected === opt.id ? 'selected' : ''}`}
                onClick={() => submitAnswer(opt.id)}
                disabled={alreadyAnswered}
              >
                {opt.option_text}
              </button>
            ))}
        </div>
        {alreadyAnswered && (
          <p className="muted">✓ Answer submitted — move to the next question.</p>
        )}
      </div>

      <div className="match-nav">
        <div className="progress-dots">
          {questions.map((q, i) => (
            <span
              key={q.id}
              className={`dot ${answeredIds.has(q.id) ? 'answered' : ''} ${i === current ? 'active' : ''}`}
              onClick={() => { setSelected(''); setCurrent(i) }}
              title={`Question ${i + 1}`}
            />
          ))}
        </div>
        <button className="btn-ghost" onClick={goNext} disabled={current >= questions.length - 1}>
          Next →
        </button>
      </div>
    </div>
  )
}
