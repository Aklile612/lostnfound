# Reunite — campus lost & found matcher

A small university lost-and-found app. Students report lost or found items. The app filters impossible pairs in the database, then Groq and Gemini score how likely two reports describe the same object.

No accounts. Open [http://localhost:3000](http://localhost:3000) after Docker is up.

## Run

```bash
cp .env.example .env
```

Put your keys in `.env` if you have them:

```
GEMINI_API_KEY=...
GROQ_API_KEY=...
```

The app still runs without keys. Matching then uses a local heuristic so you can demo the flow.

```bash
docker compose up --build
```

- Web: http://localhost:3000
- API: http://localhost:8080/health

That is the only required setup.

## Approach

The assignment asks for lost reports, found reports, and potential matches. The interesting part is what a “match” means.

I treated matching as two stages:

1. **Hard filter (SQL).** Drop pairs that cannot be the same item.
2. **Soft score (AI, then a fallback).** Rank the remaining pairs and explain why.

Finders upload a photo (or a short description), the place they found it, the date, a phone number, and a Telegram username. Owners describe the item, where they lost it, the date, optional unique marks (a crack, a sticker), and an optional photo. After an owner submits, they see ranked matches and can contact the finder.

## How matching works

### 1. Metadata filter

Before any model runs, Postgres keeps only found items that are:

- the same **category** (electronics, bags, IDs, ATM cards, keys, clothing, jewelry, books, other)
- **unclaimed**
- found **on or after** the lost date
- found **within 7 days after** the lost date

If someone found a backpack at the football field two weeks later, it never reaches the models. Same if the categories differ, or if the item was already claimed.

The reverse query is used when a found report is submitted: open lost reports in that category whose lost date is in the 7-day window before the found date.

### 2. Groq (text)

Groq (`llama-3.3-70b-versatile`) only sees text: descriptions, unique marks, campus place, extra location detail, dates, and whether photos exist. It cannot see pixels. That is intentional. It is good at contradictory brands/colors, nearby places (cafeteria vs coffee shop), and distinctive marks.

### 3. Gemini (vision + text)

Gemini (`gemini-2.0-flash`) gets the same text plus up to two photos per report. A finder can submit only a photo and a place. An owner can submit only a description. Gemini is told not to punish missing text on one side, and to treat a photo that clearly shows a different object as a rejection.

### 4. Combined score

If both models return a score: `0.45 * Groq + 0.55 * Gemini` (vision slightly heavier when photos exist). If only one model runs, that score is used. Matches below **40** are dropped.

### 5. Fallback

If both API keys are missing, or both calls fail, a local scorer uses location proximity, token overlap (including unique marks), and date closeness. The UI labels that as a heuristic so it is not mistaken for model output.

Each kept match stores Groq score, Gemini score, combined score, and a short reason.

## Important assumptions

- This is a campus, so places are a fixed list plus an optional free-text detail.
- “Found before it was lost” is impossible; “found more than a week later” is too weak to show by default. The 7-day window is a product choice, not a universal truth.
- Generic text like “black bag” should not rank highly. Shared unique marks should.
- Nearby campus places can still match (cafeteria / coffee shop). Distant ones should rarely match unless the rest of the evidence is strong.
- No login. Finders leave contact details on the found report. Owners get a shareable `/r/{id}` link to check again later.
- Claiming is intentionally weak (anyone with the link can mark a found item claimed). Fine for a 3-hour exercise, not for production.
- Photos are stored on disk, not in object storage.

## Technical decisions

- **Go** for the API: Gin, pgx.
- **Next.js** for the UI: two forms and a report page.
- **Frontend design:** TypeUI [Brutalism](https://www.typeui.sh/design-skills/brutalism) (`npx typeui.sh pull brutalism`). The skill is in `.cursor/skills/design-system/SKILL.md`. Tokens from that system: white surface (`#FFFFFF`), near-black text (`#111827`), coral primary (`#DD614C`), gold secondary (`#DAA144`), Darker Grotesque + JetBrains Mono, thick black borders, and hard offset shadows. Layout is Tailwind CSS.
- **Postgres** so the hard filter is a real query, not a loop in memory.
- **Clean architecture** in the backend: `domain` → `port` → `repository` / `ai` → `service` → `handler`. Matching rules stay in the service; HTTP does not know about Groq or Gemini.
- **Docker Compose** so a reviewer can run everything with one command.
- **Both models, on purpose.** Groq is fast text. Gemini can see photos. Using only one would ignore either language or appearance.

## What I did not build

- Accounts, email, or notifications when a new found item arrives later
- A staff inbox or moderation
- Map / GPS distance
- Vector embeddings as a semantic filter. The usual next step after SQL: embed each found description (e.g. Google `text-embedding-004`), store the vector, then when a lost report arrives embed it too and ask the database for the top 10 nearest rows. Fast, cheap, and it catches synonyms like “backpack” ≈ “bag” before Groq or Gemini run. Not built here because the hard filter already leaves only a handful of candidates.
- Automated tests (time was spent on the matching path instead)
- Production-grade claim verification

## What I would improve for a real product

- Notify the owner when a new found item appears in their window
- Confirm identity before showing contact details or marking claimed
- Let staff override the 7-day window and the category
- Add vector embeddings after the SQL filter so ranking can use nearest-neighbor search instead of sending every leftover row to an LLM
- Moderate photos and hide contact info until a match is accepted
- Add a short campus-specific synonym list (e.g. “cafe” = cafeteria)

## AI usage

I used Cursor (Grok) to implement the Go API, Next.js UI, Docker setup, and the matching prompts from the product rules above. For the look of the UI I pulled the TypeUI Brutalism design skill and applied its tokens (palette, type, borders, shadows) instead of inventing a one-off theme. I specified the architecture, the hard-filter rules, the dual-model split, the score mix, and what not to build. I then checked the generated wiring (typed-nil interfaces, Gemini REST image field names, date window SQL) and adjusted those.

## Project layout

```
backend/          Go API
frontend/         Next.js app
docker-compose.yml
.env.example
```

## Local development (optional)

Backend needs Postgres running (Compose is enough). From `backend/`:

```bash
go run ./cmd/server
```

From `frontend/`:

```bash
npm install
npm run dev
```

Set `POSTGRES_HOST=localhost` and `UPLOAD_DIR=./uploads` if the API runs on the host instead of Docker.
# lostnfound
