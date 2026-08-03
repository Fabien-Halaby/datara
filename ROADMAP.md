# Datara — Roadmap

## Phase 1 : MVP Open Source (semaines 1–3)

| Semaine | Tâche | Livrable |
|---|---|---|
| 1 | Setup repo, architecture Clean, module Go, CLI cobra | `cmd/datara` fonctionnel, `go.mod` propre |
| 1 | MCP stdio transport + tool `query_db` | Connexion Claude Desktop OK |
| 2 | AST guard PostgreSQL (wasilibs/go-pgquery) | Blocage SELECT-only, tests unitaires |
| 2 | Tool `list_tables` + `describe_table` | Introspection schema fonctionnelle |
| 2 | Tests unitaires + mocks (sqlmock) | Coverage &gt;70% |
| 3 | README pro + gif démo + badges | Repo GitHub public, 1ère release |
| 3 | CI/CD GitHub Actions + goreleaser | Binaires cross-platform auto-publiés |

---

## Phase 2 : SSE + Auth (semaine 4)

| Semaine | Tâche | Livrable |
|---|---|---|
| 4 | SSE transport HTTP (mcp-go SSE server) | Endpoint `/sse` fonctionnel |
| 4 | Auth JWT middleware + API key rotation | Protection endpoint SSE |
| 4 | Config YAML (Viper) + env vars | `datara.yaml` + `.env` |
| 4 | Docker image Alpine | Image ~20MB, `docker run` prêt |
| 4 | Health check `/health` + metrics Prometheus | Observabilité basique |

---

## Phase 3 : Mini SaaS Ready (semaines 5–7)

| Semaine | Tâche | Livrable |
|---|---|---|
| 5 | Multi-tenant : `organization_id` + isolation DB | Connexion par tenant, séparation données |
| 5 | Rate limiting (token bucket) | Protection abuse API |
| 6 | Dashboard web minimal (Next.js) | Vue connexions, logs requêtes, stats |
| 6 | Audit log complet | Traçabilité qui/quoi/quand |
| 7 | Stripe integration (freemium) | 1 connexion gratuite, $9/mois ensuite |
| 7 | Déploiement cloud-ready | `docker-compose.yml` + Helm chart basique |

---

## Phase 4 : Scale (semaines 8+, optionnel)

| Tâche | Description |
|---|---|
| Support MySQL | Parser TiDB (`pingcap/tidb/pkg/parser`) + driver MySQL |
| Query cache Redis | Cache résultats fréquents, TTL configurable |
| MCP Apps | UI interactive dans Claude (spec 2026-07-28) |
| Row-level security | RLS côté PostgreSQL pour defense in depth |
| SSO Enterprise | SAML/OIDC pour clients enterprise |

---

## KPIs & Milestones

| Milestone | Objectif | Date cible |
|---|---|---|
| MVP fonctionnel | `query_db` + AST guard + Claude Desktop | Semaine 3 |
| 1ère release GitHub | v0.1.0 avec binaires | Semaine 3 |
| SSE + Auth | Endpoint sécurisé, Docker ready | Semaine 4 |
| Mini SaaS live | Dashboard + Stripe, démo publique | Semaine 7 |
| 100 stars GitHub | Communauté open source | Mois 3 |
| 10 clients payants | Validation marché SaaS | Mois 6 |