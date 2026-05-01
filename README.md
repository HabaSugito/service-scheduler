# Service Scheduling & Notification System

A backend API where managers assign Quotes to technicians, with conflict prevention and notifications.

## Stack

- Go 1.25 / Gin / GORM
- MySQL 8.0
- golang-migrate
- Docker + docker-compose

## Setup

```bash
# Start all containers (DB + migration + app)
docker compose up -d --build
```

The app runs at `http://localhost:8080`.

### Local Development (DB in Docker only)

```bash
docker compose up -d db
# Run migrations
docker run --rm -v $(pwd)/db/migrations:/migrations \
  --network host migrate/migrate \
  -path /migrations \
  -database "mysql://schedule:schedule@tcp(localhost:3306)/schedule_db" up

DB_DSN="schedule:schedule@tcp(localhost:3306)/schedule_db?parseTime=true" go run ./cmd/server
```

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/quotes?status=unscheduled` | List unassigned quotes |
| POST | `/jobs` | Create job and assign quote to technician |
| PATCH | `/jobs/:id/complete` | Mark job as completed |
| GET | `/technicians/:id/notifications` | List technician notifications |
| GET | `/managers/:id/notifications` | List manager notifications |

### GET /quotes?status=unscheduled

Returns a list of unassigned quotes.

### POST /jobs

Creates a job and assigns a quote to a technician. `end_at` is automatically set server-side to `start_at + 2 hours`.

**Request body**
```json
{
  "quote_id": 1,
  "technician_id": 1,
  "manager_id": 1,
  "start_at": "2026-05-01T09:00:00Z"
}
```

**Response**
- `201` — Created successfully
- `400` — `start_at` is in the past / validation error
- `404` — Non-existent technician_id / quote_id / manager_id
- `409` — Technician time slot conflict / Quote already assigned

### PATCH /jobs/:id/complete

Marks a job as completed.

### GET /technicians/:id/notifications

Returns a list of notifications for a technician.

### GET /managers/:id/notifications

Returns a list of notifications for a manager.

## Seed Data Example

```sql
INSERT INTO managers (name, email) VALUES ('Alice', 'alice@example.com');
INSERT INTO technicians (name, email) VALUES ('Bob', 'bob@example.com');
INSERT INTO quotes (title, description) VALUES ('Quote A', 'Fix HVAC'), ('Quote B', 'Install outlet');
```

## Conflict Handling Strategy

### Time Overlap Check

```
existing_job.start_at < new_job.end_at AND existing_job.end_at > new_job.start_at
```

This condition detects all overlapping intervals. Two jobs that share only a boundary (one ends exactly when the other starts) are not considered conflicting.

### Why SELECT FOR UPDATE

Checking for conflicts at the application layer alone is insufficient under concurrent requests — both requests may pass the check before either inserts (TOCTOU problem).

```
Request A: conflict check → none → INSERT
Request B: conflict check → none → INSERT  (runs before A's INSERT completes → both pass)
```

By running `SELECT FOR UPDATE` inside a transaction, the first transaction locks the relevant rows and the second waits until the lock is released. The check then re-runs after the lock is released, guaranteeing conflict detection.

### MySQL Constraint Limitation

MySQL does not have a native exclusion constraint for time ranges (unlike PostgreSQL's `EXCLUDE USING`). This system therefore relies on `SELECT FOR UPDATE` combined with an application-layer check.

### Preventing Duplicate Quote Assignment

- DB level: `UNIQUE` constraint on `jobs.quote_id`
- App level: `SELECT FOR UPDATE` on `quote.status = 'unscheduled'` inside a transaction before INSERT

Enforcing at both layers allows the application to return a proper `409` response without relying on a raw DB error.

## Trade-offs

### 1. No Authentication
`manager_id` / `technician_id` are passed in the request body. Production would introduce JWT authentication and derive IDs from the token.

### 2. Notifications Are DB-Only
Only an INSERT into the Notifications table is performed. Production would integrate with SendGrid or a similar service to deliver email and push notifications.

### 3. Why Conflict Prevention Is Enforced at the DB Layer
Application-layer checks alone can be bypassed under concurrent access (TOCTOU problem). Using `SELECT FOR UPDATE` inside a transaction ensures exclusive locking at the DB level.

### 4. Fixed 2-Hour Time Slots
Fixed per task requirements. Production would accept a `duration` field for a flexible design.

### 5. No Soft Delete
Production would manage deletion with a `deleted_at` column.

### 6. Why golang-migrate
Schema changes are tracked as versioned SQL files, making CI/CD integration straightforward. GORM's AutoMigrate is not suitable for production use.
