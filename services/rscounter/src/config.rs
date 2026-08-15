//! Runtime configuration, sourced from environment variables.

use std::env;

/// Environment variable names.
const ENV_DATABASE_URL: &str = "DATABASE_URL";
const ENV_ADDR: &str = "RSCOUNTER_ADDR";

/// Defaults used when the corresponding environment variable is unset.
const DEFAULT_DATABASE_URL: &str = "postgres://rscounter:rscounter@localhost:5432/rscounter";
const DEFAULT_ADDR: &str = "0.0.0.0:8081";

/// Maximum number of pooled Postgres connections.
pub const MAX_CONNECTIONS: u32 = 5;

/// Fallback tracing filter when `RUST_LOG` is not set.
pub const DEFAULT_LOG_FILTER: &str = "info,sqlx=warn";

#[derive(Clone, Debug)]
pub struct Config {
    pub database_url: String,
    pub addr: String,
}

impl Config {
    /// Reads configuration from the environment, applying defaults where unset.
    pub fn from_env() -> Self {
        Self {
            database_url: env::var(ENV_DATABASE_URL)
                .unwrap_or_else(|_| DEFAULT_DATABASE_URL.to_string()),
            addr: env::var(ENV_ADDR).unwrap_or_else(|_| DEFAULT_ADDR.to_string()),
        }
    }
}
