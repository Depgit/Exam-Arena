import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { getSubjects, getUserStats, getLeaderboard } from '../../api/endpoints'

export default function Dashboard() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [subjects, setSubjects] = useState([])
  const [stats, setStats] = useState([])
  const [leaderboard, setLeaderboard] = useState([])
  const [lbCategory, setLbCategory] = useState(null)
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
        // Load leaderboard for the first available subject
        if (subs.length > 0) {
          setLbCategory(subs[0])
          const { data: lb } = await getLeaderboard(subs[0].code, { limit: 10 })
          setLeaderboard(lb)
        }
      } catch (err) {
        setError(err.message)
      }
    }
    load()
  }, [user.id])

  async function switchLeaderboard(sub) {
    setLbCategory(sub)
    try {
      const { data: lb } = await getLeaderboard(sub.code, { limit: 10 })
      setLeaderboard(lb)
    } catch {
      // ignore; old data stays
    }
  }

  return (
    <div className="page dashboard-page">
      <h1>Welcome back, {user.display_name || user.username} 👋</h1>
      {error && <div className="alert-error">{error}</div>}

      <div className="dashboard-layout">
        {/* ── Left / Main column ────────────────────────────────── */}
        <div className="dashboard-main">
          <div className="card-grid">
            <Link to="/app/matchmaking" className="action-card ranked">
              <div className="action-icon">⚔️</div>
              <h3>Ranked Match</h3>
              <p>Queue up and get paired with a similarly-rated opponent.</p>
            </Link>
            <Link to="/app/friend" className="action-card friend">
              <div className="action-icon">👥</div>
              <h3>Friend Match</h3>
              <p>Create or join a private room with a room code.</p>
            </Link>
            <Link to="/app/practice" className="action-card practice">
              <div className="action-icon">🎯</div>
              <h3>Practice</h3>
              <p>Solve questions solo with instant feedback. No rating impact.</p>
            </Link>
            <Link to="/app/leaderboard" className="action-card leaderboard-card">
              <div className="action-icon">🏆</div>
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
          {stats && stats.length === 0 && (
            <p className="muted">No stats yet — play a match or practice session to get started.</p>
          )}
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

        {/* ── Right column: Top 10 Leaderboard ─────────────────── */}
        <aside className="dashboard-sidebar">
          <div className="lb-widget">
            <div className="lb-widget-header">
              <h3>🏆 Top 10</h3>
              {lbCategory && (
                <Link to="/app/leaderboard" className="lb-see-all">See all →</Link>
              )}
            </div>
            {subjects.length > 1 && (
              <div className="lb-tabs">
                {subjects.slice(0, 4).map((s) => (
                  <button
                    key={s.id}
                    className={`lb-tab ${lbCategory?.id === s.id ? 'active' : ''}`}
                    onClick={() => switchLeaderboard(s)}
                  >
                    {s.name.length > 8 ? s.name.slice(0, 8) + '…' : s.name}
                  </button>
                ))}
              </div>
            )}
            <ol className="lb-list">
              {leaderboard.map((entry, idx) => {
                const isMe = entry.user_id === user.id
                const medal = idx === 0 ? '🥇' : idx === 1 ? '🥈' : idx === 2 ? '🥉' : null
                return (
                  <li
                    key={entry.user_id}
                    className={`lb-item ${isMe ? 'lb-me' : ''}`}
                    onClick={() => navigate(`/app/profile`)}
                    title={`View ${entry.username}'s profile`}
                    style={{ cursor: 'pointer' }}
                  >
                    <span className="lb-rank">{medal || `#${idx + 1}`}</span>
                    <span className="lb-name">{entry.display_name || entry.username}{isMe ? ' ★' : ''}</span>
                    <span className="lb-rating">{entry.rating}</span>
                  </li>
                )
              })}
              {leaderboard.length === 0 && (
                <li className="lb-empty muted">No players yet</li>
              )}
            </ol>
          </div>
        </aside>
      </div>
    </div>
  )
}
