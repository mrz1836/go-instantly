package dfy

import (
	"context"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the DFY Email Account Order API.
const basePath = "/api/v2/dfy-email-account-orders"

// accountsPath is the accounts sub-path of the DFY Email Account Order API.
const accountsPath = basePath + "/accounts"

// Service provides access to the Instantly.ai V2 DFY Email Account Order API.
type Service struct {
	client *instantly.Client
}

// New builds a DFY Email Account Order API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// EmailProvider is the mailbox product an order targets.
type EmailProvider int64

// The mailbox products an order can target.
const (
	// EmailProviderGoogle is Google: up to 5 mailboxes per domain, priced per
	// mailbox monthly.
	EmailProviderGoogle EmailProvider = 1

	// EmailProviderAirMail is AirMail: up to 5 mailboxes per domain, priced per
	// mailbox monthly.
	EmailProviderAirMail EmailProvider = 2

	// EmailProviderMicrosoft is Microsoft/Outlook: 50-100 mailboxes per new DFY
	// domain, priced per domain monthly.
	EmailProviderMicrosoft EmailProvider = 3
)

// ForwardingMode is how a forwarding domain is applied to an order.
type ForwardingMode string

// The ways a forwarding domain can be applied.
const (
	// ForwardingModeRedirect redirects visitors, changing the browser URL to the
	// forwarding domain.
	ForwardingModeRedirect ForwardingMode = "redirect"

	// ForwardingModeStealth keeps visitors on the domain while proxying content
	// from the forwarding domain.
	ForwardingModeStealth ForwardingMode = "stealth"
)

// OrderType is the kind of order to place.
type OrderType string

// The kinds of order that can be placed.
const (
	// OrderTypeDFY buys new Done-For-You accounts.
	OrderTypeDFY OrderType = "dfy"

	// OrderTypePreWarmedUp buys new pre-warmed-up accounts.
	OrderTypePreWarmedUp OrderType = "pre_warmed_up"

	// OrderTypeExtraAccounts adds extra accounts to already-ordered domains.
	OrderTypeExtraAccounts OrderType = "extra_accounts"
)

// Order is a single DFY email account order returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers, so an absent value stays
// distinguishable from a zero value: a nil ForwardingDomain means the order has
// no forwarding configured, which is not the same as an empty domain.
type Order struct {
	// WorkspaceID identifies the workspace the order belongs to.
	WorkspaceID string `json:"workspace_id"`

	// Domain is the domain of the ordered email accounts.
	Domain string `json:"domain"`

	// TimestampCreated is when the order was created.
	TimestampCreated string `json:"timestamp_created"`

	// ForwardingDomain is the domain emails are forwarded to, if any.
	ForwardingDomain *string `json:"forwarding_domain,omitempty"`

	// ForwardingMode is how the forwarding domain is applied, if any.
	ForwardingMode *ForwardingMode `json:"forwarding_mode,omitempty"`

	// IsPreWarmedUp reports whether the order is for pre-warmed-up accounts.
	IsPreWarmedUp *bool `json:"is_pre_warmed_up,omitempty"`

	// TimestampCancelled is when the order was cancelled, if it was.
	TimestampCancelled *string `json:"timestamp_cancelled,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded order re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (o *Order) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, o.TimestampCreated)
}

// OrderedAccount is a single email account produced by an order.
//
// Password is populated only when a list is requested with passwords, so it is a
// pointer: a nil Password means the API did not return one.
type OrderedAccount struct {
	// ID is the unique identifier of the email account order.
	ID string `json:"id"`

	// Domain is the domain of the email account.
	Domain string `json:"domain"`

	// Email is the address of the account.
	Email string `json:"email"`

	// EmailProvider is the mailbox product the account uses.
	EmailProvider EmailProvider `json:"email_provider"`

	// FirstName is the first name of the account owner.
	FirstName string `json:"first_name"`

	// LastName is the last name of the account owner.
	LastName string `json:"last_name"`

	// IsPreWarmedUp reports whether the account is pre-warmed up.
	IsPreWarmedUp bool `json:"is_pre_warmed_up"`

	// TimestampCancelled is when the account was cancelled.
	TimestampCancelled string `json:"timestamp_cancelled"`

	// TimestampCreated is when the account was created.
	TimestampCreated string `json:"timestamp_created"`

	// Password is the account password, returned only when a list is requested
	// with passwords. It can be empty when the accounts are not ready yet.
	Password *string `json:"password,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded account re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (a *OrderedAccount) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, a.TimestampCreated)
}

// ListResponse is a single page of DFY email account orders.
//
// It aliases instantly.Page[Order], the cursor-paginated envelope every resource
// shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Order]

// AccountsResponse is a single page of ordered email accounts.
//
// It aliases instantly.Page[OrderedAccount] so the generic pagination helpers
// accept ListAccounts directly.
type AccountsResponse = instantly.Page[OrderedAccount]

// List returns a single page of DFY email account orders filtered by the
// supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// ListAccounts returns a single page of the email accounts produced by orders.
//
// withPasswords requests each account's password in the response; it is a
// positional argument rather than an option because it materially changes what
// the endpoint returns. Pagination is cursor based, driven by the same options
// as List.
func (s *Service) ListAccounts(
	ctx context.Context, withPasswords bool, opts ...ListOption,
) (*AccountsResponse, error) {
	q := instantly.ApplyOptions(opts...).SetBool("with_passwords", withPasswords)

	return instantly.GetResult[AccountsResponse](ctx, s.client, q.Path(accountsPath))
}
