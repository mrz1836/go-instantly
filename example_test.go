package instantly_test

import (
	"fmt"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/email"
)

// ExampleNewClient constructs a client once and hands it to a resource service.
func ExampleNewClient() {
	client := instantly.NewClient("my-v2-api-key",
		instantly.WithUserAgent("my-app/1.0"),
	)

	svc := email.New(client)
	_ = svc

	fmt.Println("email service ready")
	// Output: email service ready
}
