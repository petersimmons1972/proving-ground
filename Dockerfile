FROM python:3.12-slim

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    git \
    nodejs \
    npm \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Set up application
WORKDIR /app

# Copy source first so the editable install can resolve the src package
COPY pyproject.toml .
COPY src/ src/
COPY profiles/ profiles/
COPY tasks/ tasks/
COPY task-fixtures/ task-fixtures/
COPY templates/ templates/
COPY prompts/ prompts/

RUN pip install --no-cache-dir -e .

# Data volume — user profiles, results, history
VOLUME ["/data"]

# Default entrypoint
ENTRYPOINT ["proving-ground", "--data-dir", "/data"]
CMD []
