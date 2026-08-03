package crm_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/crm"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// listPath is the router pattern for the crm-actions phone-numbers endpoint.
const listPath = "/api/v2/crm-actions/phone-numbers"

// idPath is the router pattern for the single-phone-number endpoint.
const idPath = "/api/v2/crm-actions/phone-numbers/:id"

// numberFixture is a spec-shaped phone number with every documented field
// populated, including the nullable price.
const numberFixture = `{
	"id": "num-1",
	"phone_number": "+15551234567",
	"country": "US",
	"locality": "San Francisco",
	"organization_id": "org-1",
	"price": 1,
	"subscription_id": "sub-1",
	"twilio_sid": "PN998",
	"renewal_date": "2026-09-01T00:00:00.000Z",
	"timestamp_created": "2026-08-01T10:00:00.000Z"
}`

// numberFixtureNoPrice is a spec-shaped phone number with the nullable price
// explicitly null, as the delete response reports it.
const numberFixtureNoPrice = `{
	"id": "num-2",
	"phone_number": "+15559876543",
	"country": "US",
	"locality": "San Francisco",
	"organization_id": "org-1",
	"price": null,
	"subscription_id": "sub-2",
	"twilio_sid": "PN999",
	"timestamp_created": "2026-08-01T11:00:00.000Z"
}`

// CRMTestSuite exercises the CRM Actions API service against the mock router.
type CRMTestSuite struct {
	instantlytest.Suite
}

// TestCRMSuite runs the CRM Actions API suite.
func TestCRMSuite(t *testing.T) {
	suite.Run(t, new(CRMTestSuite))
}

// TestListPhoneNumbers verifies the bare-array response decodes, with a number
// that carries a price and one that does not.
func (s *CRMTestSuite) TestListPhoneNumbers() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(listPath, req.URL.Path)
		_, _ = w.Write([]byte("[" + numberFixture + "," + numberFixtureNoPrice + "]"))
	})

	numbers, err := s.svc().ListPhoneNumbers(context.Background())

	s.Require().NoError(err)
	s.Require().Len(numbers, 2)

	s.Equal("num-1", numbers[0].ID)
	s.Equal("+15551234567", numbers[0].PhoneNumber)
	s.Require().NotNil(numbers[0].Price)
	s.InEpsilon(1.0, *numbers[0].Price, 1e-9)

	s.Equal("num-2", numbers[1].ID)
	s.Nil(numbers[1].Price, "a null price must decode to nil, not zero")
}

// TestDeletePhoneNumber verifies a delete returns the removed number and escapes
// the id.
func (s *CRMTestSuite) TestDeletePhoneNumber() {
	s.Router.Delete(idPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("num-2", instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(numberFixtureNoPrice))
	})

	number, err := s.svc().DeletePhoneNumber(context.Background(), "num-2")

	s.Require().NoError(err)
	s.Equal("num-2", number.ID)
	s.Nil(number.Price, "the delete response omits price, so it stays nil")
	s.Empty(number.RenewalDate, "the delete response omits renewal_date, so it stays empty")
}

// TestDeletePathEscape verifies an id carrying reserved characters is escaped
// into the request path rather than altering it.
func (s *CRMTestSuite) TestDeletePathEscape() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, numberFixtureNoPrice), nil
			},
		)},
	))

	_, err := crm.New(client).DeletePhoneNumber(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/crm-actions/phone-numbers/..%2Fadmin%3Fx=1", requestURI)
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *CRMTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.ListPhoneNumbers(ctx); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.DeletePhoneNumber(ctx, "num-1"); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *CRMTestSuite) TestParsedTimestampCreated() {
	got, err := (&crm.PhoneNumber{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&crm.PhoneNumber{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a CRM Actions service pointed at the suite's mock client.
func (s *CRMTestSuite) svc() *crm.Service {
	return crm.New(s.Client)
}
