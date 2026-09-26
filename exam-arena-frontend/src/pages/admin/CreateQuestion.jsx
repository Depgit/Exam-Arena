import { useEffect, useState } from 'react'
import { getSubjects, createQuestion, publishQuestion } from '../../api/endpoints'

const emptyOption = () => ({ option_text: '', is_correct: false })

export default function CreateQuestion() {
  const [subjects, setSubjects] = useState([])
  const [categoryId, setCategoryId] = useState('')
  const [topicId, setTopicId] = useState('')
  const [questionType, setQuestionType] = useState('mcq_single')
  const [difficulty, setDifficulty] = useState('medium')
  const [body, setBody] = useState('')
  const [explanation, setExplanation] = useState('')
  const [estimatedTime, setEstimatedTime] = useState(60)
  const [options, setOptions] = useState([emptyOption(), emptyOption(), emptyOption(), emptyOption()])
  const [error, setError] = useState('')
  const [created, setCreated] = useState(null)
  const [publishing, setPublishing] = useState(false)
  const [published, setPublished] = useState(false)

  useEffect(() => {
    getSubjects().then(({ data }) => {
      setSubjects(data)
      if (data.length) setCategoryId(data[0].id)
    })
  }, [])

  function updateOption(index, patch) {
    setOptions((prev) => prev.map((o, i) => (i === index ? { ...o, ...patch } : o)))
  }

  function toggleCorrect(index) {
    if (questionType === 'mcq_single') {
      setOptions((prev) => prev.map((o, i) => ({ ...o, is_correct: i === index })))
    } else {
      updateOption(index, { is_correct: !options[index].is_correct })
    }
  }

  function addOption() {
    setOptions((prev) => [...prev, emptyOption()])
  }

  function removeOption(index) {
    setOptions((prev) => prev.filter((_, i) => i !== index))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setCreated(null)
    setPublished(false)

    const cleanOptions = options.filter((o) => o.option_text.trim() !== '')
    if (cleanOptions.length < 2) {
      setError('At least 2 non-empty options are required.')
      return
    }
    if (!cleanOptions.some((o) => o.is_correct)) {
      setError('Mark at least one option as correct.')
      return
    }

    try {
      const { data } = await createQuestion({
        exam_category_id: categoryId,
        topic_id: topicId,
        question_type: questionType,
        difficulty,
        body,
        explanation: explanation || null,
        estimated_time_seconds: Number(estimatedTime),
        options: cleanOptions,
      })
      setCreated(data)
    } catch (err) {
      setError(err.message)
    }
  }

  async function handlePublish() {
    setPublishing(true)
    setError('')
    try {
      await publishQuestion(created.id)
      setPublished(true)
    } catch (err) {
      setError(err.message)
    } finally {
      setPublishing(false)
    }
  }

  function handleAddAnother() {
    setCreated(null)
    setPublished(false)
    setBody('')
    setExplanation('')
    setOptions([emptyOption(), emptyOption(), emptyOption(), emptyOption()])
  }

  return (
    <div className="page">
      <h1>Create Question</h1>
      {error && <div className="alert-error">{error}</div>}

      {created ? (
        <div className="form-card">
          <div className="alert-success">Question created (status: draft).</div>
          <p className="question-body">{created.body}</p>
          {!published ? (
            <button className="btn-primary" onClick={handlePublish} disabled={publishing}>
              {publishing ? 'Publishing…' : 'Publish now'}
            </button>
          ) : (
            <div className="alert-success">
              Published — it will appear in the live question bank within 15 minutes.
            </div>
          )}
          <button className="btn-ghost" onClick={handleAddAnother}>Create another question</button>
        </div>
      ) : (
        <form className="form-card" onSubmit={handleSubmit}>
          <label>
            Exam category
            <select value={categoryId} onChange={(e) => setCategoryId(e.target.value)} required>
              {subjects.map((s) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
          </label>
          <label>
            Topic ID (UUID)
            <input
              value={topicId}
              onChange={(e) => setTopicId(e.target.value)}
              placeholder="e.g. topic-uuid — no topics-listing endpoint exists yet"
              required
            />
          </label>
          <label>
            Question type
            <select value={questionType} onChange={(e) => setQuestionType(e.target.value)}>
              <option value="mcq_single">Single-choice MCQ</option>
              <option value="mcq_multiple">Multi-choice MCQ</option>
              <option value="integer">Integer answer</option>
            </select>
          </label>
          <label>
            Difficulty
            <select value={difficulty} onChange={(e) => setDifficulty(e.target.value)} required>
              <option value="easy">Easy</option>
              <option value="medium">Medium</option>
              <option value="hard">Hard</option>
            </select>
          </label>
          <label>
            Question body
            <textarea value={body} onChange={(e) => setBody(e.target.value)} required rows={3} />
          </label>
          <label>
            Explanation (optional)
            <textarea value={explanation} onChange={(e) => setExplanation(e.target.value)} rows={2} />
          </label>
          <label>
            Estimated time (seconds)
            <input type="number" min={5} value={estimatedTime} onChange={(e) => setEstimatedTime(e.target.value)} />
          </label>

          <h3>Options</h3>
          {options.map((opt, i) => (
            <div className="option-editor-row" key={i}>
              <input
                type={questionType === 'mcq_multiple' ? 'checkbox' : 'radio'}
                name="correct-option"
                checked={opt.is_correct}
                onChange={() => toggleCorrect(i)}
                title="Mark as correct"
              />
              <input
                value={opt.option_text}
                onChange={(e) => updateOption(i, { option_text: e.target.value })}
                placeholder={`Option ${i + 1}`}
              />
              {options.length > 2 && (
                <button type="button" className="btn-ghost small" onClick={() => removeOption(i)}>✕</button>
              )}
            </div>
          ))}
          <button type="button" className="btn-ghost" onClick={addOption}>+ Add option</button>

          <button className="btn-primary" type="submit">Create question</button>
        </form>
      )}
    </div>
  )
}
