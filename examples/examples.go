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
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/campaign"
	"github.com/mrz1836/go-instantly/email"
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
