import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { getSubjects, getUserStats } from '../../api/endpoints'

export default function Dashboard() {
  const { user } = useAuth()
  const [subjects, setSubjects] = useState([])
  const [stats, setStats] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [{ data: subs }, { data: st }] = await Promise.all([
          getSubjects(),
          getUserStats(user.id),
        ])
        setSubjects(subs)
        setStats(st)
      } catch (err) {
        setError(err.message)
      }
    }
    load()
  }, [user.id])

  return (
    <div className="page">
      <h1>Welcome back, {user.display_name || user.username}</h1>
      {error && <div className="alert-error">{error}</div>}

      <div className="card-grid">
        <Link to="/app/matchmaking" className="action-card">
          <h3>Ranked Match</h3>
          <p>Queue up and get paired with a similarly-rated opponent.</p>
        </Link>
        <Link to="/app/friend" className="action-card">
          <h3>Friend Match</h3>
          <p>Create or join a private room with a room code.</p>
        </Link>
        <Link to="/app/practice" className="action-card">
          <h3>Practice</h3>
          <p>Solve questions solo with instant feedback. No rating impact.</p>
        </Link>
        <Link to="/app/leaderboard" className="action-card">
          <h3>Leaderboard</h3>
          <p>See how you stack up by exam category.</p>
        </Link>
      </div>

      <h2>Exam categories</h2>
      <div className="chip-row">
        {subjects.map((s) => (
          <span key={s.id} className="chip" title={s.description}>
            {s.name}
          </span>
        ))}
      </div>

      <h2>Your stats</h2>
      {stats && stats.length === 0 && <p className="muted">No stats yet — play a match or practice session to get started.</p>}
      <div className="stats-grid">
        {stats && stats.map((s) => (
          <div key={s.exam_category_id} className="stat-card">
            <div className="stat-row"><span>Matches</span><strong>{s.total_matches}</strong></div>
            <div className="stat-row"><span>W / L / D</span><strong>{s.wins} / {s.losses} / {s.draws}</strong></div>
            <div className="stat-row"><span>Accuracy</span><strong>{s.overall_accuracy.toFixed(1)}%</strong></div>
            <div className="stat-row"><span>Win streak</span><strong>{s.current_win_streak}</strong></div>
          </div>
        ))}
      </div>
    </div>
  )
}
