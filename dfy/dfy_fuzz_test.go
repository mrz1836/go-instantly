package dfy_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/dfy"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzDFYSerialization round trips arbitrary field values through every DFY
// request body, asserting the encoding never panics and never drifts.
func FuzzDFYSerialization(f *testing.F) {
	f.Add("example.com", "john.doe", "com")
	f.Add("", "", "")
	f.Add("Ünïcödé.com", "prefix+tag", "org")
	f.Add("line\r\nbreak.com", "\x00prefix", "co")

	f.Fuzz(func(t *testing.T, domain, prefix, tld string) {
		// Encoding coerces invalid UTF-8 to the replacement character, so exact
		// equality is only asserted for input the encoder can represent.
		lossless := utf8.ValidString(domain) && utf8.ValidString(prefix) && utf8.ValidString(tld)

		create := dfy.CreateRequest{
			OrderType:  dfy.OrderTypeDFY,
			Simulation: instantly.Ptr(true),
			Items: []dfy.OrderItem{{
				Domain:           domain,
				EmailProvider:    instantly.Ptr(dfy.EmailProviderGoogle),
				ForwardingDomain: domain,
				Accounts: []dfy.AccountSpec{{
					EmailAddressPrefix: prefix,
					FirstName:          prefix,
					LastName:           domain,
				}},
			}},
		}

		instantlytest.RequireStableRoundTrip(t, create, lossless)
		instantlytest.RequireStableRoundTrip(t, dfy.SimilarDomainsRequest{Domain: domain, TLDs: []string{tld}}, lossless)
		instantlytest.RequireStableRoundTrip(t, dfy.CheckDomainsRequest{Domains: []string{domain}}, lossless)
		instantlytest.RequireStableRoundTrip(
			t, dfy.PreWarmedUpDomainsRequest{Extensions: []string{tld}, Search: domain}, lossless,
		)
		instantlytest.RequireStableRoundTrip(t, dfy.CancelAccountsRequest{Accounts: []string{prefix}}, lossless)
	})
}

// FuzzDFYResponseDecoding feeds arbitrary bytes back as DFY responses, asserting
// the client never panics and never hands back a partly decoded value.
func FuzzDFYResponseDecoding(f *testing.F) {
	f.Add(orderResultSuccess)
	f.Add(orderResultFailure)
	f.Add(orderFixture)
	f.Add(orderFixtureNulls)
	f.Add(accountReady)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"results":[{"domain":"x.com","available":true}]}`)
	f.Add(`{"domains":["a.com"],"domains_with_type":[{"domain":"a.com","account_type":2}]}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"email_provider":"not a number"}`)
	f.Add(`{"price_per_account_per_month":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := dfy.New(instantlytest.FuzzClient(http.StatusOK, body))

		requireNil := func(v any, err error, msg string) {
			if err != nil {
				require.Nil(t, v, msg)
			} else {
				require.NotNil(t, v)
			}
		}

		order, err := svc.List(ctx)
		requireNil(order, err, "a decode failure must never hand back a partly populated page")

		accounts, err := svc.ListAccounts(ctx, true)
		requireNil(accounts, err, "a decode failure must never hand back a partly populated page")

		result, err := svc.Create(ctx, dfy.CreateRequest{})
		requireNil(result, err, "a decode failure must never hand back a partly populated result")

		similar, err := svc.GenerateSimilarDomains(ctx, dfy.SimilarDomainsRequest{})
		requireNil(similar, err, "a decode failure must never hand back a partly populated result")

		check, err := svc.CheckDomains(ctx, dfy.CheckDomainsRequest{})
		requireNil(check, err, "a decode failure must never hand back a partly populated result")

		preWarmed, err := svc.PreWarmedUpDomains(ctx, dfy.PreWarmedUpDomainsRequest{})
		requireNil(preWarmed, err, "a decode failure must never hand back a partly populated result")

		cancelled, err := svc.CancelAccounts(ctx, dfy.CancelAccountsRequest{})
		requireNil(cancelled, err, "a decode failure must never hand back a partly populated result")
	})
}
