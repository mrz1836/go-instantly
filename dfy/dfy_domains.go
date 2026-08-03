package dfy

import (
	"context"

	"github.com/mrz1836/go-instantly"
)

// SimilarDomainsRequest is the body of a generate-similar-domains request.
type SimilarDomainsRequest struct {
	// Domain is the domain to base the suggestions on. Required.
	Domain string `json:"domain"`

	// TLDs are the extensions to generate similar domains for, defaulting to com
	// and org when omitted.
	TLDs []string `json:"tlds,omitempty"`
}

// SimilarDomainsResult is the outcome of a generate-similar-domains request.
type SimilarDomainsResult struct {
	// Domains are the similar, available domains that were suggested.
	Domains []string `json:"domains"`
}

// CheckDomainsRequest is the body of a check-domains request.
type CheckDomainsRequest struct {
	// Domains are the domains to check availability for. Required.
	Domains []string `json:"domains"`
}

// DomainAvailability is the availability of a single checked domain.
type DomainAvailability struct {
	// Domain is the domain that was checked.
	Domain string `json:"domain"`

	// Available reports whether the domain is available to order.
	Available bool `json:"available"`

	// UnavailableReason is why the domain is unavailable, when it is restricted.
	UnavailableReason *string `json:"unavailable_reason,omitempty"`

	// RestrictedKeyword is the restricted keyword matched when UnavailableReason
	// is restricted.
	RestrictedKeyword *string `json:"restricted_keyword,omitempty"`
}

// CheckDomainsResult is the outcome of a check-domains request.
type CheckDomainsResult struct {
	// Results are the checked domains with their availability.
	Results []DomainAvailability `json:"results"`
}

// PreWarmedUpDomainsRequest is the body of a pre-warmed-up-domains request. Both
// fields are optional.
type PreWarmedUpDomainsRequest struct {
	// Extensions filters results to the given domain extensions.
	Extensions []string `json:"extensions,omitempty"`

	// Search filters results to domains matching the term.
	Search string `json:"search,omitempty"`
}

// DomainWithType is a pre-warmed-up domain annotated with its provider.
type DomainWithType struct {
	// Domain is the pre-warmed-up domain.
	Domain string `json:"domain"`

	// AccountType is the mailbox product behind the domain.
	AccountType EmailProvider `json:"account_type"`
}

// PreWarmedUpDomainsResult is the outcome of a pre-warmed-up-domains request.
type PreWarmedUpDomainsResult struct {
	// Domains are the pre-warmed-up domains available to order.
	Domains []string `json:"domains"`

	// DomainsWithType are the domains annotated with their underlying provider.
	DomainsWithType []DomainWithType `json:"domains_with_type"`
}

// CancelAccountsRequest is the body of a cancel-accounts request.
type CancelAccountsRequest struct {
	// Accounts are the email addresses of the accounts to cancel. Required.
	Accounts []string `json:"accounts"`
}

// CancelAccountsResult is the outcome of a cancel-accounts request.
type CancelAccountsResult struct {
	// Items are the accounts that were cancelled.
	Items []OrderedAccount `json:"items"`
}

// GenerateSimilarDomains suggests available domains similar to a given one.
func (s *Service) GenerateSimilarDomains(
	ctx context.Context, req SimilarDomainsRequest,
) (*SimilarDomainsResult, error) {
	return instantly.PostResult[SimilarDomainsResult](ctx, s.client, basePath+"/domains/similar", req)
}

// CheckDomains checks whether the given domains are available to order.
func (s *Service) CheckDomains(ctx context.Context, req CheckDomainsRequest) (*CheckDomainsResult, error) {
	return instantly.PostResult[CheckDomainsResult](ctx, s.client, basePath+"/domains/check", req)
}

// PreWarmedUpDomains lists the pre-warmed-up domains available to order,
// filtered by the request.
func (s *Service) PreWarmedUpDomains(
	ctx context.Context, req PreWarmedUpDomainsRequest,
) (*PreWarmedUpDomainsResult, error) {
	return instantly.PostResult[PreWarmedUpDomainsResult](ctx, s.client, basePath+"/domains/pre-warmed-up-list", req)
}

// CancelAccounts cancels the given DFY email accounts and returns the accounts
// that were cancelled.
func (s *Service) CancelAccounts(ctx context.Context, req CancelAccountsRequest) (*CancelAccountsResult, error) {
	return instantly.PostResult[CancelAccountsResult](ctx, s.client, accountsPath+"/cancel", req)
}
