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

  if (results) {
    const sorted = [...results].sort((a, b) => (a.rank ?? 99) - (b.rank ?? 99))
    return (
      <div className="page">
        <h1>Match results</h1>
        <div className="results-list">
          {sorted.map((r) => (
            <div key={r.user_id} className={`result-row ${r.user_id === user.id ? 'me' : ''}`}>
              <span className="result-rank">#{r.rank}</span>
              <span className="result-name">{r.username}{r.user_id === user.id ? ' (you)' : ''}</span>
              <span className="result-score">{r.score} pts</span>
              {r.total != null && <span className="muted">{r.correct}/{r.total} correct</span>}
            </div>
          ))}
        </div>
        <button className="btn-primary" onClick={() => navigate('/app/dashboard')}>Back to dashboard</button>
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
        {alreadyAnswered && <p className="muted">Answer submitted — waiting for match to progress.</p>}
      </div>

      <div className="match-nav">
        <button className="btn-ghost" onClick={goNext} disabled={current >= questions.length - 1}>
          Next question
        </button>
      </div>
    </div>
  )
}
