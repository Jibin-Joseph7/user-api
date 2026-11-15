# User API

A small REST API written in Go that manages users and their dates of birth. The part I enjoyed most: the API never stores an age. It stores the date of birth and works out the age every time you ask, so the number is always right.

## What it's built with

- **Go** and **GoFiber** for the web server
- **MySQL** for storage, with **SQLC** generating the database code from plain SQL
- **Uber Zap** for structured logging
- **go-playground/validator** for checking input
- **Docker Compose** so the whole thing runs with one command

## How the code is organised

I split the code into layers so each piece does one job:

```
cmd/server/main.go     starts everything up
config/                reads settings from the environment
db/migrations/         the SQL schema
db/queries/            the SQL that SQLC turns into Go
db/sqlc/               generated code (don't edit by hand)
internal/
  handler/             HTTP handlers and input validation
  service/             business logic, including the age calculation
  repository/          talks to the database through SQLC
  routes/              wires URLs to handlers
  middleware/          request ID and request timing
  models/              request and response shapes
  logger/              Zap setup
```

A request flows handler → service → repository → database. The service never touches HTTP, and the handler never touches SQL.

## Getting it running

### The easy way: Docker

You only need [Docker Desktop](https://www.docker.com/products/docker-desktop/).

```bash
docker compose up --build
```

Once you see `server starting` in the logs, the API is live at http://localhost:3000. The database and table are created automatically the first time.

To stop it, press `Ctrl+C` and run `docker compose down`. Add `-v` to also wipe the stored data.

### Running it on your machine

You'll need Go and MySQL 8. [sqlc](https://sqlc.dev) is only needed if you change the queries.

1. Create the database and table:

```bash
mysql -u root -p -e "CREATE DATABASE user_api;"
mysql -u root -p user_api < db/migrations/001_create_users.sql
```

2. Copy `.env.example` to `.env` and put in your own password:

```
PORT=3000
DATABASE_URL=root:YOUR_PASSWORD@tcp(localhost:3306)/user_api?parseTime=true
```

The `parseTime=true` part matters. Without it, the driver can't read dates into Go's `time.Time`.

3. Start the server:

```bash
go mod tidy
go run ./cmd/server
```

If you edit anything in `db/queries` or `db/migrations`, regenerate the database code with:

```bash
sqlc generate
```

## Using the API

| Method | Endpoint | What it does | Success |
|---|---|---|---|
| POST | `/users` | Create a user | 201 |
| GET | `/users/:id` | Get one user, with age | 200 |
| PUT | `/users/:id` | Update a user | 200 |
| DELETE | `/users/:id` | Delete a user | 204 |
| GET | `/users?page=1&limit=10` | List users, with age | 200 |

Dates are written as `YYYY-MM-DD`. When something goes wrong you get a JSON message like `{"error": "..."}` with a 400 (bad input), 404 (user not found) or 500 (something broke on my side).

### Try it

Create a user:

```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","dob":"1990-05-10"}'
```

```json
{"id": 1, "name": "Alice", "dob": "1990-05-10"}
```

Fetch her, and the age shows up:

```bash
curl http://localhost:3000/users/1
```

```json
{"id": 1, "name": "Alice", "dob": "1990-05-10", "age": 36}
```

Update, delete and list:

```bash
curl -X PUT http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Smith","dob":"1991-01-20"}'

curl -X DELETE http://localhost:3000/users/1

curl "http://localhost:3000/users?page=1&limit=5"
```

Pagination notes: `page` starts at 1, `limit` defaults to 10 and is capped at 100. The total number of users comes back in the `X-Total-Count` response header.

## Details worth knowing

- **Age calculation:** it's the difference in years, minus one if this year's birthday hasn't happened yet. I deliberately avoided dividing days by 365, because leap years would make it wrong around birthdays.
- **Validation:** a name is required, and the date must be a real date in `YYYY-MM-DD` format. A date of birth in the future is rejected.
- **Request IDs:** every response carries an `X-Request-ID` header. If you send one, I reuse it. If not, a new one is generated. That makes it easy to find a single request in the logs.
- **Logging:** each request is logged with its method, path, status, request ID and how long it took.
- **Startup:** the server retries the database connection a few times, so it copes with MySQL taking a moment to start.

## Tests

```bash
go test ./...
```

The unit tests cover the age calculation: birthday already passed, today, tomorrow, born today, and leap-day birthdays on either side of March 1st.