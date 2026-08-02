package workspacemember_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/workspacemember"
)

// FuzzWorkspaceMemberSerialization round trips arbitrary field values through the
// create and update bodies, asserting the encoding never panics and never drifts.
func FuzzWorkspaceMemberSerialization(f *testing.F) {
	f.Add("member@example.com", "admin", "JD", "unibox.all")
	f.Add("", "", "", "")
	f.Add("a@b.c", "Ünïcödé", "line\r\n", "\x00perm")

	f.Fuzz(func(t *testing.T, email, role, nickname, permission string) {
		lossless := utf8.ValidString(email) && utf8.ValidString(role) &&
			utf8.ValidString(nickname) && utf8.ValidString(permission)

		instantlytest.RequireStableRoundTrip(t, workspacemember.CreateRequest{
			Email:       email,
			Role:        workspacemember.Role(role),
			Nickname:    &nickname,
			Permissions: []workspacemember.Permission{workspacemember.Permission(permission)},
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, workspacemember.UpdateRequest{
			Role:     workspacemember.Role(role),
			Nickname: instantly.Ptr(nickname),
		}, lossless)
	})
}

// FuzzWorkspaceMemberResponseDecoding feeds arbitrary bytes back as
// workspace-member responses, asserting the client never panics and never hands
// back a partly decoded value.
func FuzzWorkspaceMemberResponseDecoding(f *testing.F) {
	f.Add(memberFixture)
	f.Add(memberFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"accepted":"not a bool"}`)
	f.Add(`{"name":"not an object"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := workspacemember.New(instantlytest.FuzzClient(http.StatusOK, body))

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

		require.NotPanics(t, func() {
			_, _ = svc.Create(ctx, workspacemember.CreateRequest{})
			_, _ = svc.Delete(ctx, memberID)
		})
	})
}
