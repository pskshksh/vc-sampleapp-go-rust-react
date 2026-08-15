//! Request and response payloads exchanged over HTTP.

use chrono::NaiveDate;
use serde::{Deserialize, Serialize};

/// Body of `POST /events`.
#[derive(Debug, Deserialize)]
pub struct EventRequest {
    /// Day the event belongs to, formatted `YYYY-MM-DD`.
    pub date: NaiveDate,
}

/// Query string of `GET /counter`.
#[derive(Debug, Deserialize)]
pub struct CounterQuery {
    pub date: NaiveDate,
}

/// Counts returned by the event/counter endpoints.
#[derive(Debug, Serialize)]
pub struct CounterResponse {
    pub date: NaiveDate,
    pub day_count: i64,
    pub total_count: i64,
}

/// One day's event total, returned by the history endpoint.
#[derive(Debug, Serialize)]
pub struct DayCount {
    pub date: NaiveDate,
    pub count: i64,
}
