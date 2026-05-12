# Babu Personal OS Backend

Go API for the Flutter Personal AI Operating System.

Free-first architecture:

```txt
Flutter Mobile App
        ↓
Go Backend on Render Free
        ↓
Neon PostgreSQL Free
        ↓
Gemini/Groq AI Free API
        ↓
Firebase FCM for notifications
```

## Run Locally

```bash
cd ../profile-backend
ADMIN_TOKEN=change-me-before-deploy go run ./cmd/api
```

The API starts on `http://localhost:8080`.

The default local storage driver is `file`, so you can start without MongoDB or PostgreSQL. For Neon/Render deployment set `STORAGE_DRIVER=postgres`.

```bash
STORAGE_DRIVER=file ADMIN_TOKEN=change-me-before-deploy go run ./cmd/api
```

## Configuration

Copy `.env.example` into your hosting provider or local shell.

```bash
PORT=8080
ADMIN_TOKEN=change-me-before-deploy
ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000

STORAGE_DRIVER=postgres
DATABASE_URL=postgresql://user:password@host/dbname?sslmode=require
AUTO_MIGRATE=true

AI_PROVIDER=gemini
AI_API_KEY=your-free-tier-key
AI_MODEL=gemini-1.5-flash

FCM_SERVER_KEY=your-firebase-key
```

Use `AI_PROVIDER=groq` with `AI_MODEL=llama-3.1-8b-instant` when you want Groq instead.

## PostgreSQL

The backend supports `STORAGE_DRIVER=postgres`, `mongo`, and `file`.

For Neon:

1. Create a free Neon database.
2. Copy the pooled connection string.
3. Set it as `DATABASE_URL`.
4. Keep `sslmode=require`.
5. Leave `AUTO_MIGRATE=true` for hobby/free deployment.

The schema is in `db/migrations/0001_personal_os_postgres.sql`; the app also runs the same migration automatically on startup.

## Render Free

This repo includes:

- `Dockerfile`
- `render.yaml`

On Render, create a Blueprint or Web Service and set these secrets:

- `ADMIN_TOKEN`
- `DATABASE_URL`
- `AI_API_KEY`
- `FCM_SERVER_KEY` when notifications are ready

Free Render instances may sleep, so first requests can be slow.

## Public Endpoints

- `GET /health`
- `GET /api/profile`
- `POST /api/contact`

## Private Endpoints

Send `Authorization: Bearer <ADMIN_TOKEN>`.

- `GET /api/dashboard`
- `GET /api/contact/messages`
- `PATCH /api/contact/messages/{id}` with `{ "status": "read" }`
- `DELETE /api/contact/messages/{id}`
- `GET /api/personal/summary`
- `GET|POST /api/personal/notes`
- `DELETE /api/personal/notes/{id}`
- `GET|POST /api/personal/reminders`
- `PATCH /api/personal/reminders/{id}/toggle`
- `DELETE /api/personal/reminders/{id}`
- `GET|POST /api/personal/transactions`
- `DELETE /api/personal/transactions/{id}`
- `GET|POST /api/personal/habits`
- `PATCH /api/personal/habits/{id}/check-in`
- `DELETE /api/personal/habits/{id}`
- `POST /api/ai/daily-briefing`
- `POST /api/ai/note-summary`
- `GET /api/notifications/status`
- `POST /api/notifications/test`

## Notes

Drizzle is a TypeScript ORM, so the Go backend uses native PostgreSQL SQL migrations plus `pgx`. If a future Next.js/Admin app is added, it can use Drizzle against the same schema.
