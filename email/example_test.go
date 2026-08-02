package email_test

import (
	"context"
	"fmt"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/email"
)

// ExampleService_List fetches a single page of emails, filtered with options.
func ExampleService_List() {
	svc := email.New(instantly.NewClient("my-v2-api-key"))

	page, err := svc.List(context.Background(),
		email.WithLimit(50),
		email.WithIsUnread(true),
	)
	if err != nil {
		fmt.Println("list failed:", err)
		return
	}

	for _, msg := range page.Items {
		fmt.Println(msg.Subject)
	}

	// NextStartingAfter carries the cursor for the following page, if any.
	fmt.Println("next cursor:", page.NextStartingAfter)
}

// ExampleService_ListIter walks every page of emails with a range-over-func
// loop, so the caller never handles pagination cursors directly.
func ExampleService_ListIter() {
	svc := email.New(instantly.NewClient("my-v2-api-key"))

	for msg, err := range svc.ListIter(context.Background(), email.WithIsUnread(true)) {
		if err != nil {
			fmt.Println("iteration failed:", err)
			break
		}

		fmt.Println(msg.Subject)
	}
}
