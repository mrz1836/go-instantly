<div align="center">

# 📬&nbsp;&nbsp;go-instantly

**Unofficial Golang Library for the Instantly.ai API (V2).**

<br/>

<a href="https://github.com/mrz1836/go-instantly/releases"><img src="https://img.shields.io/github/release-pre/mrz1836/go-instantly?include_prereleases&style=flat-square&logo=github&color=black" alt="Release"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/github/go-mod/go-version/mrz1836/go-instantly?style=flat-square&logo=go&color=00ADD8" alt="Go Version"></a>
<a href="https://github.com/mrz1836/go-instantly/blob/master/LICENSE"><img src="https://img.shields.io/github/license/mrz1836/go-instantly?style=flat-square&color=blue&v=1" alt="License"></a>

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

The SDK is organized **AWS-SDK-v2 style**: a tiny root package holds the shared `Client`, and each API
resource lives in its own subpackage (`email`, `campaign`, `account`, …). Construct the client once and
hand it to a resource service. Importing a resource pulls in only that resource plus the root package
and the standard library — never `testify` or the other resources.

```go
package main

import (
	"context"
	"log"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/email"
)

func main() {
	client := instantly.NewClient("[API-KEY]")
	emails := email.New(client)

	err := emails.SendTest(context.Background(), email.SendTestRequest{
		EAccount:           "sender@example.com",
		ToAddressEmailList: "recipient@example.com",
		Subject:            "Testing the sending account",
		Body: email.Body{
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
page, err := emails.List(ctx,
	email.WithLimit(50),
	email.WithIsUnread(true),
	email.WithMode(email.ModeFocused),
)
```

Pagination is cursor based. `List` returns a single page, and `ListIter` walks every page for you as a
[range-over-func](https://go.dev/blog/range-functions) iterator:

```go
for message, err := range emails.ListIter(ctx, email.WithIsUnread(true)) {
	if err != nil {
		return err
	}
	log.Printf("%s: %s", message.ID, message.Subject)
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

Coverage is built one resource at a time against the
[Instantly V2 OpenAPI spec](https://api.instantly.ai/openapi/api_v2.json) — **171 operations across 127
endpoints in 28 resource groups**. Each resource is its own Go package (AWS-SDK-v2 style).
**94 / 171 operations · 12 / 28 resources** ship today (Email, Account, Account-Campaign Mappings,
Campaign, Lead, Lead List, Lead Label, Campaign Subsequence, Email Verification, Inbox Placement Test,
Inbox Placement Analytics, Inbox Placement Report); every remaining resource is listed below with a link to its
reference docs and its operation count, so you can see exactly what is left.

* [x] **[Email API](https://developer.instantly.ai/api-reference/email) — ([`email/`](email/email.go))**
	* [x] [`POST /api/v2/emails/test`](email/email.go) - Send a test email (`email.Service.SendTest`)
	* [x] [`GET /api/v2/emails`](email/email.go) - List emails (`email.Service.List` / `ListIter`)
	* [x] [`GET /api/v2/emails/{id}`](email/email.go) - Get an email (`email.Service.Get`)
	* [x] [`PATCH /api/v2/emails/{id}`](email/email.go) - Patch an email (`email.Service.Update`)
	* [x] [`DELETE /api/v2/emails/{id}`](email/email.go) - Delete an email (`email.Service.Delete`)
	* [x] [`POST /api/v2/emails/reply`](email/email.go) - Reply to an email (`email.Service.Reply`)
	* [x] [`POST /api/v2/emails/forward`](email/email.go) - Forward an email (`email.Service.Forward`)
	* [x] [`GET /api/v2/emails/unread/count`](email/email.go) - Count unread emails (`email.Service.CountUnread`)
	* [x] [`POST /api/v2/emails/threads/{thread_id}/mark-as-read`](email/email.go) - Mark a thread as read (`email.Service.MarkThreadAsRead`)
* [x] **[Account API](https://developer.instantly.ai/api-reference/account) — ([`account/`](account/account.go))**
	* [x] `POST /api/v2/accounts` - Create account (`account.Service.Create`)
	* [x] `GET /api/v2/accounts` - List accounts (`account.Service.List` / `ListIter`)
	* [x] `GET /api/v2/accounts/{email}` - Get account (`account.Service.Get`)
	* [x] `PATCH /api/v2/accounts/{email}` - Patch account (`account.Service.Update`)
	* [x] `DELETE /api/v2/accounts/{email}` - Delete account (`account.Service.Delete`)
	* [x] `POST /api/v2/accounts/{email}/pause` - Pause an account (`account.Service.Pause`)
	* [x] `POST /api/v2/accounts/pause` - Pause multiple accounts (`account.Service.PauseBulk`)
	* [x] `POST /api/v2/accounts/{email}/resume` - Resume an account (`account.Service.Resume`)
	* [x] `POST /api/v2/accounts/{email}/mark-fixed` - Mark an account fixed (`account.Service.MarkFixed`)
	* [x] `POST /api/v2/accounts/move` - Move accounts between workspaces (`account.Service.Move`)
	* [x] `POST /api/v2/accounts/warmup/enable` - Enable warmup (`account.Service.EnableWarmup`)
	* [x] `POST /api/v2/accounts/warmup/disable` - Disable warmup (`account.Service.DisableWarmup`)
	* [x] `POST /api/v2/accounts/warmup-analytics` - Warmup analytics (`account.Service.WarmupAnalytics`)
	* [x] `GET /api/v2/accounts/analytics/daily` - Daily analytics (`account.Service.DailyAnalytics`)
	* [x] `GET /api/v2/accounts/ctd/status` - Custom tracking domain status (`account.Service.CtdStatus`)
	* [x] `POST /api/v2/accounts/test/vitals` - Test account vitals (`account.Service.TestVitals`)
* [x] **[Account-Campaign Mappings](https://developer.instantly.ai/api-reference/accountcampaignmapping) — ([`accountcampaign/`](accountcampaign/accountcampaign.go))**
	* [x] `GET /api/v2/account-campaign-mappings/{email}` - Campaigns for an account (`accountcampaign.Service.List` / `ListIter`)
* [x] **[Campaign API](https://developer.instantly.ai/api-reference/campaign) — ([`campaign/`](campaign/campaign.go))**
	* [x] `POST /api/v2/campaigns` - Create campaign (`campaign.Service.Create`)
	* [x] `GET /api/v2/campaigns` - List campaigns (`campaign.Service.List` / `ListIter`)
	* [x] `GET /api/v2/campaigns/{id}` - Get campaign (`campaign.Service.Get`)
	* [x] `PATCH /api/v2/campaigns/{id}` - Patch campaign (`campaign.Service.Update`)
	* [x] `DELETE /api/v2/campaigns/{id}` - Delete campaign (`campaign.Service.Delete`)
	* [x] `POST /api/v2/campaigns/{id}/activate` - Activate/resume (`campaign.Service.Activate`)
	* [x] `POST /api/v2/campaigns/{id}/pause` - Pause (`campaign.Service.Pause`)
	* [x] `POST /api/v2/campaigns/{id}/duplicate` - Duplicate (`campaign.Service.Duplicate`)
	* [x] `POST /api/v2/campaigns/{id}/share` - Share (`campaign.Service.Share`)
	* [x] `POST /api/v2/campaigns/{id}/export` - Export to JSON (`campaign.Service.Export`)
	* [x] `POST /api/v2/campaigns/{id}/from-export` - Create from export (`campaign.Service.CreateFromExport`)
	* [x] `POST /api/v2/campaigns/{id}/variables` - Add variables (`campaign.Service.AddVariables`)
	* [x] `GET /api/v2/campaigns/{id}/sending-status` - Sending status (`campaign.Service.SendingStatus`)
	* [x] `GET /api/v2/campaigns/count-launched` - Launched count (`campaign.Service.CountLaunched`)
	* [x] `GET /api/v2/campaigns/search-by-contact` - Search by contact (`campaign.Service.SearchByContact`)
	* [x] `GET /api/v2/campaigns/analytics` - Analytics (`campaign.Service.Analytics`)
	* [x] `GET /api/v2/campaigns/analytics/overview` - Analytics overview (`campaign.Service.AnalyticsOverview`)
	* [x] `GET /api/v2/campaigns/analytics/daily` - Daily analytics (`campaign.Service.DailyAnalytics`)
	* [x] `GET /api/v2/campaigns/analytics/steps` - Steps analytics (`campaign.Service.StepsAnalytics`)
* [x] **[Lead API](https://developer.instantly.ai/api-reference/lead) — ([`lead/`](lead/lead.go))**
	* [x] `POST /api/v2/leads` - Create lead (`lead.Service.Create`)
	* [x] `POST /api/v2/leads/list` - List leads, POST-body filters + cursor (`lead.Service.List` / `ListIter`)
	* [x] `GET /api/v2/leads/{id}` - Get lead (`lead.Service.Get`)
	* [x] `PATCH /api/v2/leads/{id}` - Patch lead (`lead.Service.Update`)
	* [x] `DELETE /api/v2/leads/{id}` - Delete lead (`lead.Service.Delete`)
	* [x] `DELETE /api/v2/leads` - Bulk delete, body on DELETE (`lead.Service.BulkDelete`)
	* [x] `POST /api/v2/leads/add` - Bulk add (`lead.Service.BulkAdd`)
	* [x] `POST /api/v2/leads/bulk-assign` - Bulk assign (`lead.Service.BulkAssign`)
	* [x] `POST /api/v2/leads/move` - Move leads (`lead.Service.Move`)
	* [x] `POST /api/v2/leads/merge` - Merge leads (`lead.Service.Merge`)
	* [x] `POST /api/v2/leads/update-interest-status` - Update interest (`lead.Service.UpdateInterestStatus`)
	* [x] `POST /api/v2/leads/subsequence/remove` - Remove from subsequence (`lead.Service.RemoveFromSubsequence`)
	* [x] `POST /api/v2/leads/subsequence/move` - Move to subsequence (`lead.Service.MoveToSubsequence`)
* [x] **[Lead List API](https://developer.instantly.ai/api-reference/leadlist) — ([`leadlist/`](leadlist/leadlist.go))**
	* [x] `POST /api/v2/lead-lists` - Create (`leadlist.Service.Create`)
	* [x] `GET /api/v2/lead-lists` - List (`leadlist.Service.List` / `ListIter`)
	* [x] `GET /api/v2/lead-lists/{id}` - Get (`leadlist.Service.Get`)
	* [x] `PATCH /api/v2/lead-lists/{id}` - Patch (`leadlist.Service.Update`)
	* [x] `DELETE /api/v2/lead-lists/{id}` - Delete (`leadlist.Service.Delete`)
	* [x] `GET /api/v2/lead-lists/{id}/verification-stats` - Verification stats (`leadlist.Service.VerificationStats`)
* [x] **[Lead Label API](https://developer.instantly.ai/api-reference/leadlabel) — ([`leadlabel/`](leadlabel/leadlabel.go))**
	* [x] `POST /api/v2/lead-labels` - Create (`leadlabel.Service.Create`)
	* [x] `GET /api/v2/lead-labels` - List (`leadlabel.Service.List` / `ListIter`)
	* [x] `GET /api/v2/lead-labels/{id}` - Get (`leadlabel.Service.Get`)
	* [x] `PATCH /api/v2/lead-labels/{id}` - Patch (`leadlabel.Service.Update`)
	* [x] `DELETE /api/v2/lead-labels/{id}` - Delete (`leadlabel.Service.Delete`)
	* [x] `POST /api/v2/lead-labels/ai-reply-label` - Test AI reply label (`leadlabel.Service.TestAIReplyLabel`)
* [x] **[Campaign Subsequence API](https://developer.instantly.ai/api-reference/campaignsubsequence) — ([`subsequence/`](subsequence/subsequence.go))**
	* [x] `POST /api/v2/subsequences` - Create (`subsequence.Service.Create`)
	* [x] `GET /api/v2/subsequences` - List (`subsequence.Service.List` / `ListIter`)
	* [x] `GET /api/v2/subsequences/{id}` - Get (`subsequence.Service.Get`)
	* [x] `PATCH /api/v2/subsequences/{id}` - Patch (`subsequence.Service.Update`)
	* [x] `DELETE /api/v2/subsequences/{id}` - Delete (`subsequence.Service.Delete`)
	* [x] `POST /api/v2/subsequences/{id}/duplicate` - Duplicate (`subsequence.Service.Duplicate`)
	* [x] `POST /api/v2/subsequences/{id}/pause` - Pause (`subsequence.Service.Pause`)
	* [x] `POST /api/v2/subsequences/{id}/resume` - Resume (`subsequence.Service.Resume`)
	* [x] `GET /api/v2/subsequences/{id}/sending-status` - Sending status (`subsequence.Service.SendingStatus`)
* [x] **[Email Verification API](https://developer.instantly.ai/api-reference/emailverification) — ([`emailverification/`](emailverification/emailverification.go))**
	* [x] `POST /api/v2/email-verification` - Create verification (`emailverification.Service.Create`)
	* [x] `GET /api/v2/email-verification/{email}` - Check verification (`emailverification.Service.Check`)
* [x] **[Inbox Placement Test API](https://developer.instantly.ai/api-reference/inboxplacementtest) — ([`inboxtest/`](inboxtest/inboxtest.go))**
	* [x] `POST /api/v2/inbox-placement-tests` - Create test (`inboxtest.Service.Create`)
	* [x] `GET /api/v2/inbox-placement-tests` - List tests (`inboxtest.Service.List` / `ListIter`)
	* [x] `GET /api/v2/inbox-placement-tests/{id}` - Get test (`inboxtest.Service.Get`)
	* [x] `PATCH /api/v2/inbox-placement-tests/{id}` - Patch test (`inboxtest.Service.Update`)
	* [x] `DELETE /api/v2/inbox-placement-tests/{id}` - Delete test (`inboxtest.Service.Delete`)
	* [x] `GET /api/v2/inbox-placement-tests/email-service-provider-options` - ESP options (`inboxtest.Service.ESPOptions`)
* [x] **[Inbox Placement Analytics API](https://developer.instantly.ai/api-reference/inboxplacementanalytics) — ([`inboxanalytics/`](inboxanalytics/inboxanalytics.go))**
	* [x] `GET /api/v2/inbox-placement-analytics` - List analytics (`inboxanalytics.Service.List` / `ListIter`)
	* [x] `GET /api/v2/inbox-placement-analytics/{id}` - Get analytics (`inboxanalytics.Service.Get`)
	* [x] `POST /api/v2/inbox-placement-analytics/stats-by-test-id` - Stats by test id (`inboxanalytics.Service.StatsByTestID`)
	* [x] `POST /api/v2/inbox-placement-analytics/deliverability-insights` - Deliverability insights (`inboxanalytics.Service.DeliverabilityInsights`)
	* [x] `POST /api/v2/inbox-placement-analytics/stats-by-date` - Stats by date (`inboxanalytics.Service.StatsByDate`)
* [x] **[Inbox Placement Report API](https://developer.instantly.ai/api-reference/inboxplacementblacklistandspamassassinreport) — ([`inboxreport/`](inboxreport/inboxreport.go))**
	* [x] `GET /api/v2/inbox-placement-reports` - List reports (`inboxreport.Service.List` / `ListIter`)
	* [x] `GET /api/v2/inbox-placement-reports/{id}` - Get report (`inboxreport.Service.Get`)

**Planned coverage** — every remaining V2 resource, ordered by operation count, each linking to its
reference docs:

* [ ] **[SuperSearchEnrichment](https://developer.instantly.ai/api-reference/supersearchenrichment)** - 11 operations
* [ ] **[BlockListEntry](https://developer.instantly.ai/api-reference/blocklistentry)** - 9 operations
* [ ] **[Webhook](https://developer.instantly.ai/api-reference/webhook)** - 8 operations
* [ ] **[Workspace](https://developer.instantly.ai/api-reference/workspace)** - 8 operations
* [ ] **[DFYEmailAccountOrder](https://developer.instantly.ai/api-reference/dfyemailaccountorder)** - 7 operations
* [ ] **[CustomTag](https://developer.instantly.ai/api-reference/customtag)** - 6 operations
* [ ] **[WorkspaceGroupMember](https://developer.instantly.ai/api-reference/workspacegroupmember)** - 5 operations
* [ ] **[WorkspaceMember](https://developer.instantly.ai/api-reference/workspacemember)** - 5 operations
* [ ] **[WebhookEvent](https://developer.instantly.ai/api-reference/webhookevent)** - 4 operations
* [ ] **[APIKey](https://developer.instantly.ai/api-reference/apikey)** - 3 operations
* [ ] **[OAuth](https://developer.instantly.ai/api-reference/oauth)** - 3 operations
* [ ] **[BackgroundJob](https://developer.instantly.ai/api-reference/backgroundjob)** - 2 operations
* [ ] **[CRMActions](https://developer.instantly.ai/api-reference/crmactions)** - 2 operations
* [ ] **[WorkspaceBilling](https://developer.instantly.ai/api-reference/workspacebilling)** - 2 operations
* [ ] **[AuditLog](https://developer.instantly.ai/api-reference/auditlog)** - 1 operation
* [ ] **[CustomTagMapping](https://developer.instantly.ai/api-reference/customtagmapping)** - 1 operation

> **Notes on the counts.** The **Analytics** area (7 operations) is cross-cutting — 3 endpoints live
> under **Account** and 4 under **Campaign**, so they are counted within those resources rather than as
> a separate group. Three additional tags — **CustomPromptTemplate**, **EmailTemplate**, and
> **SalesFlow** — are declared in the V2 API but currently expose no public endpoints.

Adding a resource? Follow the durable recipe in
[`docs/adding-a-resource.md`](docs/adding-a-resource.md) — the `email` package is the reference
implementation.

</details>

<details>
<summary><strong><code>Client Configuration (Functional Options)</code></strong></summary>
<br/>

The client is **immutable after construction**: configure it once with functional options. This is
safer than mutating shared state and is the idiomatic Go pattern. Available options:

| Option | Purpose |
|--------|---------|
| `WithHTTPClient(*http.Client)` | Custom transport, proxy, or timeout |
| `WithBaseURL(string)` | Point at a test server or gateway |
| `WithUserAgent(string)` | Set the `User-Agent` header |
| `WithHTTPHeader(key, value)` | Add an extra header to every request |

```go
package main

import (
    "net/http"
    "time"

    "github.com/mrz1836/go-instantly"
)

// ....

client := instantly.NewClient("[API-KEY]",
    instantly.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
    instantly.WithUserAgent("my-app/1.0"),
)

// ...
```

Need an endpoint the typed resource packages do not wrap yet? The low-level plumbing is exported as an
escape hatch: `client.Get`, `client.Post`, `client.Patch`, `client.Put`, `client.Delete`, `client.Do`,
and `client.DoRaw` (raw bytes, for CSV and other non-JSON responses).
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
<summary><strong><code>GitHub Workflows</code></strong></summary>
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

A benchmark suite now covers the hot paths — query encoding and path building (`Query.Encode`,
`BuildPath`), the response-error probe (`checkResponse`), cursor iteration (`Iterate`), the reusable
list helpers (`JoinPath`, `ApplyOptions`), and a representative resource decode and round-trip
(`CampaignDecode`, `CampaignGet`). Run `magex bench` to measure them on your own hardware; numbers vary
by machine, so they are reproduced locally rather than pinned here.

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

[![License](https://img.shields.io/github/license/mrz1836/go-instantly.svg?style=flat&v=1)](LICENSE)
