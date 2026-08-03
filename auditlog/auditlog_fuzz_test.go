package auditlog_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/auditlog"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzAuditLogResponseDecoding feeds arbitrary bytes back as audit-log
// responses, asserting the client never panics and never hands back a partly
// decoded value.
//
// The records are read-only, so there is no request-body serialization to fuzz.
func FuzzAuditLogResponseDecoding(f *testing.F) {
	f.Add(logFixture)
	f.Add(logFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"items":[{"id":"l1"}],"next_starting_after":"cursor-2"}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"activity_type":"not a number"}`)
	f.Add(`{"affected_count":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := auditlog.New(instantlytest.FuzzClient(http.StatusOK, body))

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}
	})
}
