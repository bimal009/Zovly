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

### Internal AI service

`go-core` talks to the Python AI service over plain HTTP/JSON:

- `go-core` calls the AI service at `AI_SERVICE_URL` (default `http://localhost:8000`).
- The FastAPI app mounts everything under `/api/v1/ml` (e.g. `/api/v1/ml/chat/reply`,
  `/api/v1/ml/embed/faq`).

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

- **Content Hub** — upload one video → auto-post to TikTok, Instagram, Facebook with AI captions + hashtags per platform.
- **Analytics Engine** — cross-platform performance in one view, with AI insights and ad-budget suggestions.
- **AI Lead Gen** — auto-reply to DMs, comment monitoring → lead capture, AI follow-up sequences.
- **AI Ad Manager** — connect Meta Ads; AI suggests what to boost and optimizes campaigns.
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

See the per-app READMEs in `apps/` for service-specific details, and the shared
packages under `packages/` for the database schema and UI library.

---

## Credits & License

Created by **Bimal Pandey**. All rights reserved.

This project and its code may **not be used, copied, modified, or distributed without
explicit written permission** from Bimal Pandey.
