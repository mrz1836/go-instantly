# Adding a Resource Package

This is the durable recipe for wrapping another Instantly.ai V2 API resource. The
`email` package is the reference implementation — when in doubt, copy its shape.

## Architecture in one breath

- **Tiny root** (`package instantly`): the `Client` (immutable, functional
  options), the low-level plumbing (`Do`, `Get`, `Post`, `Patch`, `Put`,
  `Delete`, `DoRaw`), the shared query builder (`Query`), the generic paginator
  (`Iterate`), typed errors (`APIError`), and cross-resource enums (`SortOrder`).
- **One package per resource** (`campaign`, `lead`, …). Each imports **only**
  `github.com/mrz1836/go-instantly` and the standard library. Types live in the
  resource package. A resource never imports another resource.
- **`internal/instantlytest`**: the shared test harness (mock router, base
  `Suite`, assertion helpers). It imports `testify`, so it is reachable **only**
  from `_test.go` files. Consumers never compile it.

The import graph is acyclic by construction: every edge points into the root
sink. `campaign → instantly → stdlib`. There are no lateral edges.

## Files per resource

```
<resource>/
  <resource>.go        Service{ client *instantly.Client } + New(client) + types + methods
  options.go           ListOption func(*instantly.Query) + WithX + resource-local enums   (if it lists)
  iter.go              ListIter via instantly.Iterate                                      (if cursor-paged)
  doc.go               package doc comment
  <resource>_test.go       testify suite over internal/instantlytest
  <resource>_iter_test.go  iterator tests                                                  (if it has iter.go)
  <resource>_fuzz_test.go  round-trip + response-decode fuzz for non-trivial bodies
```

Split a large `<resource>.go` (e.g. `campaign`) into `<resource>_actions.go` to
stay under the `funlen`/file-size linters.

## Step by step

1. **Read the spec.** `.instantly-docs/endpoints-by-tag.txt` is the per-tag
   index. Pull full schemas from `.instantly-docs/api_v2.openapi.json` — e.g.
   `jq '.paths | to_entries[] | select(.key | startswith("/api/v2/campaigns"))'`.
2. **`New` + `Service`.** Exactly like `email.New`. Hold the `*instantly.Client`.
3. **Types.** One struct per request/response body. Rules:
   - Nullable fields → pointers, so "the API said nothing" stays distinguishable
     from "the API said zero". Optional request fields → pointers or `omitempty`.
   - Numbers the API sends as JSON numbers-that-might-be-anything → `*float64`.
   - Undocumented/free-form payloads → `json.RawMessage`, preserved verbatim.
   - De-stutter inside the package: `campaign.ListResponse`, not
     `campaign.CampaignListResponse`. The primary type keeps the resource name
     (`campaign.Campaign`).
4. **Methods.** Each takes `ctx` first and routes through a verb wrapper:
   `s.client.Get(ctx, path, out)`, `s.client.Post(ctx, path, req, out)`, etc.
   Escape every path parameter with `url.PathEscape`. Build list paths with
   `q.Path(basePath)`.
5. **Options** (`options.go`), if the resource lists. `type ListOption
   func(*instantly.Query)`; one `WithX` per documented query parameter; local
   enums for enum params (`SetEnum`). Reuse `instantly.SortOrder` for sort.
6. **Iterator** (`iter.go`), if the list is cursor-paged. A ~10-line closure over
   `instantly.Iterate` — copy `email/iter.go`. Append the cursor option **last**
   so it overrides a caller-supplied one.
7. **Tests.** Embed `instantlytest.Suite`; success + failure per method; a
   `Len()` completeness guard over every option; nullable-vs-zero fixtures; the
   HTTP-200-embedded-error case where the endpoint documents it; fuzz for
   non-trivial bodies. Aim for ~100% coverage (`email` is the bar).
8. **Wire up.** Tick the README coverage checkbox and add an `examples/` snippet.

## Endpoint-shape cookbook

The spec is more than clean CRUD. The template covers every shape:

| Shape | How |
|---|---|
| GET list (cursor-paged) | `q.Path(basePath)` + `iter.go` over `Iterate` |
| POST-list (`POST /leads/list`, body carries `starting_after`) | request struct with cursor field; `ListIter` still uses `Iterate`, feeding the cursor into the body |
| Bulk action (`POST /accounts/pause`) | request struct with `[]string` ids; returns a summary or `nil` |
| DELETE with body (`DELETE /leads`) | call `s.client.Do(ctx, http.MethodDelete, path, req, out)` directly |
| Non-paginated analytics GET | return `[]T` or a struct; no `iter.go` |
| Action (`activate`/`pause`/`duplicate`/`share`) | POST to the sub-path; decode into the object or pass `nil` out |
| Non-id path param (`/accounts/{email}`) | `url.PathEscape(email)` |
| CSV download (`/block-lists-entries/download`) | `s.client.DoRaw(...)` → `[]byte` |
| Singleton (`/workspaces/current`) | no id argument |

## Verification (every chunk)

```bash
magex test            # fast: lint + unit
magex test:coverrace  # race + coverage
magex lint            # 60+ linters, zero new //nolint without justification
magex format:check
go vet ./examples
go test ./<resource> -run xxx -fuzz='^Fuzz...$' -fuzztime=20s   # each new fuzz target
```

Everything must be green with **zero network** — every request routes through
`internal/instantlytest` or a `RoundTripFunc`. Spot-check the import closure:
`go list -deps ./<resource>` should show only stdlib + `instantly`.

## Gotchas the linters enforce

- `funcorder`: unexported methods come **after** the exported ones; the
  constructor comes right after its struct. Put test helper methods (`svc()`)
  at the end of the file.
- `testifylint`: use `assert.*` (not `require.*`) inside HTTP handlers — a
  handler runs on the server goroutine, where `FailNow` is illegal.
- `gocognit`/`funlen`: extract helpers rather than growing one function.
- Static errors only: `var errX = errors.New(...)`, wrapped with `%w`.
