// Package main is an example of how to use the Instantly.ai V2 Go SDK.
//
// The code below is illustrative: it is compiled and vetted, but it is never
// executed by CI, because every call would reach the live Instantly.ai API. Run
// it yourself only with a real V2 API key and a sending account you own.
//
// V2 credentials are distinct from V1 credentials. A V1 key will not
// authenticate against V2, so create a V2 key before running any of this.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/apikey"
	"github.com/mrz1836/go-instantly/auditlog"
	"github.com/mrz1836/go-instantly/backgroundjob"
	"github.com/mrz1836/go-instantly/blocklist"
	"github.com/mrz1836/go-instantly/campaign"
	"github.com/mrz1836/go-instantly/crm"
	"github.com/mrz1836/go-instantly/customtag"
	"github.com/mrz1836/go-instantly/dfy"
	"github.com/mrz1836/go-instantly/email"
	"github.com/mrz1836/go-instantly/emailverification"
	"github.com/mrz1836/go-instantly/inboxanalytics"
	"github.com/mrz1836/go-instantly/inboxtest"
	"github.com/mrz1836/go-instantly/oauth"
	"github.com/mrz1836/go-instantly/supersearch"
	"github.com/mrz1836/go-instantly/webhook"
	"github.com/mrz1836/go-instantly/webhookevent"
	"github.com/mrz1836/go-instantly/workspace"
	"github.com/mrz1836/go-instantly/workspacemember"
)

func main() {
	// Authentication: a single V2 API key, sent as a bearer token on every
	// request. Keep it out of source control.
	apiKey := os.Getenv("INSTANTLY_API_KEY")
	if apiKey == "" {
		log.Fatal("INSTANTLY_API_KEY is required")
	}

	// The client is immutable after construction: configure it once with
	// functional options. Here it gets a custom timeout and a user agent.
	client := instantly.NewClient(apiKey,
		instantly.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
		instantly.WithUserAgent("my-app/1.0"),
	)

	// Each resource lives in its own package, and every service shares the one
	// client. Build the ones you need.
	emails := email.New(client)
	campaigns := campaign.New(client)
	verifications := emailverification.New(client)
	inboxTests := inboxtest.New(client)
	inboxAnalytics := inboxanalytics.New(client)
	enrichment := supersearch.New(client)
	webhooks := webhook.New(client)
	webhookEvents := webhookevent.New(client)
	workspaces := workspace.New(client)
	members := workspacemember.New(client)
	blocked := blocklist.New(client)
	tags := customtag.New(client)
	keys := apikey.New(client)
	audit := auditlog.New(client)
	jobs := backgroundjob.New(client)
	phoneNumbers := crm.New(client)
	oauthSessions := oauth.New(client)
	dfyOrders := dfy.New(client)

	ctx := context.Background()

	if err := sendTestEmail(ctx, emails); err != nil {
		log.Fatal(err)
	}

	if err := listEmails(ctx, emails); err != nil {
		log.Fatal(err)
	}

	if err := iterateEmails(ctx, emails); err != nil {
		log.Fatal(err)
	}

	if err := readEmail(ctx, emails); err != nil {
		log.Fatal(err)
	}

	if err := replyAndForward(ctx, emails); err != nil {
		log.Fatal(err)
	}

	if err := manageInbox(ctx, emails); err != nil {
		log.Fatal(err)
	}

	if err := exploreCampaigns(ctx, campaigns); err != nil {
		log.Fatal(err)
	}

	if err := verifyEmail(ctx, verifications); err != nil {
		log.Fatal(err)
	}

	if err := inboxPlacement(ctx, inboxTests, inboxAnalytics); err != nil {
		log.Fatal(err)
	}

	if err := enrichLeads(ctx, enrichment); err != nil {
		log.Fatal(err)
	}

	if err := manageWebhooks(ctx, webhooks, webhookEvents); err != nil {
		log.Fatal(err)
	}

	if err := manageWorkspace(ctx, workspaces, members); err != nil {
		log.Fatal(err)
	}

	if err := curateBlockList(ctx, blocked, tags); err != nil {
		log.Fatal(err)
	}

	if err := manageAPIKeys(ctx, keys); err != nil {
		log.Fatal(err)
	}

	if err := reviewAuditLog(ctx, audit); err != nil {
		log.Fatal(err)
	}

	if err := trackBackgroundJobs(ctx, jobs); err != nil {
		log.Fatal(err)
	}

	if err := managePhoneNumbers(ctx, phoneNumbers); err != nil {
		log.Fatal(err)
	}

	if err := connectMailbox(ctx, oauthSessions); err != nil {
		log.Fatal(err)
	}

	if err := orderMailboxes(ctx, dfyOrders); err != nil {
		log.Fatal(err)
	}
}

// orderMailboxes checks a domain's availability, then simulates a Done-For-You
// order to get a price quote without placing it.
//
// Simulation runs the full validation and pricing without charging: inspect
// OrderIsValid and OrderError on the result. Email providers are typed, and
// prices the API declares nullable are pointers.
func orderMailboxes(ctx context.Context, dfyOrders *dfy.Service) error {
	availability, err := dfyOrders.CheckDomains(ctx, dfy.CheckDomainsRequest{
		Domains: []string{"example.com"},
	})
	if err != nil {
		return err
	}

	for _, result := range availability.Results {
		log.Printf("domain %s available: %t", sanitize(result.Domain), result.Available)
	}

	// Simulation returns a price quote without placing the order or charging.
	quote, err := dfyOrders.Create(ctx, dfy.CreateRequest{
		OrderType:  dfy.OrderTypeDFY,
		Simulation: instantly.Ptr(true),
		Items: []dfy.OrderItem{{
			Domain:        "example.com",
			EmailProvider: instantly.Ptr(dfy.EmailProviderGoogle),
			Accounts: []dfy.AccountSpec{{
				EmailAddressPrefix: "sales",
				FirstName:          "Sales",
				LastName:           "Team",
			}},
		}},
	})
	if err != nil {
		return err
	}

	if !quote.OrderIsValid {
		log.Printf("order would be rejected: %s", sanitize(quote.OrderError))
		return nil
	}

	log.Printf("order would cost $%.2f total", quote.TotalPrice)

	return nil
}

// connectMailbox starts a Google OAuth session and polls its status once.
//
// A session that ends in an error is not decoded into a status: the API
// delivers the OAuth error code inside an HTTP 200 body, and the client
// converts that into an *instantly.APIError, so the error code (for example
// access_denied) is read through errors.As.
func connectMailbox(ctx context.Context, oauthSessions *oauth.Service) error {
	session, err := oauthSessions.InitGoogle(ctx)
	if err != nil {
		return err
	}

	// Redirect the user to session.AuthURL, then poll until they finish.
	log.Printf("send the user to %s", sanitize(session.AuthURL))

	status, err := oauthSessions.SessionStatus(ctx, session.SessionID)
	if err != nil {
		// A session that ended in an error surfaces as an APIError carrying the
		// OAuth error code, delivered at HTTP 200.
		var apiErr *instantly.APIError
		if errors.As(err, &apiErr) {
			log.Printf("oauth session failed: %s", sanitize(apiErr.Code))
		}

		return err
	}

	log.Printf("session %s is %s", sanitize(session.SessionID), sanitize(string(status.Status)))

	return nil
}

// managePhoneNumbers lists the phone numbers the organization owns and releases
// one.
//
// The list endpoint answers with a bare array, so the whole set comes back at
// once. Price is a pointer because it is nullable and omitted from a delete
// response.
func managePhoneNumbers(ctx context.Context, phoneNumbers *crm.Service) error {
	numbers, err := phoneNumbers.ListPhoneNumbers(ctx)
	if err != nil {
		return err
	}

	log.Printf("the organization owns %d phone numbers", len(numbers))

	if len(numbers) == 0 {
		return nil
	}

	deleted, err := phoneNumbers.DeletePhoneNumber(ctx, numbers[0].ID)
	if err != nil {
		return err
	}

	log.Printf("released phone number %s", sanitize(deleted.PhoneNumber))

	return nil
}

// trackBackgroundJobs lists the running jobs and reads one, selecting just the
// data fields it cares about.
//
// Job types and statuses are typed; the status filter takes a raw string
// because a single request can filter on several statuses at once. The data
// payload is free-form, so WithDataFields narrows it to the fields you need.
func trackBackgroundJobs(ctx context.Context, jobs *backgroundjob.Service) error {
	page, err := jobs.List(ctx,
		backgroundjob.WithType(backgroundjob.TypeMoveLeads),
		backgroundjob.WithStatus("pending,in-progress"),
		backgroundjob.WithSortColumn(backgroundjob.SortColumnCreatedAt),
		backgroundjob.WithSortOrder(instantly.SortOrderDesc),
	)
	if err != nil {
		return err
	}

	log.Printf("%d lead-move jobs are still running", len(page.Items))

	if len(page.Items) == 0 {
		return nil
	}

	job, err := jobs.Get(ctx, page.Items[0].ID,
		backgroundjob.WithDataFields("success_count,failed_count,total_to_process"),
	)
	if err != nil {
		return err
	}

	log.Printf("job %s is %.0f%% complete", sanitize(job.ID), job.Progress)

	return nil
}

// reviewAuditLog lists recent login activity from the workspace audit log.
//
// Activity types are typed named constants, and the date filters take a
// time.Time formatted to the wire for you. Nullable fields such as UserName are
// pointers, so an absent value stays distinguishable from an empty one.
func reviewAuditLog(ctx context.Context, audit *auditlog.Service) error {
	page, err := audit.List(ctx,
		auditlog.WithActivityType(auditlog.ActivityTypeUserLogin),
		auditlog.WithStartDate(time.Now().AddDate(0, 0, -7)),
		auditlog.WithLimit(100),
	)
	if err != nil {
		return err
	}

	log.Printf("fetched %d login events in the last week", len(page.Items))

	for _, record := range page.Items {
		if record.UserName != nil {
			log.Printf("login by %s from %s", sanitize(*record.UserName), sanitize(record.IPAddress))
		}
	}

	return nil
}

// manageAPIKeys creates a scoped API key, lists the workspace keys, and revokes
// one.
//
// Scopes are typed named constants (apikey.ScopeCampaignsRead rather than a
// bare string), so an invalid scope is a compile error. The full secret Key is
// only returned when the key is first created.
func manageAPIKeys(ctx context.Context, keys *apikey.Service) error {
	created, err := keys.Create(ctx, apikey.CreateRequest{
		Name: "CI deploy key",
		Scopes: []apikey.Scope{
			apikey.ScopeCampaignsRead,
			apikey.ScopeEmailsRead,
		},
	})
	if err != nil {
		return err
	}

	// The full token is exposed only here — store it now.
	log.Printf("created API key %s with %d scopes", sanitize(created.ID), len(created.Scopes))

	page, err := keys.List(ctx, apikey.WithLimit(50))
	if err != nil {
		return err
	}

	log.Printf("workspace has %d API keys", len(page.Items))

	deleted, err := keys.Delete(ctx, created.ID)
	if err != nil {
		return err
	}

	log.Printf("revoked API key %s", sanitize(deleted.ID))

	return nil
}

// exploreCampaigns lists campaigns and reads one, showing the typed options and
// the time helpers a resource exposes.
//
// Enum-like filters are typed (campaign.StatusActive rather than a bare int),
// date filters take a time.Time and are formatted to the wire for you, and the
// string timestamp fields carry a Parsed accessor so a decoded value still
// re-encodes byte-for-byte.
func exploreCampaigns(ctx context.Context, campaigns *campaign.Service) error {
	page, err := campaigns.List(ctx,
		campaign.WithLimit(25),
		campaign.WithStatus(campaign.StatusActive),
	)
	if err != nil {
		return err
	}

	log.Printf("fetched %d active campaigns", len(page.Items))

	// TimestampCreated stays a string on the model; parse it only when a
	// time.Time is what you need.
	if len(page.Items) > 0 {
		if created, perr := page.Items[0].ParsedTimestampCreated(); perr == nil {
			log.Printf("first campaign created %s", created.Format(time.RFC3339))
		}
	}

	// Analytics date filters accept a time.Time directly.
	daily, err := campaigns.DailyAnalytics(ctx,
		campaign.WithStartDate(time.Now().AddDate(0, -1, 0)),
		campaign.WithEndDate(time.Now()),
	)
	if err != nil {
		return err
	}

	log.Printf("got %d days of campaign analytics", len(daily))

	return nil
}

// verifyEmail submits an address for verification and polls a pending result.
//
// Verification can be asynchronous: an address that takes longer than ten
// seconds comes back with a StatusPending result, which Check then polls to
// completion. Setting a WebhookURL on the request receives the result instead.
func verifyEmail(ctx context.Context, verifications *emailverification.Service) error {
	result, err := verifications.Create(ctx, emailverification.CreateRequest{
		Email: "prospect@example.com",
	})
	if err != nil {
		return err
	}

	// A pending status is a normal result, not an error: the verification is
	// still running. Read VerificationStatus, not the request-level Status.
	if result.VerificationStatus == emailverification.StatusPending {
		if result, err = verifications.Check(ctx, result.Email); err != nil {
			return err
		}
	}

	log.Printf("verification for %s: %s", sanitize(result.Email), sanitize(string(result.VerificationStatus)))

	return nil
}

// inboxPlacement lists inbox placement tests, reads the provider options a test
// can target, creates a one-time test, and pulls its placement analytics.
//
// Placement analytics require a test id, so List takes it as a positional
// argument rather than an option.
func inboxPlacement(
	ctx context.Context, tests *inboxtest.Service, analytics *inboxanalytics.Service,
) error {
	page, err := tests.List(ctx, inboxtest.WithStatus(inboxtest.StatusActive))
	if err != nil {
		return err
	}

	log.Printf("fetched %d active inbox placement tests", len(page.Items))

	// The provider/audience combinations a test can send to come from ESPOptions.
	options, err := tests.ESPOptions(ctx)
	if err != nil {
		return err
	}

	log.Printf("there are %d ESP options to choose from", len(options))

	created, err := tests.Create(ctx, inboxtest.CreateRequest{
		Name:             "Weekly deliverability check",
		Type:             inboxtest.TypeOneTime,
		SendingMethod:    inboxtest.SendingFromInstantly,
		EmailSubject:     "Are we landing in the inbox?",
		EmailBody:        "<p>Seed message.</p>",
		Emails:           []string{"sender@example.com"},
		RecipientsLabels: options,
	})
	if err != nil {
		return err
	}

	log.Printf("created inbox placement test %s", sanitize(created.ID))

	events, err := analytics.List(ctx, created.ID, inboxanalytics.WithLimit(100))
	if err != nil {
		return err
	}

	log.Printf("test %s has %d placement events so far", sanitize(created.ID), len(events.Items))

	return nil
}

// enrichLeads counts, previews, and enriches leads from a SuperSearch query.
//
// The SuperSearch query DSL is deeply nested and free-form, so it is passed as a
// json.RawMessage and preserved verbatim: build it however you like — a struct
// you marshal, or a raw literal as shown here.
func enrichLeads(ctx context.Context, enrichment *supersearch.Service) error {
	filters := json.RawMessage(`{"department":["engineering"],"title":["cto","vp engineering"]}`)

	// Count first to size the job, then preview a sample before committing to it.
	count, err := enrichment.CountLeads(ctx, supersearch.SearchRequest{
		SearchFilters:  filters,
		SkipOwnedLeads: instantly.Ptr(true),
	})
	if err != nil {
		return err
	}

	log.Printf("the query matches %.0f leads", count.NumberOfLeads)

	preview, err := enrichment.PreviewLeads(ctx, supersearch.SearchRequest{SearchFilters: filters})
	if err != nil {
		return err
	}

	for _, lead := range preview.Leads {
		log.Printf("preview: %s at %s", sanitize(lead.FullName), sanitize(lead.CompanyName))
	}

	// Enriching the leads pulls them into a list and returns the enrichment.
	enriched, err := enrichment.EnrichLeads(ctx, supersearch.EnrichLeadsRequest{
		SearchFilters:       filters,
		Limit:               100,
		ListName:            "SuperSearch — engineering leaders",
		WorkEmailEnrichment: instantly.Ptr(true),
	})
	if err != nil {
		return err
	}

	log.Printf("started enrichment %s", sanitize(enriched.ID))

	return nil
}

// manageWebhooks creates a webhook, sends a test delivery, and reviews the
// recent delivery events and their aggregate success rate.
//
// A webhook disabled by repeated delivery failures reads back with a Status of
// StatusError; webhooks.Resume reactivates it. The custom headers a webhook
// sends are a free-form map, passed as a json.RawMessage.
func manageWebhooks(
	ctx context.Context, webhooks *webhook.Service, events *webhookevent.Service,
) error {
	hook, err := webhooks.Create(ctx, webhook.CreateRequest{
		TargetHookURL: "https://example.com/instantly-hook",
		Name:          instantly.Ptr("Reply notifier"),
		EventType:     webhook.EventReplyReceived,
		Headers:       json.RawMessage(`{"Authorization":"Bearer [WEBHOOK-SECRET]"}`),
	})
	if err != nil {
		return err
	}

	log.Printf("created webhook %s", sanitize(hook.ID))

	// A test delivery reports the target's status without waiting for a real
	// event to fire.
	result, err := webhooks.Test(ctx, hook.ID)
	if err != nil {
		return err
	}

	log.Printf("test delivery succeeded: %t", result.Success)

	// Review the recent failed deliveries and the overall success rate.
	failures, err := events.List(ctx, webhookevent.WithSuccess(false), webhookevent.WithLimit(20))
	if err != nil {
		return err
	}

	log.Printf("%d recent failed deliveries", len(failures.Items))

	summary, err := events.Summary(ctx, "", "")
	if err != nil {
		return err
	}

	log.Printf("overall delivery success rate: %.1f%%", summary.SuccessRate*100)

	return nil
}

// manageWorkspace reads the current workspace and invites a member to it.
//
// The Workspace API operates on a single workspace — the one the API key
// authenticates against — so its methods take no workspace id. Roles and
// permissions are typed, so an invalid value is a compile error rather than a
// rejected request.
func manageWorkspace(
	ctx context.Context, workspaces *workspace.Service, members *workspacemember.Service,
) error {
	current, err := workspaces.Get(ctx)
	if err != nil {
		return err
	}

	log.Printf("workspace %s owned by %s", sanitize(current.Name), sanitize(current.Owner))

	member, err := members.Create(ctx, workspacemember.CreateRequest{
		Email: "teammate@example.com",
		Role:  workspacemember.RoleEditor,
		Permissions: []workspacemember.Permission{
			workspacemember.PermissionCampaignsView,
			workspacemember.PermissionUniboxAll,
		},
	})
	if err != nil {
		return err
	}

	log.Printf("invited member %s as %s", sanitize(member.Email), sanitize(string(member.Role)))

	return nil
}

// curateBlockList bulk-adds block list entries, downloads the list as CSV, and
// creates a custom tag to organize resources.
//
// The download endpoint answers with text/csv rather than JSON, so Download
// hands back the raw bytes for you to write to a file or parse yourself.
func curateBlockList(ctx context.Context, blocked *blocklist.Service, tags *customtag.Service) error {
	created, err := blocked.BulkCreate(ctx, blocklist.BulkCreateRequest{
		BLValues: []string{"spam.example.com", "noreply@example.com"},
	})
	if err != nil {
		return err
	}

	log.Printf("blocked %.0f values (%.0f invalid)", created.ValidCount, created.InvalidCount)

	// Download the whole list as CSV — the raw bytes are returned unchanged.
	csv, err := blocked.Download(ctx, false, "")
	if err != nil {
		return err
	}

	log.Printf("downloaded %d bytes of block list CSV", len(csv))

	// Custom tags organize accounts and campaigns; a nil error is success.
	tag, err := tags.Create(ctx, customtag.CreateRequest{
		Label:       "Cold outreach",
		Description: instantly.Ptr("Accounts reserved for cold campaigns"),
	})
	if err != nil {
		return err
	}

	log.Printf("created tag %s", sanitize(tag.Label))

	return nil
}

// sendTestEmail sends a test email and shows how to inspect a typed API error.
//
// This endpoint reports a sending-account failure inside an otherwise
// successful HTTP 200 response. The client converts that into a real error, so
// a nil return is the only signal of success, and the underlying code is
// available through errors.As.
func sendTestEmail(ctx context.Context, emails *email.Service) error {
	err := emails.SendTest(ctx, email.SendTestRequest{
		EAccount:           "sender@example.com",
		ToAddressEmailList: "recipient@example.com",
		Subject:            "Testing the sending account",
		Body: email.Body{
			HTML: "<p>Hello from go-instantly.</p>",
			Text: "Hello from go-instantly.",
		},
	})
	if err != nil {
		var apiErr *instantly.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.Code {
			case instantly.ErrCodeAccountAuthError:
				log.Printf("the sending account failed to authenticate")
			case instantly.ErrCodeAccountNotFound:
				log.Printf("the sending account does not exist")
			case instantly.ErrCodeAccountUnknownError:
				log.Printf("the sending account failed for an unknown reason")
			default:
				log.Printf("api error %d: %s", apiErr.StatusCode, apiErr.Code)
			}
		}

		return err
	}

	log.Printf("test email accepted")

	return nil
}

// listEmails fetches a single page of emails using functional options.
//
// Only the options actually passed are sent as query parameters, so filters
// left out never reach the API as empty values.
func listEmails(ctx context.Context, emails *email.Service) error {
	page, err := emails.List(ctx,
		email.WithLimit(25),
		email.WithIsUnread(true),
		email.WithMode(email.ModeFocused),
		email.WithSortOrder(instantly.SortOrderDesc),
		email.WithType(email.TypeReceived),
	)
	if err != nil {
		return err
	}

	// Reading the emails themselves is shown in iterateEmails below.
	log.Printf("fetched %d emails", len(page.Items))

	// Pagination is cursor based. NextStartingAfter carries the cursor for the
	// following page and is empty once the last page has been reached. Pass it
	// back with email.WithStartingAfter to page by hand, or use ListIter to have
	// the client walk the pages for you.
	if page.NextStartingAfter != "" {
		log.Printf("next page cursor: %s", sanitize(page.NextStartingAfter))
	}

	return nil
}

// iterateEmails walks every page of results with the range-over-func iterator.
//
// Each page is a separate request against an endpoint rate limited to 20
// requests per minute, so narrow the result set with options rather than
// walking the whole mailbox.
func iterateEmails(ctx context.Context, emails *email.Service) error {
	count := 0

	for message, err := range emails.ListIter(ctx,
		email.WithLimit(100),
		email.WithLatestOfThread(true),
	) {
		// Iteration stops at the first error, which arrives with a nil email.
		if err != nil {
			return err
		}

		count++

		log.Printf("email %s: %s", sanitize(message.ID), sanitize(message.Subject))

		// Breaking out stops iteration immediately, and issues no further
		// requests.
		if count >= 500 {
			break
		}
	}

	log.Printf("walked %d emails", count)

	return nil
}

// readEmail fetches a single email and marks it as read.
func readEmail(ctx context.Context, emails *email.Service) error {
	message, err := emails.Get(ctx, "[EMAIL-ID]")
	if err != nil {
		return err
	}

	log.Printf("email %s from %s", sanitize(message.ID), sanitize(message.EAccount))

	// Fields the API declares as nullable are pointers, so an absent value stays
	// distinguishable from a zero value: a nil ContentPreview means the API
	// reported nothing, which is not the same as an empty preview.
	if message.ContentPreview != nil {
		log.Printf("preview: %s", sanitize(*message.ContentPreview))
	}

	// Only the fields set on the request are patched; an omitted field leaves the
	// current value unchanged. instantly.Ptr builds the pointer inline.
	if _, err = emails.Update(ctx, message.ID, email.UpdateRequest{
		IsUnread: instantly.Ptr(false),
	}); err != nil {
		return err
	}

	log.Printf("email marked as read")

	return nil
}

// replyAndForward replies to an email and forwards it to another recipient.
func replyAndForward(ctx context.Context, emails *email.Service) error {
	reply, err := emails.Reply(ctx, email.ReplyRequest{
		ReplyToUUID: "[EMAIL-ID]",
		EAccount:    "sender@example.com",
		Subject:     "Re: your question",
		Body: email.Body{
			HTML: "<p>Thanks for reaching out.</p>",
			Text: "Thanks for reaching out.",
		},
		CCAddressEmailList: "teammate@example.com",
	})
	if err != nil {
		return err
	}

	log.Printf("reply sent: %s", sanitize(reply.ID))

	forward, err := emails.Forward(ctx, email.ForwardRequest{
		ReplyToUUID:         "[EMAIL-ID]",
		ToAddressEmailList:  "colleague@example.com",
		EAccount:            "sender@example.com",
		Subject:             "Fwd: your question",
		IncludeOriginalBody: instantly.Ptr(true),
		Body: &email.Body{
			HTML: "<p>Passing this along.</p>",
			Text: "Passing this along.",
		},
	})
	if err != nil {
		return err
	}

	log.Printf("forward sent: %s", sanitize(forward.ID))

	return nil
}

// manageInbox counts unread emails, clears a thread, and deletes an email.
func manageInbox(ctx context.Context, emails *email.Service) error {
	unread, err := emails.CountUnread(ctx)
	if err != nil {
		return err
	}

	log.Printf("%d unread emails", unread)

	if err = emails.MarkThreadAsRead(ctx, "[THREAD-ID]"); err != nil {
		return err
	}

	log.Printf("thread marked as read")

	// Deleting is permanent, and returns the email that was removed.
	deleted, err := emails.Delete(ctx, "[EMAIL-ID]")
	if err != nil {
		return err
	}

	log.Printf("deleted email %s", sanitize(deleted.ID))

	return nil
}

// sanitize strips line breaks from a value before it is logged.
//
// Every string in an API response is remote-controlled, and writing one to a
// log verbatim lets it forge additional log lines. Sample code gets copied, so
// it sanitizes rather than trusting the API.
func sanitize(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.ReplaceAll(value, "\r", " ")
}
