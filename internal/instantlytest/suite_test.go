package instantlytest_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// harnessSuite embeds the reusable Suite so its SetupSuite, TearDownSuite, and
// RunFailures helpers are exercised directly rather than only through the
// resource suites that consume them.
type harnessSuite struct {
	instantlytest.Suite
}

// TestHarnessSuite runs the harness suite.
func TestHarnessSuite(t *testing.T) {
	suite.Run(t, new(harnessSuite))
}

// TestServerIsWired verifies SetupSuite booted a server and pointed a client at it.
func (s *harnessSuite) TestServerIsWired() {
	s.Require().NotNil(s.Client)
	s.Require().NotNil(s.Server)
	s.Require().NotNil(s.Router)
}

// TestRunFailures drives a failure case end to end through the shared runner,
// covering FailHandler, WriteAPIErrorEnvelope, and AssertAPIError.
func (s *harnessSuite) TestRunFailures() {
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "not found", Method: http.MethodGet, Path: "/api/v2/thing", Status: http.StatusNotFound,
			Call: func() error { return s.Client.Get(context.Background(), "/api/v2/thing", nil) },
		},
	})
}
