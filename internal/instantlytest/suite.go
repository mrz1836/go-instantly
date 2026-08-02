package instantlytest

import (
	"net/http/httptest"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
)

const (
	// APIKey is the V2 API key every suite request authenticates with.
	APIKey = "test-api-key"

	// AuthHeader is the exact Authorization header value the client must send.
	AuthHeader = "Bearer " + APIKey

	// MediaTypeJSON is the media type used for both Accept and Content-Type.
	MediaTypeJSON = "application/json"

	// SuccessBody is a minimal successful JSON response body.
	SuccessBody = `{"status":"success"}`
)

// Suite is the reusable base for every resource test suite.
//
// It boots an in-process mock API server and points a client at it, so no test
// in this repository ever reaches api.instantly.ai. Embed it, then register
// routes on Router and call the resource service under test with Client:
//
//	type EmailTestSuite struct {
//		instantlytest.Suite
//	}
//
//	func (s *EmailTestSuite) TestGet() {
//		svc := email.New(s.Client)
//		s.Router.Get("/api/v2/emails/:id", handler)
//		// ...
//	}
type Suite struct {
	suite.Suite

	// Router is the mock router requests are matched against.
	Router *Router

	// Server is the in-process HTTP server the Router is mounted on.
	Server *httptest.Server

	// Client is an Instantly client pointed at Server.
	Client *instantly.Client
}

// SetupSuite starts the mock API server and points a client at it.
func (s *Suite) SetupSuite() {
	s.Router = NewRouter()
	s.Server = httptest.NewServer(s.Router)
	s.Client = instantly.NewClient(APIKey, instantly.WithBaseURL(s.Server.URL))
}

// TearDownSuite shuts the mock API server down.
func (s *Suite) TearDownSuite() {
	if s.Server != nil {
		s.Server.Close()
	}
}
