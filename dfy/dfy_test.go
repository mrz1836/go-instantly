package dfy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/dfy"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns for the DFY endpoints. None carries a path parameter.
const (
	ordersPath      = "/api/v2/dfy-email-account-orders"
	accountsPath    = "/api/v2/dfy-email-account-orders/accounts"
	cancelPath      = "/api/v2/dfy-email-account-orders/accounts/cancel"
	similarPath     = "/api/v2/dfy-email-account-orders/domains/similar"
	checkPath       = "/api/v2/dfy-email-account-orders/domains/check"
	preWarmedUpPath = "/api/v2/dfy-email-account-orders/domains/pre-warmed-up-list"
)

// orderFixture is a spec-shaped order with every nullable field populated.
const orderFixture = `{
	"workspace_id": "ws-1",
	"domain": "example.com",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"forwarding_domain": "forward.example.com",
	"forwarding_mode": "redirect",
	"is_pre_warmed_up": true,
	"timestamp_cancelled": "2026-08-02T10:00:00.000Z"
}`

// orderFixtureNulls is a spec-shaped order with every nullable field null.
const orderFixtureNulls = `{
	"workspace_id": "ws-1",
	"domain": "bare.com",
	"timestamp_created": "2026-08-01T11:00:00.000Z",
	"forwarding_domain": null,
	"forwarding_mode": null,
	"is_pre_warmed_up": null,
	"timestamp_cancelled": null
}`

// accountReady is a spec-shaped ordered account carrying a password.
const accountReady = `{
	"id": "acc-1",
	"domain": "example.com",
	"email": "john@example.com",
	"email_provider": 1,
	"first_name": "John",
	"last_name": "Doe",
	"is_pre_warmed_up": false,
	"timestamp_cancelled": "",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"password": "secret"
}`

// accountBare is a spec-shaped ordered account without a password.
const accountBare = `{
	"id": "acc-2",
	"domain": "example.com",
	"email": "jane@example.com",
	"email_provider": 3,
	"first_name": "Jane",
	"last_name": "Roe",
	"is_pre_warmed_up": true,
	"timestamp_cancelled": "",
	"timestamp_created": "2026-08-01T11:00:00.000Z"
}`

// orderResultSuccess is a placed order with nested items and a null legacy price.
const orderResultSuccess = `{
	"order_placed": true,
	"order_is_valid": true,
	"simulation": false,
	"unavailable_domains": [],
	"blacklist_domains": [],
	"invalid_domains": [],
	"invalid_forwarding_domains": [],
	"invalid_accounts": [],
	"missing_domain_orders": [],
	"provider_mismatch_domains": [],
	"unsupported_provider_domains": [],
	"unavailable_email_providers": [],
	"domains_without_accounts": [],
	"free_domains": [],
	"number_of_domains_ordered": 1,
	"number_of_accounts_ordered": 2,
	"price_per_account_per_month": null,
	"price_per_account_per_month_by_account_type": {"1": 5, "2": 4},
	"price_per_domain_per_year": 100,
	"total_domains_price_per_year": 100,
	"total_accounts_price_per_month": 10,
	"total_price_per_month": 10,
	"total_price_per_year": 100,
	"total_price": 110,
	"total_discount": 0,
	"order_items": [
		{
			"domain": "example.com",
			"accounts": [{"email_address_prefix": "john.doe", "first_name": "John", "last_name": "Doe"}],
			"email_provider": 1,
			"domain_price": 100,
			"domain_monthly_price": null,
			"accounts_price": 10,
			"total_price": 110,
			"total_discount": 0
		}
	],
	"payment_method_last_4_digits": "1234",
	"payment_method_brand": "Visa",
	"payment_method_name_on_card": "John Doe"
}`

// orderResultFailure is a rejected order carrying the order_error discriminator.
const orderResultFailure = `{
	"order_placed": false,
	"order_is_valid": false,
	"simulation": false,
	"order_error": "unavailable_domains",
	"unavailable_domains": ["taken.com"],
	"blacklist_domains": [],
	"invalid_domains": [],
	"invalid_forwarding_domains": [],
	"invalid_accounts": [
		{"domain": "bad.com", "email": "x@bad.com", "first_name": "X", "last_name": "Y", "reason": "First name is required"}
	],
	"missing_domain_orders": [],
	"provider_mismatch_domains": [],
	"unsupported_provider_domains": [],
	"unavailable_email_providers": [2],
	"domains_without_accounts": [],
	"free_domains": [],
	"number_of_domains_ordered": 0,
	"number_of_accounts_ordered": 0,
	"price_per_account_per_month": 5,
	"price_per_domain_per_year": 0,
	"total_domains_price_per_year": 0,
	"total_accounts_price_per_month": 0,
	"total_price_per_month": 0,
	"total_price_per_year": 0,
	"total_price": 0,
	"total_discount": 0,
	"order_items": [],
	"payment_method_last_4_digits": "",
	"payment_method_brand": "",
	"payment_method_name_on_card": ""
}`

// DFYTestSuite exercises the DFY Email Account Order API service against the mock
// router.
type DFYTestSuite struct {
	instantlytest.Suite
}

// TestDFYSuite runs the DFY Email Account Order API suite.
func TestDFYSuite(t *testing.T) {
	suite.Run(t, new(DFYTestSuite))
}

// TestCreate verifies the request body reaches the API and a placed order
// decodes, including nested items and the nullable prices.
func (s *DFYTestSuite) TestCreate() {
	var gotBody []byte

	s.Router.Post(ordersPath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		gotBody = body

		_, _ = w.Write([]byte(orderResultSuccess))
	})

	result, err := s.svc().Create(context.Background(), dfy.CreateRequest{
		OrderType:  dfy.OrderTypeDFY,
		Simulation: instantly.Ptr(true),
		Items: []dfy.OrderItem{{
			Domain:        "example.com",
			EmailProvider: instantly.Ptr(dfy.EmailProviderGoogle),
			Accounts: []dfy.AccountSpec{{
				EmailAddressPrefix: "john.doe",
				FirstName:          "John",
				LastName:           "Doe",
			}},
		}},
	})

	s.Require().NoError(err)

	// Assert on the request body after the call, off the server goroutine.
	var sent dfy.CreateRequest
	s.Require().NoError(json.Unmarshal(gotBody, &sent))
	s.Equal(dfy.OrderTypeDFY, sent.OrderType)
	s.Require().Len(sent.Items, 1)
	s.Equal("example.com", sent.Items[0].Domain)
	s.Require().NotNil(sent.Items[0].EmailProvider)
	s.Equal(dfy.EmailProviderGoogle, *sent.Items[0].EmailProvider)
	s.Require().Len(sent.Items[0].Accounts, 1)
	s.Equal("john.doe", sent.Items[0].Accounts[0].EmailAddressPrefix)
	s.Require().NotNil(sent.Simulation)
	s.True(*sent.Simulation)

	s.True(result.OrderPlaced)
	s.True(result.OrderIsValid)
	s.Nil(result.PricePerAccountPerMonth, "a null legacy price must decode to nil")
	s.JSONEq(`{"1":5,"2":4}`, string(result.PricePerAccountPerMonthByAccountType))
	s.InEpsilon(110.0, result.TotalPrice, 1e-9)

	s.Require().Len(result.OrderItems, 1)
	s.Equal("example.com", result.OrderItems[0].Domain)
	s.Nil(result.OrderItems[0].DomainMonthlyPrice, "a null per-domain price must decode to nil")
	s.Require().Len(result.OrderItems[0].Accounts, 1)
	s.Equal("john.doe", result.OrderItems[0].Accounts[0].EmailAddressPrefix)
	s.Equal("Visa", result.PaymentMethodBrand)
}

// TestCreateFailureShaped verifies a rejected order is a normal response, not an
// error: the order_error discriminator and its detail fields decode.
func (s *DFYTestSuite) TestCreateFailureShaped() {
	s.Router.Post(ordersPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(orderResultFailure))
	})

	result, err := s.svc().Create(context.Background(), dfy.CreateRequest{
		OrderType: dfy.OrderTypeDFY,
		Items:     []dfy.OrderItem{{Domain: "taken.com"}},
	})

	// order_error is a normal field (not the top-level `error` the client turns
	// into an APIError), so a rejected order still returns a nil error.
	s.Require().NoError(err)
	s.False(result.OrderPlaced)
	s.False(result.OrderIsValid)
	s.Equal("unavailable_domains", result.OrderError)
	s.Equal([]string{"taken.com"}, result.UnavailableDomains)
	s.Equal([]dfy.EmailProvider{dfy.EmailProviderAirMail}, result.UnavailableEmailProviders)
	s.Require().Len(result.InvalidAccounts, 1)
	s.Equal("First name is required", result.InvalidAccounts[0].Reason)
	s.Require().NotNil(result.PricePerAccountPerMonth)
	s.InEpsilon(5.0, *result.PricePerAccountPerMonth, 1e-9)
}

// TestList verifies a page of orders decodes and the options are sent.
func (s *DFYTestSuite) TestList() {
	s.Router.Get(ordersPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(ordersPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{orderFixture}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(), dfy.WithLimit(50))

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)
	s.Equal("cursor-2", page.NextStartingAfter)

	got := page.Items[0]
	s.Equal("example.com", got.Domain)
	s.Require().NotNil(got.ForwardingDomain)
	s.Equal("forward.example.com", *got.ForwardingDomain)
	s.Require().NotNil(got.ForwardingMode)
	s.Equal(dfy.ForwardingModeRedirect, *got.ForwardingMode)
	s.Require().NotNil(got.IsPreWarmedUp)
	s.True(*got.IsPreWarmedUp)
	s.Require().NotNil(got.TimestampCancelled)
}

// TestListNulls verifies an order with every nullable field null decodes to nil
// pointers rather than zero values.
func (s *DFYTestSuite) TestListNulls() {
	s.Router.Get(ordersPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(instantlytest.Page([]string{orderFixtureNulls}, "")))
	})

	page, err := s.svc().List(context.Background())

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)

	got := page.Items[0]
	s.Equal("bare.com", got.Domain)
	s.Nil(got.ForwardingDomain)
	s.Nil(got.ForwardingMode)
	s.Nil(got.IsPreWarmedUp)
	s.Nil(got.TimestampCancelled)
}

// TestListWithoutOptions verifies an unfiltered order list sends no query string.
func (s *DFYTestSuite) TestListWithoutOptions() {
	s.Router.Get(ordersPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestListAccountsWithPasswords verifies with_passwords is sent as true and the
// password decodes.
func (s *DFYTestSuite) TestListAccountsWithPasswords() {
	s.Router.Get(accountsPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("true", req.URL.Query().Get("with_passwords"))
		s.Equal("25", req.URL.Query().Get("limit"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{accountReady}, "")))
	})

	page, err := s.svc().ListAccounts(context.Background(), true, dfy.WithLimit(25))

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)

	got := page.Items[0]
	s.Equal("acc-1", got.ID)
	s.Equal(dfy.EmailProviderGoogle, got.EmailProvider)
	s.Require().NotNil(got.Password)
	s.Equal("secret", *got.Password)
}

// TestListAccountsWithoutPasswords verifies with_passwords is sent as false and
// the omitted password decodes to nil.
func (s *DFYTestSuite) TestListAccountsWithoutPasswords() {
	s.Router.Get(accountsPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("false", req.URL.Query().Get("with_passwords"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{accountBare}, "")))
	})

	page, err := s.svc().ListAccounts(context.Background(), false)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)

	got := page.Items[0]
	s.Equal("acc-2", got.ID)
	s.Equal(dfy.EmailProviderMicrosoft, got.EmailProvider)
	s.Nil(got.Password, "a missing password must decode to nil")
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *DFYTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option dfy.ListOption
		key    string
		value  string
	}{
		{"limit", dfy.WithLimit(50), "limit", "50"},
		{"starting after", dfy.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
	}

	s.Require().Len(tests, 2, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestGenerateSimilarDomains verifies the request body reaches the API and the
// suggestions decode.
func (s *DFYTestSuite) TestGenerateSimilarDomains() {
	s.Router.Post(similarPath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)

		var got dfy.SimilarDomainsRequest
		s.NoError(json.Unmarshal(body, &got))
		s.Equal("example.com", got.Domain)
		s.Equal([]string{"com", "org"}, got.TLDs)

		_, _ = w.Write([]byte(`{"domains":["acme.com","acme.org"]}`))
	})

	result, err := s.svc().GenerateSimilarDomains(context.Background(), dfy.SimilarDomainsRequest{
		Domain: "example.com",
		TLDs:   []string{"com", "org"},
	})

	s.Require().NoError(err)
	s.Equal([]string{"acme.com", "acme.org"}, result.Domains)
}

// TestCheckDomains verifies the request body reaches the API and the
// availability results decode, including the nullable restriction fields.
func (s *DFYTestSuite) TestCheckDomains() {
	s.Router.Post(checkPath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)

		var got dfy.CheckDomainsRequest
		s.NoError(json.Unmarshal(body, &got))
		s.Equal([]string{"example.com", "equifax.com"}, got.Domains)

		_, _ = w.Write([]byte(`{"results":[` +
			`{"domain":"example.com","available":true},` +
			`{"domain":"equifax.com","available":false,"unavailable_reason":"restricted","restricted_keyword":"equifax"}` +
			`]}`))
	})

	result, err := s.svc().CheckDomains(context.Background(), dfy.CheckDomainsRequest{
		Domains: []string{"example.com", "equifax.com"},
	})

	s.Require().NoError(err)
	s.Require().Len(result.Results, 2)

	s.True(result.Results[0].Available)
	s.Nil(result.Results[0].UnavailableReason)

	s.False(result.Results[1].Available)
	s.Require().NotNil(result.Results[1].UnavailableReason)
	s.Equal("restricted", *result.Results[1].UnavailableReason)
	s.Require().NotNil(result.Results[1].RestrictedKeyword)
	s.Equal("equifax", *result.Results[1].RestrictedKeyword)
}

// TestPreWarmedUpDomains verifies the request body reaches the API and the
// results decode, including the provider-annotated domains.
func (s *DFYTestSuite) TestPreWarmedUpDomains() {
	s.Router.Post(preWarmedUpPath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)

		var got dfy.PreWarmedUpDomainsRequest
		s.NoError(json.Unmarshal(body, &got))
		s.Equal([]string{"com", "org"}, got.Extensions)
		s.Equal("acme", got.Search)

		_, _ = w.Write([]byte(`{"domains":["acme.com"],"domains_with_type":[{"domain":"acme.com","account_type":2}]}`))
	})

	result, err := s.svc().PreWarmedUpDomains(context.Background(), dfy.PreWarmedUpDomainsRequest{
		Extensions: []string{"com", "org"},
		Search:     "acme",
	})

	s.Require().NoError(err)
	s.Equal([]string{"acme.com"}, result.Domains)
	s.Require().Len(result.DomainsWithType, 1)
	s.Equal("acme.com", result.DomainsWithType[0].Domain)
	s.Equal(dfy.EmailProviderAirMail, result.DomainsWithType[0].AccountType)
}

// TestCancelAccounts verifies the request body reaches the API and the cancelled
// accounts decode.
func (s *DFYTestSuite) TestCancelAccounts() {
	s.Router.Post(cancelPath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)

		var got dfy.CancelAccountsRequest
		s.NoError(json.Unmarshal(body, &got))
		s.Equal([]string{"john@example.com"}, got.Accounts)

		_, _ = w.Write([]byte(`{"items":[` + accountBare + `]}`))
	})

	result, err := s.svc().CancelAccounts(context.Background(), dfy.CancelAccountsRequest{
		Accounts: []string{"john@example.com"},
	})

	s.Require().NoError(err)
	s.Require().Len(result.Items, 1)
	s.Equal("acc-2", result.Items[0].ID)
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *DFYTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: ordersPath, Status: http.StatusBadRequest,
			Call: func() error { _, err := svc.Create(ctx, dfy.CreateRequest{}); return err },
		},
		{
			Name: "list", Method: http.MethodGet, Path: ordersPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
		{
			Name: "list accounts", Method: http.MethodGet, Path: accountsPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.ListAccounts(ctx, false); return err },
		},
		{
			Name: "generate similar", Method: http.MethodPost, Path: similarPath, Status: http.StatusBadRequest,
			Call: func() error {
				_, err := svc.GenerateSimilarDomains(ctx, dfy.SimilarDomainsRequest{})
				return err
			},
		},
		{
			Name: "check domains", Method: http.MethodPost, Path: checkPath, Status: http.StatusBadRequest,
			Call: func() error { _, err := svc.CheckDomains(ctx, dfy.CheckDomainsRequest{}); return err },
		},
		{
			Name: "pre-warmed-up", Method: http.MethodPost, Path: preWarmedUpPath, Status: http.StatusUnauthorized,
			Call: func() error {
				_, err := svc.PreWarmedUpDomains(ctx, dfy.PreWarmedUpDomainsRequest{})
				return err
			},
		},
		{
			Name: "cancel accounts", Method: http.MethodPost, Path: cancelPath, Status: http.StatusBadRequest,
			Call: func() error { _, err := svc.CancelAccounts(ctx, dfy.CancelAccountsRequest{}); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor on both the order
// and the ordered account parses a valid timestamp and rejects an invalid one.
func (s *DFYTestSuite) TestParsedTimestampCreated() {
	order, err := (&dfy.Order{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, order.Year())

	_, err = (&dfy.Order{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)

	account, err := (&dfy.OrderedAccount{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, account.Year())

	_, err = (&dfy.OrderedAccount{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a DFY Email Account Order service pointed at the suite's mock client.
func (s *DFYTestSuite) svc() *dfy.Service {
	return dfy.New(s.Client)
}
