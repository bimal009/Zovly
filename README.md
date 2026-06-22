# SocialOS (Zovly)

> Universal business platform — **Social Media · AI Lead Gen · Bookings · Inventory · Payments · Analytics**

One platform where a business uploads a video once and posts it everywhere, sees all
its analytics in a single view, connects inventory and payments, and runs AI-powered
lead generation across DMs and comments. The platform asks *what kind of business you
are* and switches on only the modules you need.

---

## Why this exists

No single tool today combines video distribution, inventory sync, AI-driven lead gen,
bookings, and cross-platform analytics — especially one smart enough to detect your
business type and activate only the relevant modules.

| Capability | Best existing tool | The gap it leaves |
| --- | --- | --- |
| Video upload + post to all platforms | Later, Buffer, Hootsuite | No business operations |
| Cross-platform analytics | Sprout Social | No payments / inventory |
| Bookings (salons, gyms, clinics) | Calendly, Vagaro | No social layer |
| Inventory management | Shopify | No social / bookings |
| Payments | GoHighLevel, HoneyBook | No video / social |
| AI DMs / lead gen | ManyChat, Sintra.ai | No inventory / bookings |
| AI ad running | Meta Advantage+ | Single platform only |

---

## Architecture at a glance

```
                         ┌──────────────────┐
   Browser  ──────────►  │  web (Next.js 15)│  :3000   UI, SSR, auth, simple API routes
                         └────────┬─────────┘
                                  │ HTTP (REST + polling)
                         ┌────────▼─────────┐
   Platform webhooks ──► │  go-core (Go/Gin)│  :8080   REST API, webhooks,
   (Meta / TikTok / WA)  └──┬───────┬───────┘          business logic, scheduler, workers
                            │       │
              HTTP/JSON ────┘       └──── TCP
            (AI_SERVICE_URL)              │
                  │                       ▼
          ┌───────▼────────┐      ┌──────────────┐
          │ ai (FastAPI)   │      │ Redis        │  cache · streams · rate-limit
          │ :8000          │      │ (Upstash)    │
          │ RAG, embeddings│      └──────────────┘
          │ Whisper, ML    │      ┌──────────────┐
          └───────┬────────┘      │ Neon Postgres│  app data + pgvector embeddings
                  └───────────────┤ (+ pgvector) │
                                  └──────────────┘
```

### A note on internal transport — FastAPI, not gRPC

The original architecture spec (`SocialOS-Architecture-v3-grpc.pdf`) proposed gRPC over
TCP between `go-core` and the Python ML service. **This project currently uses plain
HTTP/JSON (FastAPI) instead.**

- `go-core` calls the AI service over HTTP at `AI_SERVICE_URL` (default `http://localhost:8000`).
- The FastAPI app mounts everything under `/api/v1/ml` (e.g. `/api/v1/ml/chat/reply`,
  `/api/v1/ml/embed/faq`).
- This keeps the build simple — no `.proto` files, no codegen step, no stub
  regeneration on every change. gRPC remains a possible future migration once
  type-safe contracts or streaming transcription become worth the added build friction.

---

## Monorepo layout

Turborepo + pnpm workspaces.

```
apps/
  web/        # Next.js 15 App Router — UI, auth, simple API routes
  go-core/    # Go + Gin — REST, webhooks, business logic, workers
  ai/         # Python + FastAPI — RAG, embeddings, Whisper transcription, ML
packages/
  database/         # shared DB schema / client
  ui/               # shared React component library (shadcn base)
  eslint-config/    # shared ESLint config
  typescript-config/# shared tsconfig
docker-compose.yaml # Redis + RedisInsight for local dev
```

### `go-core` internals

```
apps/go-core/
  cmd/
    api/main.go              # wires everything, starts the HTTP server
    workers/                 # chat worker, chat retry worker, token refresh worker
  api/routes/                # route registration
  internal/
    config/                  # app config, database, redis
    handler/                 # business, facebook, instagram, faq, imagekit,
                             #   inbox, paddle, plans, product, service, webhook
    service/                 # business logic (chat, instagram, paddle, plans, ...)
    repo/                    # data access
    middlewares/  models/  dto/  constants/
  pkg/
    logger/  responses/  utils/
```

### `ai` (FastAPI) internals

```
apps/ai/
  main.py                    # FastAPI app, mounts /api/v1/ml + /health
  app/
    routes/                  # embed_router, chat_router
    core/chat/               # chat, chunking, embedding, helpers, skills
    models/                  # business_knowledge, chat, faq
    config/db.py             # Postgres / pgvector connection
  pyproject.toml             # uv-managed deps (FastAPI, faster-whisper,
                             #   langchain, sentence-transformers, pgvector, ...)
```

---

## Core feature modules

- **Content Hub** — upload one video → auto-post to TikTok, Instagram, YouTube, Facebook with AI captions + hashtags per platform.
- **Analytics Engine** — cross-platform performance in one view, with AI insights and ad-budget suggestions.
- **AI Lead Gen** — auto-reply to DMs, comment monitoring → lead capture, AI follow-up sequences.
- **AI Ad Manager** — connect Meta + TikTok + Google Ads; AI suggests what to boost and optimizes campaigns.
- **Integrations** — Google Workspace, Paddle payments, WhatsApp Business, and all major social platforms.

Onboarding detects the business type — **Product Seller**, **Service Business**, or
**Both** — and activates only the relevant modules (inventory & checkout vs. bookings &
calendar sync).

---

## How it works

### Webhooks → Redis streams → workers

All inbound platform webhooks land on `go-core` handlers. Payloads are signature-verified
(Meta `X-Hub-Signature-256`, TikTok HMAC, WhatsApp), then pushed to **Redis Streams** for
async processing so the webhook response stays fast and Meta never retries.

Consumer groups give parallel goroutine workers, backpressure, automatic retry of failed
messages (PEL), and replay for debugging.

### AI & RAG pipeline

When a customer DMs a business, the AI service assembles context from several sources —
business persona, knowledge RAG (pgvector), recent conversation history, similar past
chats, and the customer profile — within a token budget, then generates a reply.

- **Vector store:** pgvector on Neon (HNSW index). Qdrant is the documented scale-out path.
- **Multimodal:** images are passed straight to the model (no separate OCR); voice notes
  are transcribed (faster-whisper) and stored as a browser-playable file for
  dashboard replay.

### Live updates (polling)

The dashboard stays current by **polling** `go-core` REST endpoints on an interval — new
messages, leads, published posts, analytics updates, bookings, and payments are picked up
on the next poll. (WebSockets are a possible future upgrade for lower-latency push.)

---

## Getting started

### Prerequisites

- Node.js ≥ 18 and **pnpm** 9
- Go (see `apps/go-core/go.mod` for the toolchain version)
- Python ≥ 3.11 and [**uv**](https://docs.astral.sh/uv/)
- Docker (for Redis) and a Neon/Postgres database with the `vector` extension

### 1. Install workspace deps

```sh
pnpm install
```

### 2. Redis

This project uses **Upstash** (serverless Redis) — point `REDIS_URL` at your Upstash
database (TLS, `rediss://`). For fully-local dev you can instead run the bundled Redis +
RedisInsight:

```sh
docker compose up -d
# Redis on :6379, RedisInsight UI on http://localhost:5540
```

### 3. Run the apps

**web + any TS packages (Turborepo):**

```sh
pnpm dev                      # all turbo dev tasks
pnpm dev --filter=web         # just the web app
```

**go-core (API + workers):**

```sh
cd apps/go-core
go run ./cmd/api              # HTTP API on :8080
go run ./cmd/workers          # background workers (chat, retries, token refresh)
```

**ai (FastAPI):**

```sh
cd apps/ai
uv sync
uv run uvicorn main:app --reload --port 8000
# health check: http://localhost:8000/api/v1/health
```

> `go-core` reaches the AI service via `AI_SERVICE_URL` (defaults to `http://localhost:8000`).

---

## Environment variables

Each app reads its own `.env`. Never commit secrets — keep a `.env.example` with empty
values. Key variables:

```sh
# Database / cache
DATABASE_URL=postgresql://user:pass@ep-xxx.neon.tech/neondb?sslmode=require
REDIS_URL=rediss://default:pass@xxx.upstash.io:6379   # Upstash (TLS)

# Internal services
AI_SERVICE_URL=http://localhost:8000        # go-core → FastAPI AI service
JWT_SECRET=your-256-bit-secret

# Media (ImageKit)
IMAGEKIT_PUBLIC_KEY=
IMAGEKIT_PRIVATE_KEY=
IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/your_imagekit_id

# Meta (Instagram + Facebook)
META_APP_ID=
META_APP_SECRET=
META_WEBHOOK_VERIFY_TOKEN=

# WhatsApp
WA_PHONE_NUMBER_ID=
WA_ACCESS_TOKEN=
WA_WEBHOOK_VERIFY_TOKEN=

# TikTok
TIKTOK_CLIENT_KEY=
TIKTOK_CLIENT_SECRET=
TIKTOK_WEBHOOK_SECRET=

# Google Workspace
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URI=

# Payments (Paddle)
PADDLE_API_KEY=
PADDLE_WEBHOOK_SECRET=

# AI
ANTHROPIC_API_KEY=

# Web (public)
NEXT_PUBLIC_API_URL=
```

---

## Tech stack

| Layer | Technology |
| --- | --- |
| Frontend | Next.js 15 (App Router), React, Tailwind, shadcn |
| Backend | Go + Gin — REST, webhooks, workers |
| AI / ML | Python + FastAPI — RAG, embeddings, Whisper transcription |
| Database | Neon Postgres + pgvector |
| Cache / queue | Redis via Upstash (cache, streams, rate limiting) |
| Object storage / media | ImageKit (upload, optimization, signed URLs) |
| Payments | Paddle |
| AI model | Claude (Anthropic) |
| Monorepo | Turborepo + pnpm |

---

## Reference

The full design rationale — database schema, platform API capabilities, media pipeline,
onboarding flow, and the phased execution roadmap — lives in
[`SocialOS-Architecture-v3-grpc.pdf`](./SocialOS-Architecture-v3-grpc.pdf).

> ⚠️ That document specifies **gRPC** for internal `go-core ↔ ai` communication. The
> current implementation uses **HTTP/JSON (FastAPI)** instead — see
> [the internal transport note](#a-note-on-internal-transport--fastapi-not-grpc) above.
