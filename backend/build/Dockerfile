# -------- STAGE 1: ----------
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o authx ./cmd/server


# ---------- STAGE 2: --------

FROM alpine:latest

WORKDIR /app

# Add non-root user

RUN adduser -D -g '' appuser
USER appuser

COPY --from=builder /app/authx .

COPY .env .env

EXPOSE 8080

ENTRYPOINT [ "./authx" ]
