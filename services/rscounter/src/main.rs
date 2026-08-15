//! rscounter — Rust counter service backed by Postgres.
//!
//! Records one event row per hit and reports the count for a given day plus the
//! all-time total. Consumed by `goapi`.

mod config;
mod db;
mod error;
mod handlers;
mod models;

use sqlx::postgres::PgPoolOptions;

use crate::config::{Config, DEFAULT_LOG_FILTER, MAX_CONNECTIONS};
use crate::handlers::AppState;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    init_tracing();

    let cfg = Config::from_env();
    tracing::info!(addr = %cfg.addr, "starting rscounter");

    let pool = PgPoolOptions::new()
        .max_connections(MAX_CONNECTIONS)
        .connect(&cfg.database_url)
        .await?;

    db::init_schema(&pool).await?;

    let app = handlers::routes(AppState { pool: pool.clone() });

    let listener = tokio::net::TcpListener::bind(&cfg.addr).await?;
    tracing::info!(addr = %cfg.addr, "rscounter listening");

    // Serve until a shutdown signal arrives, then let in-flight requests finish.
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    // Close the connection pool cleanly before exiting.
    pool.close().await;
    tracing::info!("rscounter stopped");

    Ok(())
}

/// Completes when the process receives SIGINT (Ctrl-C) or SIGTERM.
async fn shutdown_signal() {
    let ctrl_c = async {
        tokio::signal::ctrl_c()
            .await
            .expect("failed to install Ctrl-C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
            .expect("failed to install SIGTERM handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }

    tracing::info!("shutdown signal received, shutting down gracefully");
}

fn init_tracing() {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| DEFAULT_LOG_FILTER.into()),
        )
        .init();
}
