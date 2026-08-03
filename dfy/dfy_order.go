package dfy

import (
	"context"
	"encoding/json"

	"github.com/mrz1836/go-instantly"
)

// CreateRequest is the body of a place-order request.
type CreateRequest struct {
	// Items are the domains and accounts to order.
	Items []OrderItem `json:"items"`

	// OrderType is the kind of order to place. Required.
	OrderType OrderType `json:"order_type"`

	// Simulation runs the order as a price quote without placing it or charging
	// when set to true.
	Simulation *bool `json:"simulation,omitempty"`
}

// OrderItem is a single domain (and its accounts) to order.
type OrderItem struct {
	// Domain is the domain to use for the email accounts. Required.
	Domain string `json:"domain"`

	// EmailProvider is the mailbox product to order, defaulting to Google when
	// omitted.
	EmailProvider *EmailProvider `json:"email_provider,omitempty"`

	// ForwardingDomain is an optional domain to forward emails to, which must
	// differ from Domain.
	ForwardingDomain string `json:"forwarding_domain,omitempty"`

	// Accounts are the email accounts to create on the domain. Ignored for
	// pre-warmed-up domains, whose accounts already exist.
	Accounts []AccountSpec `json:"accounts,omitempty"`
}

// AccountSpec is a single email account to create on a domain.
type AccountSpec struct {
	// EmailAddressPrefix is the part of the address before the @. Required.
	EmailAddressPrefix string `json:"email_address_prefix"`

	// FirstName is the first name of the account owner. Required.
	FirstName string `json:"first_name"`

	// LastName is the last name of the account owner. Required.
	LastName string `json:"last_name"`
}

// InvalidAccount is one account an order rejected, with the reason it was
// invalid.
type InvalidAccount struct {
	// Domain is the domain of the rejected account.
	Domain string `json:"domain"`

	// Email is the address of the rejected account.
	Email string `json:"email"`

	// FirstName is the first name supplied for the rejected account.
	FirstName string `json:"first_name"`

	// LastName is the last name supplied for the rejected account.
	LastName string `json:"last_name"`

	// Reason is why the account was invalid.
	Reason string `json:"reason"`
}

// ResultAccount is a single account echoed back on an order item.
type ResultAccount struct {
	// EmailAddressPrefix is the part of the address before the @.
	EmailAddressPrefix string `json:"email_address_prefix"`

	// FirstName is the first name of the account owner.
	FirstName string `json:"first_name"`

	// LastName is the last name of the account owner.
	LastName string `json:"last_name"`
}

// OrderResultItem is one domain of an order, with its pricing.
type OrderResultItem struct {
	// Domain is the domain the item is for.
	Domain string `json:"domain"`

	// Accounts are the accounts ordered for the domain.
	Accounts []ResultAccount `json:"accounts"`

	// EmailProvider is the mailbox product ordered for the domain.
	EmailProvider EmailProvider `json:"email_provider"`

	// ForwardingDomain is the forwarding domain for the item, if any.
	ForwardingDomain string `json:"forwarding_domain,omitempty"`

	// DomainPrice is the price for the domain.
	DomainPrice float64 `json:"domain_price"`

	// DomainMonthlyPrice is the monthly domain-bundle price, populated only for
	// domain-level billing providers such as Microsoft/Outlook.
	DomainMonthlyPrice *float64 `json:"domain_monthly_price,omitempty"`

	// AccountsPrice is the total price for the accounts in the item.
	AccountsPrice float64 `json:"accounts_price"`

	// TotalPrice is the total price for the item.
	TotalPrice float64 `json:"total_price"`

	// TotalDiscount is the total discount for the item.
	TotalDiscount float64 `json:"total_discount"`
}

// OrderResult is the outcome of a place-order request, whether placed,
// simulated, or rejected.
//
// OrderError is the failure discriminator: it is a normal response field (spelt
// order_error, not error) that is empty on success and names the reason a
// rejected order failed, pointing at the matching detail field (for example
// UnavailableDomains). Prices the API declares nullable are pointers so an
// absent price stays distinguishable from a price of zero.
type OrderResult struct {
	// OrderPlaced reports whether the order was placed.
	OrderPlaced bool `json:"order_placed"`

	// OrderIsValid reports whether the order is valid and could be placed.
	OrderIsValid bool `json:"order_is_valid"`

	// Simulation reports whether the request ran as a simulation.
	Simulation bool `json:"simulation"`

	// OrderError names the reason a rejected order failed, empty on success.
	OrderError string `json:"order_error,omitempty"`

	// UnavailableDomains are domains not available for order.
	UnavailableDomains []string `json:"unavailable_domains"`

	// BlacklistDomains are domains that are blacklisted.
	BlacklistDomains []string `json:"blacklist_domains"`

	// BlacklistKeywords are the restricted keywords matched in BlacklistDomains.
	BlacklistKeywords []string `json:"blacklist_keywords,omitempty"`

	// InvalidDomains are domains that are invalid.
	InvalidDomains []string `json:"invalid_domains"`

	// InvalidForwardingDomains are forwarding domains that are invalid.
	InvalidForwardingDomains []string `json:"invalid_forwarding_domains"`

	// InvalidAccounts are accounts that were rejected, with their reasons.
	InvalidAccounts []InvalidAccount `json:"invalid_accounts"`

	// MissingDomainOrders are domains missing a prior order, for extra-account
	// orders.
	MissingDomainOrders []string `json:"missing_domain_orders"`

	// ProviderMismatchDomains are domains whose requested provider does not match
	// their existing accounts.
	ProviderMismatchDomains []string `json:"provider_mismatch_domains"`

	// UnsupportedProviderDomains are domains whose existing provider does not
	// support extra-account orders through this endpoint.
	UnsupportedProviderDomains []string `json:"unsupported_provider_domains"`

	// UnavailableEmailProviders are the requested providers not available to
	// order right now.
	UnavailableEmailProviders []EmailProvider `json:"unavailable_email_providers"`

	// DomainsWithoutAccounts are domains whose accounts field was empty.
	DomainsWithoutAccounts []string `json:"domains_without_accounts"`

	// FreeDomains are domains that are free, for example during a promotion.
	FreeDomains []string `json:"free_domains"`

	// NumberOfDomainsOrdered is how many domains were ordered.
	NumberOfDomainsOrdered float64 `json:"number_of_domains_ordered"`

	// NumberOfAccountsOrdered is how many accounts were ordered.
	NumberOfAccountsOrdered float64 `json:"number_of_accounts_ordered"`

	// PricePerAccountPerMonth is the legacy monthly per-mailbox price, null for
	// mixed provider orders.
	PricePerAccountPerMonth *float64 `json:"price_per_account_per_month"`

	// PricePerAccountPerMonthByAccountType carries provider-specific monthly
	// mailbox prices keyed by account type, which the API models as a free-form
	// object, so it is preserved verbatim.
	PricePerAccountPerMonthByAccountType json.RawMessage `json:"price_per_account_per_month_by_account_type,omitempty"`

	// PricePerDomainPerMonth is the monthly per-domain price, populated only for
	// domain-level billing providers.
	PricePerDomainPerMonth *float64 `json:"price_per_domain_per_month,omitempty"`

	// PricePerDomainPerYear is the yearly per-domain price.
	PricePerDomainPerYear float64 `json:"price_per_domain_per_year"`

	// TotalDomainsPricePerYear is the total yearly domain price.
	TotalDomainsPricePerYear float64 `json:"total_domains_price_per_year"`

	// TotalAccountsPricePerMonth is the total monthly account price.
	TotalAccountsPricePerMonth float64 `json:"total_accounts_price_per_month"`

	// TotalPricePerMonth is the total monthly price.
	TotalPricePerMonth float64 `json:"total_price_per_month"`

	// TotalPricePerYear is the total yearly price.
	TotalPricePerYear float64 `json:"total_price_per_year"`

	// TotalPrice is the total price for the order.
	TotalPrice float64 `json:"total_price"`

	// TotalDiscount is the total discount applied to the order.
	TotalDiscount float64 `json:"total_discount"`

	// OrderItems are the ordered items with their pricing.
	OrderItems []OrderResultItem `json:"order_items"`

	// PaymentMethodLast4Digits is the last 4 digits of the payment method.
	PaymentMethodLast4Digits string `json:"payment_method_last_4_digits"`

	// PaymentMethodBrand is the brand of the payment method.
	PaymentMethodBrand string `json:"payment_method_brand"`

	// PaymentMethodNameOnCard is the name on the payment card.
	PaymentMethodNameOnCard string `json:"payment_method_name_on_card"`

	// CheckoutRequired reports whether payment through a hosted checkout is needed
	// before the order can be placed.
	CheckoutRequired bool `json:"checkout_required,omitempty"`

	// CheckoutURL is the hosted checkout URL, for orders that require a payment
	// method.
	CheckoutURL string `json:"checkout_url,omitempty"`

	// CartOrderID is the identifier used to fulfill the order after checkout.
	CartOrderID string `json:"cart_order_id,omitempty"`

	// PaymentFailureReason is a provider-independent payment failure reason,
	// present only when a payment attempt failed.
	PaymentFailureReason string `json:"payment_failure_reason,omitempty"`
}

// Create places a DFY email account order, or simulates one when the request
// sets Simulation.
//
// The endpoint reports a rejected order inside an otherwise successful response
// rather than as an HTTP error: inspect OrderPlaced, OrderIsValid, and
// OrderError on the result. A nil error means the request was processed, not
// that an order was placed.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*OrderResult, error) {
	return instantly.PostResult[OrderResult](ctx, s.client, basePath, req)
}
