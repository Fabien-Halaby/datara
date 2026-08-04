# 🛡️ Datara

**Enterprise MCP Data Gateway - a Zero-Trust bridge between AI agents and your private databases.**

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](https://go.dev)
[![MCP](https://img.shields.io/badge/Protocol-MCP-8A2BE2?style=flat)](https://modelcontextprotocol.io)
[![PostgreSQL](https://img.shields.io/badge/Postgres-supported-336791?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-MVP-orange)]()

Datara sits between AI agents (Claude, ChatGPT, Cursor, …) and your company's private databases. It speaks the [Model Context Protocol](https://modelcontextprotocol.io) on one side and raw SQL on the other - and in between, it guarantees that **no AI-generated query can ever modify your data**.

---

## The problem

Companies want to let AI agents explore their production data. But two things stand in the way:

- **Databases are locked inside private networks** - an LLM running in the cloud has no direct path to a Postgres instance sitting behind a corporate firewall.
- **Letting a language model run raw SQL is dangerous.** A single hallucinated `DELETE` or `DROP` can destroy data that took years to accumulate.

## The solution

Datara runs as a single lightweight binary, deployed as close to the data as possible (locally, or on the company's private server). It exposes the database through the standard MCP protocol while parsing and validating **every single query at the AST level** - not with regex or keyword blacklists, but by walking the real SQL parse tree Postgres itself would produce. Anything that isn't a clean `SELECT` is rejected before it ever reaches the database.

---

## Key features

| | |
|---|---|
| 🔒 **AST-based read-only enforcement** | Every query is parsed via `libpg_query` (the actual Postgres parser). Only single `SELECT` statements pass - no `INSERT`, `UPDATE`, `DELETE`, `DROP`, `SELECT ... INTO`, or locking clauses. |
| ⚡ **High performance** | Sub-5ms response times, a single static Go binary, ~15MB RAM footprint. |
| 🔌 **Dual-transport MCP** | `stdio` for Claude Desktop, Cursor, and local CLI tools. `SSE`/HTTP (roadmap) for web apps and cloud microservices. |
| 📋 **Structured audit log** | Every blocked and executed query is logged as JSON lines, ready to feed into any log pipeline. |
| 🏗️ **Clean Architecture** | The core domain (`SecurityPolicy`, `SQLQuery`, `QueryResult`) has zero knowledge of Postgres, MCP, or any other implementation detail - new database engines and transports plug in without touching business logic. |

---

## Architecture

```mermaid
flowchart TD
    A["AI agent<br/>(Claude, Cursor, ChatGPT)"] -->|MCP over stdio| B[Transport layer]

    subgraph Datara["Datara - single Go binary"]
        B --> C["Security core<br/>AST validator - SELECT only"]
        C -->|allowed| D["Datasource<br/>Postgres driver"]
        C -->|blocked| E["Audit log<br/>JSON lines → stderr"]
        D --> E
    end

    D --> F[(Private PostgreSQL database)]
```

Every query - whether it's allowed or rejected - is written to the audit log. Nothing bypasses the security core; the datasource layer only ever receives queries that have already been validated.

---

## Quickstart

### 1. Clone and build

```bash
git clone https://github.com/Fabien-Halaby/datara.git
cd datara
go mod tidy
go build -o bin/datara-cli ./cmd/datara/cli
```

> `pg_query_go` uses cgo (it embeds `libpg_query`, written in C) - make sure `gcc`/`build-essential` is installed.

### 2. Spin up a sample database (optional)

A ready-to-use `docker-compose.yml` seeds a Postgres instance with ~110,000 rows across `customers`, `products`, and `orders` - enough to test real-world query patterns.

```bash
docker-compose up -d
```

### 3. Run Datara

```bash
export DATARA_POSTGRES_DSN="postgres://datara:datara@localhost:5432/datara?sslmode=disable"
export DATARA_MAX_ROWS=1000  # optional, defaults to 1000
./bin/datara-cli
```

### 4. Connect it to Claude Desktop

Add Datara to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "datara": {
      "command": "/absolute/path/to/bin/datara-cli",
      "args": [],
      "env": {
        "DATARA_POSTGRES_DSN": "postgres://datara:datara@localhost:5432/datara?sslmode=disable",
        "DATARA_MAX_ROWS": "500"
      }
    }
  }
}
```

Restart Claude Desktop, and the `query_database` tool will appear in its tool list. Ask it something like:

> *"Use query_database to list the 5 most recently created customers."*

Then try to break it:

> *"Delete all cancelled orders from the database."*

Datara rejects the query before it ever touches Postgres - while your AI agent gets a clear, actionable error message instead of a silent failure.

---

## Security model

Datara's security guarantee doesn't rely on pattern-matching keywords in a string - that approach is trivially bypassed with comments, whitespace tricks, or alternate SQL syntax. Instead, every incoming query is:

1. Parsed into a real Postgres AST using [`pg_query_go`](https://github.com/pganalyze/pg_query_go) (a Go wrapper around `libpg_query`, the same C library Postgres itself uses to parse SQL).
2. Rejected outright unless the parse tree contains **exactly one `SELECT` statement**.
3. Further checked for side-effect-bearing constructs even within a `SELECT` - `SELECT ... INTO` (which creates a table) and locking clauses (`FOR UPDATE`/`FOR SHARE`) are blocked.
4. Optionally checked against a table allowlist, if one is configured in the active `SecurityPolicy`.

Every decision - allow or block - is recorded in the audit log with a timestamp and the reason for rejection.

---

## Project structure

```
datara/
├── cmd/datara/
│   ├── cli/            → entry point, stdio transport (current MVP)
│   └── api/             → entry point, SSE transport (roadmap)
├── internal/
│   ├── core/
│   │   ├── domain/      → SQLQuery, SecurityPolicy, QueryResult, domain events
│   │   └── ports/        → DataSource, SQLValidator, AuditLogger interfaces
│   ├── security/astvalidator/  → Postgres AST-based validator
│   ├── datasource/postgres/     → Postgres implementation (pgx)
│   ├── audit/                    → JSON-lines audit logger
│   ├── transport/mcpstdio/        → hand-rolled JSON-RPC 2.0 / MCP stdio server
│   └── config/                    → environment-based configuration
└── docker-compose.yml    → sample seeded Postgres instance for local testing
```

---

## Roadmap

- [x] Postgres support with strict read-only AST validation
- [x] MCP `stdio` transport
- [x] Structured JSON audit logging
- [ ] MySQL, SQLite, and SQL Server support
- [ ] `SSE`/HTTP transport for web and cloud microservices
- [ ] Configurable table/column allowlists
- [ ] Per-client rate limiting
- [ ] Multi-tenant SaaS mode (auth, billing, usage dashboards)

---

## Tech stack

| Component | Choice |
|---|---|
| Language | Go 1.26 |
| SQL parsing | [`pg_query_go`](https://github.com/pganalyze/pg_query_go) (`libpg_query`) |
| Postgres driver | [`pgx/v5`](https://github.com/jackc/pgx) |
| Protocol | [Model Context Protocol](https://modelcontextprotocol.io), implemented by hand over stdio |
| Architecture | Clean Architecture, with DDD patterns applied selectively (Value Objects, a domain-modeled `SecurityPolicy`, lightweight domain events) |

---

## License

MIT - see [LICENSE](LICENSE) for details.