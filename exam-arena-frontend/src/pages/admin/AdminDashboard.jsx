import { Link } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

export default function AdminDashboard() {
  const { user } = useAuth()
  return (
    <div className="page">
      <h1>Admin overview</h1>
      <p className="muted">Signed in as {user.username} (admin)</p>
      <div className="card-grid">
        <Link to="/admin/questions/new" className="action-card">
          <h3>Create Question</h3>
          <p>Add a new question and publish it to the live question bank.</p>
        </Link>
        <Link to="/admin/stats" className="action-card">
          <h3>System Stats</h3>
          <p>Users, questions, and match volume at a glance.</p>
        </Link>
      </div>
    </div>
  )
}
