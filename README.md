# Calculator

![CI](https://github.com/isabelagutierrez/calculator/actions/workflows/ci.yml/badge.svg)

Full-stack calculator with a React/TypeScript interface and a dependency-free
Go API. It supports addition, subtraction, multiplication, division,
exponents, square roots, percentages, parentheses, history, and keyboard input.

## Project structure

```text
backend/
  cmd/server/
  internal/{api,calculator}/
frontend/src/
  calculator/
  components/
docs/coverage/
```

## Requirements

- Go 1.22+
- Node.js 22.12+ (Vitest 5 does not run on Node 20)
- npm
- Docker (optional)

## Local development

```bash
cd backend
go run ./cmd/server
```

In another terminal:

```bash
cd frontend
npm ci
npm run dev
```

Open `http://localhost:5173`. Vite listens on all interfaces, so a phone on
the same network can use `http://<computer-ip>:5173`.

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | Backend listen port |
| `CORS_ALLOWED_ORIGIN` | `*` | Allowed browser origin |
| `VITE_API_BASE_URL` | current origin | API base URL used by the frontend |

## Docker

```bash
docker compose up --build
```

The application is available at `http://localhost:3000`. Stop it with
`docker compose down`. Nginx serves the frontend and proxies `/api/` to Go.

## API

All POST requests require `Content-Type: application/json` (a charset
parameter is accepted).

| Method | Route | Body | Success |
| --- | --- | --- | --- |
| GET | `/health` | none | `{"status":"ok"}` |
| POST | `/api/add`, `/api/subtract`, `/api/multiply`, `/api/divide`, `/api/power` | `{"left":8,"right":4}` | `{"result":...}` |
| POST | `/api/square-root`, `/api/percentage` | `{"value":81}` | `{"result":...}` |
| POST | `/api/calculate` | `{"expression":"8 + 8"}` | `{"result":16}` |

Errors use `{"error":"message"}`. Status codes are: 200 success, 400 invalid
JSON or fields, 404 unknown route, 405 wrong method, 413 body over 8 KiB,
415 wrong content type, 422 invalid mathematical expression or operation, and
500 unexpected server failure.

```bash
curl -H "Content-Type: application/json" -d '{"left":8,"right":4}' http://localhost:8080/api/add
curl -H "Content-Type: application/json" -d '{"expression":"2 ^ 3 ^ 2"}' http://localhost:8080/api/calculate
curl -H "Content-Type: application/json" -d '{"left":8,"right":0}' http://localhost:8080/api/divide
```

On Windows PowerShell, use `curl.exe` instead of the PowerShell `curl` alias.

## Expression evaluator

The parser uses a lexer and recursive descent grammar:

```text
expression = term { ("+" | "-") term }
term       = unary { ("*" | "/") unary }
unary      = { "+" | "-" } power
power      = primary [ "^" unary ]
primary    = number | "(" expression ")"
```

Power is right associative and has higher precedence than unary signs, so
`-2 ^ 2` is `-4`. ASCII operators and `×`, `÷`, `−`, and `xʸ` are accepted.
Expressions are limited to 500 runes, 100 parenthesis levels, and 8 KiB HTTP
bodies. The backend is the only source of truth; it never uses `eval`.

## Tests and coverage

```bash
make test
make lint
make coverage
```

Backend coverage is written to `backend/coverage.html`; frontend HTML and LCOV
reports are written to `frontend/coverage/`. Text snapshots are kept in
[docs/coverage/](docs/coverage/). Coverage thresholds are 90% for backend
internal packages and 85% lines/statements/functions plus 80% branches for
frontend.

`npm audit fix` leaves moderate vulnerabilities in development-only
TypeScript/ESLint tooling; resolving them requires the breaking `--force`
upgrade, which is intentionally not used.

## Design decisions and limitations

The pure frontend reducer is independently testable and stores tokenized
expressions. The Go parser keeps arithmetic rules, positions, and defensive
limits in one place. Nginx provides same-origin production proxying, while
timeouts and recovery middleware protect both layers. HTTP 400 means malformed
input; HTTP 422 means valid input that cannot be calculated. `float64` is used
for API interoperability and therefore has normal floating-point precision.

Percentage means divide by 100, not contextual percentage arithmetic. Implicit
multiplication such as `2(3)` is not supported. History is in memory and is
lost on reload. The display shows up to 12 significant digits.

## Keyboard shortcuts

Digits, `.`, `+`, `-`, `*`, `/`, `^`, `%`, parentheses, Enter, Backspace,
Delete, `c`, `Escape`, and `±` are supported. Ctrl, Meta, and Alt modified
events are ignored so browser shortcuts remain available.

## AI use

The implementation was reviewed and verified with AI assistance. The
chronological prompts and a manual verification summary are in
[PROMPTS.md](PROMPTS.md).
