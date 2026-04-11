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

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Install radon (pinned version for score continuity)
RUN pip install --break-system-packages radon==6.0.1

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

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/proving-ground", "--data-dir", "/data"]
CMD []
