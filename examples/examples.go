// Package main is an example of how to use the Instantly.ai V2 Go client.
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
	"os"
	"strings"

	"github.com/mrz1836/go-instantly"
)

func main() {
	// Authentication: a single V2 API key, sent as a bearer token on every
	// request. Keep it out of source control.
	apiKey := os.Getenv("INSTANTLY_API_KEY")
	if apiKey == "" {
		log.Fatal("INSTANTLY_API_KEY is required")
	}

	client := instantly.NewClient(apiKey)
	ctx := context.Background()

	if err := sendTestEmail(ctx, client); err != nil {
		log.Fatal(err)
	}

	if err := listEmails(ctx, client); err != nil {
		log.Fatal(err)
	}

	if err := iterateEmails(ctx, client); err != nil {
		log.Fatal(err)
	}

	if err := readEmail(ctx, client); err != nil {
		log.Fatal(err)
	}

	if err := replyAndForward(ctx, client); err != nil {
		log.Fatal(err)
	}

	if err := manageInbox(ctx, client); err != nil {
		log.Fatal(err)
	}
}

// sendTestEmail sends a test email and shows how to inspect a typed API error.
//
// This endpoint reports a sending-account failure inside an otherwise
// successful HTTP 200 response. The client converts that into a real error, so
// a nil return is the only signal of success, and the underlying code is
// available through errors.As.
func sendTestEmail(ctx context.Context, client *instantly.Client) error {
	err := client.SendTestEmail(ctx, instantly.SendTestEmailRequest{
		EAccount:           "sender@example.com",
		ToAddressEmailList: "recipient@example.com",
		Subject:            "Testing the sending account",
		Body: instantly.EmailBody{
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
func listEmails(ctx context.Context, client *instantly.Client) error {
	page, err := client.ListEmails(ctx,
		instantly.WithEmailLimit(25),
		instantly.WithEmailIsUnread(true),
		instantly.WithEmailMode(instantly.EmailModeFocused),
		instantly.WithEmailSortOrder(instantly.SortOrderDesc),
		instantly.WithEmailType(instantly.EmailTypeReceived),
	)
	if err != nil {
		return err
	}

	// Reading the emails themselves is shown in iterateEmails below.
	log.Printf("fetched %d emails", len(page.Items))

	// Pagination is cursor based. NextStartingAfter carries the cursor for the
	// following page and is empty once the last page has been reached. Pass it
	// back with WithEmailStartingAfter to page by hand, or use ListEmailsIter to
	// have the client walk the pages for you.
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
func iterateEmails(ctx context.Context, client *instantly.Client) error {
	count := 0

	for email, err := range client.ListEmailsIter(ctx,
		instantly.WithEmailLimit(100),
		instantly.WithEmailLatestOfThread(true),
	) {
		// Iteration stops at the first error, which arrives with a nil email.
		if err != nil {
			return err
		}

		count++

		log.Printf("email %s: %s", sanitize(email.ID), sanitize(email.Subject))

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
func readEmail(ctx context.Context, client *instantly.Client) error {
	email, err := client.GetEmail(ctx, "[EMAIL-ID]")
	if err != nil {
		return err
	}

	log.Printf("email %s from %s", sanitize(email.ID), sanitize(email.EAccount))

	// Fields the API declares as nullable are pointers, so an absent value
	// stays distinguishable from a zero value: a nil ContentPreview means the
	// API reported nothing, which is not the same as an empty preview.
	if email.ContentPreview != nil {
		log.Printf("preview: %s", sanitize(*email.ContentPreview))
	}

	// Only the fields set on the request are patched; an omitted field leaves
	// the current value unchanged.
	markAsRead := false

	if _, err = client.UpdateEmail(ctx, email.ID, instantly.UpdateEmailRequest{
		IsUnread: &markAsRead,
	}); err != nil {
		return err
	}

	log.Printf("email marked as read")

	return nil
}

// replyAndForward replies to an email and forwards it to another recipient.
func replyAndForward(ctx context.Context, client *instantly.Client) error {
	reply, err := client.ReplyToEmail(ctx, instantly.ReplyToEmailRequest{
		ReplyToUUID: "[EMAIL-ID]",
		EAccount:    "sender@example.com",
		Subject:     "Re: your question",
		Body: instantly.EmailBody{
			HTML: "<p>Thanks for reaching out.</p>",
			Text: "Thanks for reaching out.",
		},
		CCAddressEmailList: "teammate@example.com",
	})
	if err != nil {
		return err
	}

	log.Printf("reply sent: %s", sanitize(reply.ID))

	includeOriginal := true

	forward, err := client.ForwardEmail(ctx, instantly.ForwardEmailRequest{
		ReplyToUUID:         "[EMAIL-ID]",
		ToAddressEmailList:  "colleague@example.com",
		EAccount:            "sender@example.com",
		Subject:             "Fwd: your question",
		IncludeOriginalBody: &includeOriginal,
		Body: &instantly.EmailBody{
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
func manageInbox(ctx context.Context, client *instantly.Client) error {
	unread, err := client.CountUnreadEmails(ctx)
	if err != nil {
		return err
	}

	log.Printf("%d unread emails", unread)

	if err = client.MarkThreadAsRead(ctx, "[THREAD-ID]"); err != nil {
		return err
	}

	log.Printf("thread marked as read")

	// Deleting is permanent, and returns the email that was removed.
	deleted, err := client.DeleteEmail(ctx, "[EMAIL-ID]")
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
