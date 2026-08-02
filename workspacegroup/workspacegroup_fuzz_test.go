package workspacegroup_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/workspacegroup"
)

// FuzzWorkspaceGroupSerialization round trips arbitrary field values through the
// create body, asserting the encoding never panics and never drifts.
func FuzzWorkspaceGroupSerialization(f *testing.F) {
	f.Add("ws-sub-1")
	f.Add("")
	f.Add("Ünïcödé\r\n\x00")

	f.Fuzz(func(t *testing.T, subWorkspaceID string) {
		lossless := utf8.ValidString(subWorkspaceID)

		instantlytest.RequireStableRoundTrip(t, workspacegroup.CreateRequest{
			SubWorkspaceID: subWorkspaceID,
		}, lossless)
	})
}

// FuzzWorkspaceGroupResponseDecoding feeds arbitrary bytes back as
// workspace-group-member responses, asserting the client never panics and never
// hands back a partly decoded value.
func FuzzWorkspaceGroupResponseDecoding(f *testing.F) {
	f.Add(memberFixture)
	f.Add(memberFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"has_admin_workspace":true,"workspace_name":"X"}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"status":123}`)
	f.Add(`{"has_admin_workspace":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := workspacegroup.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, memberID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated member")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		admin, err := svc.Admin(ctx)
		if err != nil {
			require.Nil(t, admin, "a decode failure must never hand back a partly populated admin")
		} else {
			require.NotNil(t, admin)
		}

		require.NotPanics(t, func() {
			_, _ = svc.Create(ctx, workspacegroup.CreateRequest{})
			_, _ = svc.Delete(ctx, memberID)
		})
	})
}
