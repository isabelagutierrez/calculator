# Calculator

A full-stack calculator with a React and TypeScript frontend and a Go REST API.

## Features

- Addition, subtraction, multiplication, division, exponentiation, square root, and percentage.
- Parenthesized expressions with standard operator precedence.
- Input validation and user-friendly error messages.
- Keyboard support and responsive mobile layout.
- Internal history panel with independent scrolling.
- Backend validation for missing, invalid, non-finite, and unsafe operations.

## Requirements

- Go 1.22+
- Node.js 20 or newer
- npm

## Run locally

Start the backend:

```powershell
cd backend
go run ./cmd/server
```

The API listens on `http://localhost:8080`.

In another terminal, start the frontend:

```powershell
cd frontend
npm install
npm run dev
```

The Vite development server listens on `http://localhost:5173` and proxies
`/api` requests to the backend.

To access the frontend from a phone on the same local network:

```powershell
npm run dev
```

Then open `http://<computer-ip>:5173` on the phone. The Vite host is configured
for local-network access in [frontend/vite.config.ts](frontend/vite.config.ts).

## API

All operation endpoints accept `POST` requests with JSON and return:

```json
{ "result": 16 }
```

### Binary operations

```http
POST /api/add
Content-Type: application/json

{ "left": 8, "right": 8 }
```

Available binary endpoints:

- `/api/add`
- `/api/subtract`
- `/api/multiply`
- `/api/divide`
- `/api/power`

### Unary operations

```http
POST /api/square-root
Content-Type: application/json

{ "value": 81 }
```

Available unary endpoints:

- `/api/square-root`
- `/api/percentage`

### Expression evaluation

The frontend sends accumulated expressions to the backend:

```http
POST /api/calculate
Content-Type: application/json

{ "expression": "8 × 8 + (2 ^ 3)" }
```

The evaluator uses this grammar:

```text
expression = term { ("+" | "-") term }
term       = unary { ("*" | "/") unary }
unary      = { "+" | "-" } power
power      = primary [ "^" unary ]
primary    = number | "(" expression ")"
```

Unary signs are accepted in a loop, so `--2` is valid. Exponentiation binds
more tightly than a unary minus, making `-2 ^ 2` equal to `-4`; parentheses
make `(-2) ^ 2` equal to `4`. The `^` operator is right-associative, so
`2 ^ 3 ^ 2` equals `512`. The evaluator also supports `+`, `-`, `×`, `÷`,
`^`, and parentheses. The `/health` endpoint returns `{ "status": "ok" }`.

Errors use a JSON response such as:

```json
{ "error": "cannot divide by zero" }
```

## Testing and coverage

Run backend tests and coverage:

```powershell
cd backend
go test ./... -cover
go test ./internal/... -coverprofile coverage.out
go tool cover -html=coverage.out -o coverage.html
```

Run frontend tests:

```powershell
cd frontend
npm test -- --run
npm run build
npm run lint
```

The root [Makefile](Makefile) also provides `make backend`, `make frontend`,
`make test`, `make docker-up`, `make docker-down`, `make docker-build`, and
`make docker-logs` targets when Make is available.

## Docker

Build and start the frontend and backend together:

```powershell
docker compose up --build
```

Open the application at `http://localhost:3000`. The frontend container serves
the production React build through Nginx and proxies API requests to the
backend container. The backend is not exposed directly to the host.

Stop the containers with:

```powershell
docker compose down
```

The same Docker commands are available through Make:

```powershell
make docker-up       # Build and start the application
make docker-down     # Stop and remove containers
make docker-build    # Build images without starting them
make docker-logs     # Follow container logs
```

## Design rationale

The frontend owns input presentation, responsive layout, keyboard handling, and
history display. The backend owns expression evaluation and arithmetic
validation so the same rules can be tested independently of the browser.
Handlers are kept separate from the calculator package, allowing arithmetic
unit tests and HTTP contract tests to remain focused.
