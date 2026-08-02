<div align="center">

# 📬&nbsp;&nbsp;go-instantly

**Unofficial Golang Library for the Instantly.ai API (V2).**

<br/>

<a href="https://github.com/mrz1836/go-instantly/releases"><img src="https://img.shields.io/github/release-pre/mrz1836/go-instantly?include_prereleases&style=flat-square&logo=github&color=black" alt="Release"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/github/go-mod/go-version/mrz1836/go-instantly?style=flat-square&logo=go&color=00ADD8" alt="Go Version"></a>
<a href="https://github.com/mrz1836/go-instantly/blob/master/LICENSE"><img src="https://img.shields.io/github/license/mrz1836/go-instantly?style=flat-square&color=blue" alt="License"></a>

<br/>

<table align="center" border="0">
  <tr>
    <td align="right">
       <code>CI / CD</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/go-instantly/actions"><img src="https://img.shields.io/github/actions/workflow/status/mrz1836/go-instantly/fortress.yml?branch=master&label=build&logo=github&style=flat-square" alt="Build"></a>
       <a href="https://github.com/mrz1836/go-instantly/actions"><img src="https://img.shields.io/github/last-commit/mrz1836/go-instantly?style=flat-square&logo=git&logoColor=white&label=last%20update" alt="Last Commit"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Quality</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://goreportcard.com/report/github.com/mrz1836/go-instantly"><img src="https://goreportcard.com/badge/github.com/mrz1836/go-instantly?style=flat-square" alt="Go Report"></a>
       <a href="https://codecov.io/gh/mrz1836/go-instantly"><img src="https://codecov.io/gh/mrz1836/go-instantly/branch/master/graph/badge.svg?style=flat-square" alt="Coverage"></a>
    </td>
  </tr>

  <tr>
    <td align="right">
       <code>Security</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://scorecard.dev/viewer/?uri=github.com/mrz1836/go-instantly"><img src="https://api.scorecard.dev/projects/github.com/mrz1836/go-instantly/badge?style=flat-square" alt="Scorecard"></a>
       <a href=".github/SECURITY.md"><img src="https://img.shields.io/badge/policy-active-success?style=flat-square&logo=security&logoColor=white" alt="Security"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Community</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/go-instantly/graphs/contributors"><img src="https://img.shields.io/github/contributors/mrz1836/go-instantly?style=flat-square&color=orange" alt="Contributors"></a>
       <a href="https://mrz1818.com/"><img src="https://img.shields.io/badge/donate-bitcoin-ff9900?style=flat-square&logo=bitcoin" alt="Bitcoin"></a>
    </td>
  </tr>
</table>

</div>

<br/>
<br/>

<div align="center">

### <code>Project Navigation</code>

</div>

<table align="center">
  <tr>
    <td align="center" width="33%">
       🚀&nbsp;<a href="#-installation"><code>Installation</code></a>
    </td>
    <td align="center" width="33%">
       🧪&nbsp;<a href="#-examples--tests"><code>Examples&nbsp;&&nbsp;Tests</code></a>
    </td>
    <td align="center" width="33%">
       📚&nbsp;<a href="#-documentation"><code>Documentation</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
       🤝&nbsp;<a href="#-contributing"><code>Contributing</code></a>
    </td>
    <td align="center">
      🛠️&nbsp;<a href="#-code-standards"><code>Code&nbsp;Standards</code></a>
    </td>
    <td align="center">
      ⚡&nbsp;<a href="#-benchmarks"><code>Benchmarks</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
      🤖&nbsp;<a href="#-ai-usage--assistant-guidelines"><code>AI&nbsp;Usage</code></a>
    </td>
    <td align="center">
       ⚖️&nbsp;<a href="#-license"><code>License</code></a>
    </td>
    <td align="center">
       👥&nbsp;<a href="#-maintainers"><code>Maintainers</code></a>
    </td>
  </tr>
</table>
<br/>

## 📦 Installation

**go-instantly** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).
```shell script
go get github.com/mrz1836/go-instantly
```

<br/>

## 💡 Usage

This library targets the **Instantly.ai API V2 only**. V2 is not backwards compatible with V1 and uses
its own credentials — a V1 key will not authenticate against V2, so
[create a V2 API key](https://developer.instantly.ai/getting-started/authorization) in your workspace
settings and pass it to `NewClient`. Every request is then authenticated with an
`Authorization: Bearer <key>` header against `https://api.instantly.ai`.

```go
package main

import (
	"context"
	"log"

	"github.com/mrz1836/go-instantly"
)

func main() {
	client := instantly.NewClient("[API-KEY]")

	err := client.SendTestEmail(context.Background(), instantly.SendTestEmailRequest{
		EAccount:           "sender@example.com",
		ToAddressEmailList: "recipient@example.com",
		Subject:            "Testing the sending account",
		Body: instantly.EmailBody{
			HTML: "<p>Hello from go-instantly.</p>",
			Text: "Hello from go-instantly.",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

List endpoints take **functional options**, so only the filters you actually pass are sent:

```go
page, err := client.ListEmails(ctx,
	instantly.WithEmailLimit(50),
	instantly.WithEmailIsUnread(true),
	instantly.WithEmailMode(instantly.EmailModeFocused),
)
```

Pagination is cursor based. `ListEmails` returns a single page, and `ListEmailsIter` walks every page
for you as a [range-over-func](https://go.dev/blog/range-functions) iterator:

```go
for email, err := range client.ListEmailsIter(ctx, instantly.WithEmailIsUnread(true)) {
	if err != nil {
		return err
	}
	log.Printf("%s: %s", email.ID, email.Subject)
}
```

Some endpoints report a failure inside an otherwise successful HTTP 200 response. This library turns
those into real errors, so `err != nil` catches every failure regardless of which shape it arrived in:

```go
var apiErr *instantly.APIError
if errors.As(err, &apiErr) && apiErr.Code == instantly.ErrCodeAccountAuthError {
	// the sending account failed to authenticate
}
```

<br/>

## 📚 Documentation

View the generated [documentation](https://pkg.go.dev/github.com/mrz1836/go-instantly?tab=doc), or the
[Instantly.ai API reference](https://developer.instantly.ai/api-reference/introduction) for endpoint
details, scopes, and per-endpoint rate limits.

> **Heads up!** `go-instantly` is intentionally light on dependencies. The only external package it
uses is the excellent `testify` suite—and that's just for our tests. You can drop this library into
your projects without dragging along extra baggage.

<br/>

<details>
<summary><strong><code>Supported API Coverage</code></strong></summary>
<br/>

Coverage is built one resource at a time. The [Email API](https://developer.instantly.ai/api-reference/email)
ships today; every remaining V2 resource is listed below with its operation count so you can see
exactly what is left.

* [x] **[Email API](https://developer.instantly.ai/api-reference/email) — ([email.go](email.go))**
	* [x] [`POST /api/v2/emails/test`](email.go) - Send a test email
	* [x] [`GET /api/v2/emails`](email.go) - List emails
	* [x] [`GET /api/v2/emails/{id}`](email.go) - Get an email
	* [x] [`PATCH /api/v2/emails/{id}`](email.go) - Patch an email
	* [x] [`DELETE /api/v2/emails/{id}`](email.go) - Delete an email
	* [x] [`POST /api/v2/emails/reply`](email.go) - Reply to an email
	* [x] [`POST /api/v2/emails/forward`](email.go) - Forward an email
	* [x] [`GET /api/v2/emails/unread/count`](email.go) - Count unread emails
	* [x] [`POST /api/v2/emails/threads/{thread_id}/mark-as-read`](email.go) - Mark all emails in a thread as read

**Coming soon** — the remaining V2 resources, with the number of operations each one covers:

* [ ] **Campaign** - 19 operations
* [ ] **Account** - 16 operations
* [ ] **Lead** - 13 operations
* [ ] **SuperSearchEnrichment** - 11 operations
* [ ] **BlockListEntry** - 9 operations
* [ ] **CampaignSubsequence** - 9 operations
* [ ] **Webhook** - 8 operations
* [ ] **Workspace** - 8 operations
* [ ] **Analytics** - 7 operations
* [ ] **DFYEmailAccountOrder** - 7 operations
* [ ] **CustomTag** - 6 operations
* [ ] **InboxPlacementTest** - 6 operations
* [ ] **LeadLabel** - 6 operations
* [ ] **LeadList** - 6 operations
* [ ] **InboxPlacementAnalytics** - 5 operations
* [ ] **WorkspaceGroupMember** - 5 operations
* [ ] **WorkspaceMember** - 5 operations
* [ ] **WebhookEvent** - 4 operations
* [ ] **APIKey** - 3 operations
* [ ] **OAuth** - 3 operations
* [ ] **BackgroundJob** - 2 operations
* [ ] **CRMActions** - 2 operations
* [ ] **EmailVerification** - 2 operations
* [ ] **InboxPlacementBlacklistAndSpamAssassinReport** - 2 operations
* [ ] **WorkspaceBilling** - 2 operations
* [ ] **AccountCampaignMapping** - 1 operation
* [ ] **AuditLog** - 1 operation
* [ ] **CustomTagMapping** - 1 operation

</details>

<details>
<summary><strong><code>Custom HTTPClient Support</code></strong></summary>
<br/>

Every field on the client is exported, so the HTTP client, the API key, and the base URL can all be
replaced — useful for custom transports, proxies, or pointing the client at a test server.

```go
package main

import (
    "net/http"
    "time"

    "github.com/mrz1836/go-instantly"
)

// ....

client := instantly.NewClient("[API-KEY]")

client.HTTPClient = &http.Client{
    Timeout: 30 * time.Second,
}

// ...
```
</details>

<details>
<summary><strong><code>Development Setup (Getting Started)</code></strong></summary>
<br/>

Install [MAGE-X](https://github.com/mrz1836/mage-x) build tool for development:

```bash
# Install MAGE-X for development and building
go install github.com/mrz1836/mage-x/cmd/magex@latest
magex update:install
```
</details>

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

This project uses [goreleaser](https://github.com/goreleaser/goreleaser) for streamlined binary and library deployment to GitHub. To get started, install it via:

```bash
brew install goreleaser
```

The release process is defined in the [.goreleaser.yml](.goreleaser.yml) configuration file.

Then create and push a new Git tag using:

```bash
magex version:bump bump=patch push=true branch=master
```

This process ensures consistent, repeatable releases with properly versioned artifacts and citation metadata.

</details>

<details>
<summary><strong><code>Build Commands</code></strong></summary>
<br/>

View all build commands

```bash script
magex help
```

</details>

<details>
<summary><strong>GitHub Workflows</strong></summary>
<br/>

All workflows are driven by modular configuration in [`.github/env/`](.github/env/README.md) — no YAML editing required.

**[View all workflows and the control center →](.github/docs/workflows.md)**

</details>

<details>
<summary><strong><code>Updating Dependencies</code></strong></summary>
<br/>

To update all dependencies (Go modules, linters, and related tools), run:

```bash
magex deps:update
```

This command ensures all dependencies are brought up to date in a single step, including Go modules and any managed tools. It is the recommended way to keep your development environment and CI in sync with the latest versions.

</details>

<br/>

## 🧪 Examples & Tests

All unit tests and fuzz tests run via [GitHub Actions](https://github.com/mrz1836/go-instantly/actions) and use [Go version 1.25.x](https://go.dev/doc/go1.25). View the [configuration file](.github/workflows/fortress.yml).

Run all tests (fast):

```bash script
magex test
```

Run all tests with race detector (slower):
```bash script
magex test:race
```

Browse the runnable usage samples in [examples/examples.go](examples/examples.go).

> **Note:** the test suite runs entirely against an in-repo mock router. No test contacts the live
Instantly.ai API, and the examples file is illustrative — it is compiled, never executed by CI.

<br/>

## ⚡ Benchmarks

Run the Go benchmarks:

```bash script
magex bench
```

### 📊 Performance Results

_Published benchmark results are pending._ Once a benchmark suite lands, measured numbers will be
recorded here rather than estimated.

<br/>

## 🛠️ Code Standards
Read more about this Go project's [code standards](.github/CODE_STANDARDS.md).

<br/>

## 🤖 AI Usage & Assistant Guidelines
Read the [AI Usage & Assistant Guidelines](.github/tech-conventions/ai-compliance.md) for details on how AI is used in this project and how to interact with the AI assistants.

<br/>

## 👥 Maintainers
| [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) |
|:------------------------------------------------------------------------------------------------:|
|                                [MrZ](https://github.com/mrz1836)                                 |

<br/>

## 🤝 Contributing
View the [contributing guidelines](.github/CONTRIBUTING.md) and please follow the [code of conduct](.github/CODE_OF_CONDUCT.md).

### How can I help?
All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:.
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/mrz1836) :clap:
or by making a [**bitcoin donation**](https://mrz1818.com/?tab=tips&utm_source=github&utm_medium=sponsor-link&utm_campaign=go-instantly&utm_term=go-instantly&utm_content=go-instantly) to ensure this journey continues indefinitely! :rocket:


[![Stars](https://img.shields.io/github/stars/mrz1836/go-instantly?label=Please%20like%20us&style=social)](https://github.com/mrz1836/go-instantly/stargazers)

<br/>

## 📝 License

[![License](https://img.shields.io/github/license/mrz1836/go-instantly.svg?style=flat)](LICENSE)
