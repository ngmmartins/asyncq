# Async Job Scheduler – MVP Launch Checklist

## 1. Core Functionality
- [x] Create job (with optional `run_at`)
- [x] Retrieve job by ID
- [x] View job status
- [x] Search jobs
- [x] Cancel job (before running)
- [x] Schedule created job (deferred)
- [x] Worker picks and executes queued jobs
- [x] Retry logic with delay and max attempts
- [ ] Email sending (basic text support)

---

## 2. Security & Access Control
- [x] API key or token-based authentication (per user/account/team)
- [ ] Enforce rate limits and quota (basic tier, fair usage)
- [ ] Isolate job data between clients (multi-tenant ready or scoped access)

---

## 3. Developer Experience
- [ ] Public API documentation (e.g., Swagger/OpenAPI or markdown)
- [ ] Error responses are consistent (e.g., structured JSON: `status`, `message`, `code`)
- [ ] Sample code or Postman collection
- [ ] Clear request/response examples

---

## 4. Email Delivery
- [x] Basic SMTP configuration via env vars
- [ ] Better SMTP failure diagnostics (`invalid address`, `unreachable host`, `auth failed`)
- [ ] Optional: support custom SMTP per user/account
- [ ] Optional: plan for HTML + attachments

---

## 5. Operations & Monitoring
- [ ] Logging with proper levels (info, debug, error)
- [ ] Graceful shutdown (cancel job or wait until finished)
- [ ] Health check endpoints (`/health`)
- [ ] Worker tick interval is configurable
- [ ] (Optional) Metrics: number of jobs processed, retries, failures

---

## 6. Product Readiness
- [ ] Define plans (free, basic, pro?) or usage limits
- [ ] Terms of service & privacy policy (basic versions)
- [ ] Landing page or marketing site (even minimal)
- [ ] Logo / name / favicon (placeholder ok)
- [ ] Domain name registered

---

## 7. Deployment
- [ ] Dockerfile + docker-compose for local deployment
- [ ] Deployment to cloud/VPS (e.g., Fly.io, Railway, Hetzner, etc.)
- [ ] Env-var driven config (DB, Redis, SMTP, secrets)
- [ ] PostgreSQL & Redis provisioning
- [ ] TLS / HTTPS enabled

---

## 8. Launch Announcement Prep
- [ ] Launch post (LinkedIn, Reddit, Twitter, Hacker News, etc.)
- [ ] Explain use case clearly: “Schedule and retry async jobs easily”
- [ ] Examples or demos (GIF or screenshot or screencast)
- [ ] Post on platforms (e.g., Product Hunt, IndieHackers, Dev.to)
