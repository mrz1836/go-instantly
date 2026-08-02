package workspace_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/workspace"
)

// FuzzWorkspaceSerialization round trips arbitrary field values through the
// request bodies, asserting the encoding never panics and never drifts.
func FuzzWorkspaceSerialization(f *testing.F) {
	f.Add("My Workspace", "https://example.com/logo.png", "new@example.com", "token-123")
	f.Add("", "", "", "")
	f.Add("Ünïcödé", "line\r\n", "\x00owner", "agency.example.com")

	f.Fuzz(func(t *testing.T, name, logo, email, misc string) {
		lossless := utf8.ValidString(name) && utf8.ValidString(logo) &&
			utf8.ValidString(email) && utf8.ValidString(misc)

		instantlytest.RequireStableRoundTrip(t, workspace.UpdateRequest{
			Name:       name,
			OrgLogoURL: &logo,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, workspace.SetDomainRequest{Domain: misc}, lossless)

		instantlytest.RequireStableRoundTrip(t, workspace.ChangeOwnerRequest{
			Email: email,
			Sec:   misc,
		}, lossless)
	})
}

// FuzzWorkspaceResponseDecoding feeds arbitrary bytes back as workspace
// responses, asserting the client never panics and never hands back a partly
// decoded value.
func FuzzWorkspaceResponseDecoding(f *testing.F) {
	f.Add(workspaceFixture)
	f.Add(workspaceFixtureNulls)
	f.Add(`{"name":"x","verified":true,"verification":[]}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"default_opportunity_value":"not a number"}`)
	f.Add(`{"add_unsub_to_block":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := workspace.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated workspace")
		} else {
			require.NotNil(t, got)
		}

		info, err := svc.DomainInfo(ctx)
		if err != nil {
			require.Nil(t, info, "a decode failure must never hand back a partly populated domain info")
		} else {
			require.NotNil(t, info)
		}

		require.NotPanics(t, func() {
			_, _ = svc.Update(ctx, workspace.UpdateRequest{})
			_, _ = svc.ScheduleRemoval(ctx)
			_, _ = svc.ChangeOwner(ctx, workspace.ChangeOwnerRequest{})
		})
	})
}
