FROM node:24-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev --no-audit --no-fund

FROM node:24-alpine
ENV NODE_ENV=production COACH_WORKDIR=/tmp/coach CLAUDE_CONFIG_DIR=/data/claude
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY package.json ./
COPY src ./src
# Agent SDK session transcripts (needed for `resume`) live in CLAUDE_CONFIG_DIR -> mount a volume there.
RUN mkdir -p /data/claude /tmp/coach && chown -R node /data/claude /tmp/coach
USER node
EXPOSE 8090
CMD ["node", "src/main.ts"]
