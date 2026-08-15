# syntax=docker/dockerfile:1

# ---- build stage ----
FROM rust:1.95-slim AS build
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
COPY src ./src
RUN cargo build --release

# ---- runtime stage ----
FROM debian:trixie-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -r -u 10001 appuser
COPY --from=build /app/target/release/rscounter /usr/local/bin/rscounter
USER appuser
ENV RSCOUNTER_ADDR=0.0.0.0:8081
EXPOSE 8081
ENTRYPOINT ["rscounter"]
