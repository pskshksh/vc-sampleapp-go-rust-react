//! Database schema setup and queries.

use chrono::NaiveDate;
use sqlx::{PgPool, Row};

use crate::models::{CounterResponse, DayCount};

/// Column alias used by the count queries.
const COUNT_COLUMN: &str = "n";

const CREATE_EVENTS_TABLE: &str = r#"
    CREATE TABLE IF NOT EXISTS events (
        id         BIGSERIAL PRIMARY KEY,
        event_date DATE        NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    )
"#;

const CREATE_EVENTS_INDEX: &str =
    "CREATE INDEX IF NOT EXISTS idx_events_event_date ON events (event_date)";

const INSERT_EVENT: &str = "INSERT INTO events (event_date) VALUES ($1)";

const COUNT_BY_DATE: &str = "SELECT COUNT(*) AS n FROM events WHERE event_date = $1";

const COUNT_ALL: &str = "SELECT COUNT(*) AS n FROM events";

const HISTORY: &str =
    "SELECT event_date, COUNT(*) AS n FROM events GROUP BY event_date ORDER BY event_date DESC";

/// Lightweight connectivity check for the readiness probe: confirms the pool
/// can round-trip a query to Postgres.
pub async fn ping(pool: &PgPool) -> Result<(), sqlx::Error> {
    sqlx::query("SELECT 1").execute(pool).await?;
    Ok(())
}

/// Ensures the schema exists so the service is runnable on a fresh database.
///
/// Each statement runs separately: Postgres rejects multiple commands in one
/// prepared statement.
pub async fn init_schema(pool: &PgPool) -> Result<(), sqlx::Error> {
    tracing::info!("ensuring database schema exists");
    sqlx::query(CREATE_EVENTS_TABLE).execute(pool).await?;
    sqlx::query(CREATE_EVENTS_INDEX).execute(pool).await?;
    Ok(())
}

/// Records a new event for the given day.
pub async fn insert_event(pool: &PgPool, date: NaiveDate) -> Result<(), sqlx::Error> {
    tracing::debug!(%date, "inserting event");
    sqlx::query(INSERT_EVENT).bind(date).execute(pool).await?;
    Ok(())
}

/// Reads the per-day and all-time counts.
pub async fn load_counts(pool: &PgPool, date: NaiveDate) -> Result<CounterResponse, sqlx::Error> {
    let day_count: i64 = sqlx::query(COUNT_BY_DATE)
        .bind(date)
        .fetch_one(pool)
        .await?
        .get(COUNT_COLUMN);

    let total_count: i64 = sqlx::query(COUNT_ALL).fetch_one(pool).await?.get(COUNT_COLUMN);

    tracing::debug!(%date, day_count, total_count, "loaded counts");
    Ok(CounterResponse {
        date,
        day_count,
        total_count,
    })
}

/// Reads the per-day event totals, most recent day first.
pub async fn load_history(pool: &PgPool) -> Result<Vec<DayCount>, sqlx::Error> {
    let rows = sqlx::query(HISTORY).fetch_all(pool).await?;
    let history: Vec<DayCount> = rows
        .into_iter()
        .map(|row| DayCount {
            date: row.get("event_date"),
            count: row.get(COUNT_COLUMN),
        })
        .collect();

    tracing::debug!(days = history.len(), "loaded history");
    Ok(history)
}
