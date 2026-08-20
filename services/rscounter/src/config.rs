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

#[cfg(test)]
mod tests {
    use super::*;

    // These tests mutate process-wide environment variables, so keep them in a
    // single test to avoid races between parallel test threads.
    #[test]
    fn from_env_reads_overrides_and_defaults() {
        // SAFETY: single-threaded within this test; no other test touches these
        // variables concurrently.
        unsafe {
            env::set_var(ENV_DATABASE_URL, "postgres://u:p@db:5432/app");
            env::remove_var(ENV_ADDR);
        }

        let cfg = Config::from_env();
        assert_eq!(cfg.database_url, "postgres://u:p@db:5432/app");
        // ENV_ADDR is unset, so the documented default applies.
        assert_eq!(cfg.addr, DEFAULT_ADDR);

        unsafe {
            env::remove_var(ENV_DATABASE_URL);
        }
        let cfg = Config::from_env();
        assert_eq!(cfg.database_url, DEFAULT_DATABASE_URL);
    }

    #[test]
    fn defaults_are_well_formed() {
        assert!(DEFAULT_DATABASE_URL.starts_with("postgres://"));
        assert!(DEFAULT_ADDR.contains(':'));
        assert!(MAX_CONNECTIONS > 0);
    }
}
