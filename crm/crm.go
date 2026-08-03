package crm

import (
	"context"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the CRM Actions API.
const basePath = "/api/v2/crm-actions"

// phoneNumbersPath is the phone-numbers sub-path of the CRM Actions API.
const phoneNumbersPath = basePath + "/phone-numbers"

// Service provides access to the Instantly.ai V2 CRM Actions API.
type Service struct {
	client *instantly.Client
}

// New builds a CRM Actions API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// PhoneNumber is a single phone number an organization owns.
//
// Price is a pointer because the API declares it nullable and omits it entirely
// from a delete response: a nil Price means the API reported no price, which is
// not the same as a price of zero.
type PhoneNumber struct {
	// ID is the unique identifier of the phone number record.
	ID string `json:"id"`

	// PhoneNumber is the number in international E.164 format.
	PhoneNumber string `json:"phone_number"`

	// Country is the country code the number is registered in.
	Country string `json:"country"`

	// Locality is the city or region associated with the number.
	Locality string `json:"locality"`

	// OrganizationID identifies the organization that owns the number.
	OrganizationID string `json:"organization_id"`

	// Price is the monthly price for the number, in USD.
	Price *float64 `json:"price,omitempty"`

	// SubscriptionID is the billing subscription linked to the number.
	SubscriptionID string `json:"subscription_id,omitempty"`

	// TwilioSID is the Twilio resource SID for the number.
	TwilioSID string `json:"twilio_sid,omitempty"`

	// RenewalDate is the next renewal date for the number's subscription.
	RenewalDate string `json:"renewal_date,omitempty"`

	// TimestampCreated is when the phone number record was created.
	TimestampCreated string `json:"timestamp_created"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded record re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (p *PhoneNumber) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, p.TimestampCreated)
}

// ListPhoneNumbers returns every phone number the organization owns.
//
// The endpoint answers with a bare JSON array rather than a paginated envelope,
// so the whole list is returned at once.
func (s *Service) ListPhoneNumbers(ctx context.Context) ([]PhoneNumber, error) {
	var out []PhoneNumber
	if err := s.client.Get(ctx, phoneNumbersPath, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// DeletePhoneNumber deletes a phone number and returns the number that was
// deleted.
//
// The delete response omits the price and renewal date, so those fields are
// their zero values on the returned record.
func (s *Service) DeletePhoneNumber(ctx context.Context, id string) (*PhoneNumber, error) {
	return instantly.DeleteResult[PhoneNumber](ctx, s.client, instantly.JoinPath(phoneNumbersPath, id))
}
