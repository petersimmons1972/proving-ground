# Stage 1: Build Go binary
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/proving-ground ./cmd/proving-ground

# Stage 2: Runtime image
FROM alpine:3.19

# System deps: node for claude CLI, python3+pip for radon, git for task fixtures
RUN apk add --no-cache \
    nodejs \
    npm \
    python3 \
    py3-pip \
    git \
    ca-certificates

# Install Claude Code CLI — pinned for reproducible benchmark runs (#26)
RUN npm install -g @anthropic-ai/claude-code@1.0.3

# Install radon (pinned version for score continuity) and pytest-json-report
# for harness-controlled test scoring (B-3).
RUN pip install --break-system-packages radon==6.0.1 pytest pytest-json-report

# Copy Go binary
COPY --from=builder /out/proving-ground /usr/local/bin/proving-ground

# Copy language-agnostic benchmark content
COPY tasks/ /app/tasks/
COPY profiles/ /app/profiles/
COPY prompts/ /app/prompts/
COPY task-fixtures/ /app/task-fixtures/
COPY templates/ /app/templates/

WORKDIR /app

# Data volume — user profiles, results, history
VOLUME ["/data"]

# Non-root user — eliminates root execution of agent subprocesses inside
# the container (B-1, #18). The /data directory is pre-owned so the volume
# mount does not require root at runtime.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup && mkdir -p /data && chown appuser:appgroup /data

USER appuser

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/proving-ground", "--data-dir", "/data"]
CMD []
