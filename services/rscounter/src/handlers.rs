//! HTTP handlers and route wiring.

use axum::{
    Json, Router,
    extract::{Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
};
use sqlx::PgPool;

use crate::db;
use crate::error::AppError;
use crate::models::{CounterQuery, CounterResponse, DayCount, EventRequest};

/// Shared state handed to every handler.
#[derive(Clone)]
pub struct AppState {
    pub pool: PgPool,
}

/// Builds the router with all routes registered.
pub fn routes(state: AppState) -> Router {
    Router::new()
        .route("/health", get(health))
        .route("/events", post(record_event))
        .route("/counter", get(get_counter))
        .route("/history", get(get_history))
        .with_state(state)
}

async fn health() -> impl IntoResponse {
    (StatusCode::OK, Json(serde_json::json!({ "status": "ok" })))
}

/// Records a new event for the given day and returns the fresh counts.
async fn record_event(
    State(state): State<AppState>,
    Json(req): Json<EventRequest>,
) -> Result<Json<CounterResponse>, AppError> {
    tracing::info!(date = %req.date, "recording event");
    db::insert_event(&state.pool, req.date).await?;
    let counts = db::load_counts(&state.pool, req.date).await?;
    Ok(Json(counts))
}

/// Reads the current counts for the given day without recording an event.
async fn get_counter(
    State(state): State<AppState>,
    Query(q): Query<CounterQuery>,
) -> Result<Json<CounterResponse>, AppError> {
    tracing::info!(date = %q.date, "reading counter");
    let counts = db::load_counts(&state.pool, q.date).await?;
    Ok(Json(counts))
}

/// Returns the per-day event totals, most recent day first.
async fn get_history(State(state): State<AppState>) -> Result<Json<Vec<DayCount>>, AppError> {
    tracing::info!("reading history");
    let history = db::load_history(&state.pool).await?;
    Ok(Json(history))
}
