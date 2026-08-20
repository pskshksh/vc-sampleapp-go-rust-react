import { useEffect, useState } from 'react'
import { formatDate } from './format'
import './App.css'

// Shape returned by goapi's GET /api/today.
type Today = {
  date: string
  day_count: number
  total_count: number
}

// One row from goapi's GET /api/history.
type DayCount = {
  date: string
  count: number
}

function App() {
  const [today, setToday] = useState<Today | null>(null)
  const [history, setHistory] = useState<DayCount[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    // Record today's visit, then load the per-day history.
    fetch('/api/today')
      .then((res) => {
        if (!res.ok) throw new Error(`goapi returned ${res.status}`)
        return res.json() as Promise<Today>
      })
      .then((data) => {
        setToday(data)
        return fetch('/api/history')
      })
      .then((res) => (res.ok ? (res.json() as Promise<DayCount[]>) : []))
      .then(setHistory)
      .catch((err: unknown) => setError(err instanceof Error ? err.message : String(err)))
  }, [])

  return (
    <main className="app">
      <h1>Today</h1>

      {error && <p className="error">Could not reach goapi: {error}</p>}

      {!error && !today && <p className="muted">Loading…</p>}

      {today && (
        <>
          <p className="date">{formatDate(today.date)}</p>
          <div className="counters">
            <div className="counter">
              <span className="value">{today.day_count}</span>
              <span className="label">visits today</span>
            </div>
            <div className="counter">
              <span className="value">{today.total_count}</span>
              <span className="label">visits all-time</span>
            </div>
          </div>

          {history.length > 0 && (
            <section className="history">
              <h2>History</h2>
              <ul>
                {history.map((day) => (
                  <li key={day.date}>
                    <span>{formatDate(day.date)}</span>
                    <span className="count">{day.count}</span>
                  </li>
                ))}
              </ul>
            </section>
          )}
        </>
      )}
    </main>
  )
}

export default App
