# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /goapi .

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /goapi /goapi
ENV GOAPI_ADDR=:8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/goapi"]
