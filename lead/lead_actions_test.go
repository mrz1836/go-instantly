package lead_test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/lead"
)

// TestBulkDelete verifies the delete carries a request body and the count is
// unwrapped for the caller.
func (s *LeadTestSuite) TestBulkDelete() {
	s.Router.Delete(collectionPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodDelete, req.Method)

		var received lead.BulkDeleteRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{"l1", "l2"}, received.IDs)
		s.Equal("campaign-1", received.CampaignID)

		_, _ = w.Write([]byte(`{"count":2}`))
	})

	count, err := s.svc().BulkDelete(context.Background(), lead.BulkDeleteRequest{
		IDs:        []string{"l1", "l2"},
		CampaignID: "campaign-1",
	})

	s.Require().NoError(err)
	s.Equal(int64(2), count)
}

// TestBulkDeleteFailure verifies a failed bulk delete reports zero and an error.
func (s *LeadTestSuite) TestBulkDeleteFailure() {
	s.Router.Delete(collectionPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusUnauthorized, "Unauthorized", "bad key")
	})

	count, err := s.svc().BulkDelete(context.Background(), lead.BulkDeleteRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusUnauthorized)
	s.Zero(count)
}

// TestBulkAdd verifies the leads array is sent and the import summary decodes.
func (s *LeadTestSuite) TestBulkAdd() {
	s.Router.Post(addPath, func(w http.ResponseWriter, req *http.Request) {
		var received lead.BulkAddRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.JSONEq(`[{"email":"a@b.com"}]`, string(received.Leads))
		s.Equal("campaign-1", received.CampaignID)

		_, _ = w.Write([]byte(
			`{"status":"success","leads_uploaded":1,"total_sent":1,"skipped_count":0,` +
				`"duplicated_leads":0,"duplicate_email_count":0,"invalid_email_count":0,` +
				`"incomplete_count":0,"in_blocklist":0,"blocklist_used":null,"remaining_in_plan":99,` +
				`"created_leads":[{"id":"lead-uuid-1"}]}`,
		))
	})

	got, err := s.svc().BulkAdd(context.Background(), lead.BulkAddRequest{
		Leads:      json.RawMessage(`[{"email":"a@b.com"}]`),
		CampaignID: "campaign-1",
	})

	s.Require().NoError(err)
	s.Equal("success", got.Status)
	s.Equal(int64(1), got.LeadsUploaded)
	s.Require().NotNil(got.RemainingInPlan)
	s.Equal(int64(99), *got.RemainingInPlan)
	s.Nil(got.BlocklistUsed)
	s.JSONEq(`[{"id":"lead-uuid-1"}]`, string(got.CreatedLeads))
}

// TestBulkAddFailure verifies a failed add returns no summary.
func (s *LeadTestSuite) TestBulkAddFailure() {
	s.Router.Post(addPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusPaymentRequired, "Payment Required", "limit")
	})

	got, err := s.svc().BulkAdd(context.Background(), lead.BulkAddRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusPaymentRequired)
	s.Nil(got)
}

// TestBulkAssign verifies the assignees are sent and success is a nil error.
func (s *LeadTestSuite) TestBulkAssign() {
	s.Router.Post(bulkAssignPath, func(w http.ResponseWriter, req *http.Request) {
		var received lead.BulkAssignRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{"user-1", "user-2"}, received.OrganizationUserIDs)

		w.WriteHeader(http.StatusOK)
	})

	err := s.svc().BulkAssign(context.Background(), lead.BulkAssignRequest{
		OrganizationUserIDs: []string{"user-1", "user-2"},
	})

	s.Require().NoError(err)
}

// TestBulkAssignFailure verifies a failed assign surfaces the envelope.
func (s *LeadTestSuite) TestBulkAssignFailure() {
	s.Router.Post(bulkAssignPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusForbidden, "Forbidden", "nope")
	})

	err := s.svc().BulkAssign(context.Background(), lead.BulkAssignRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusForbidden)
}

// TestMerge verifies the merge body is sent and the merged lead decodes.
func (s *LeadTestSuite) TestMerge() {
	s.Router.Post(mergePath, func(w http.ResponseWriter, req *http.Request) {
		var received lead.MergeRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("lead-a", received.LeadID)
		s.Equal("lead-b", received.DestinationLeadID)

		_, _ = w.Write([]byte(leadFixture))
	})

	got, err := s.svc().Merge(context.Background(), lead.MergeRequest{
		LeadID:            "lead-a",
		DestinationLeadID: "lead-b",
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestMergeFailure verifies a failed merge returns no lead.
func (s *LeadTestSuite) TestMergeFailure() {
	s.Router.Post(mergePath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no lead")
	})

	got, err := s.svc().Merge(context.Background(), lead.MergeRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestUpdateInterestStatus verifies the required interest_value is always sent,
// even when nil, and success is a nil error.
func (s *LeadTestSuite) TestUpdateInterestStatus() {
	s.Router.Post(interestPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("lead@example.com", received["lead_email"])
		s.InDelta(1, received["interest_value"], 0)

		w.WriteHeader(http.StatusOK)
	})

	err := s.svc().UpdateInterestStatus(context.Background(), lead.UpdateInterestStatusRequest{
		LeadEmail:     "lead@example.com",
		InterestValue: instantly.Ptr(lead.InterestInterested),
	})

	s.Require().NoError(err)
}

// TestUpdateInterestStatusClears verifies a nil interest value is sent as an
// explicit null rather than omitted, since the field is required.
func (s *LeadTestSuite) TestUpdateInterestStatusClears() {
	s.Router.Post(interestPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Contains(received, "interest_value", "the required field must be present")
		s.Nil(received["interest_value"], "a nil interest value is sent as null")

		w.WriteHeader(http.StatusOK)
	})

	err := s.svc().UpdateInterestStatus(context.Background(), lead.UpdateInterestStatusRequest{
		LeadEmail: "lead@example.com",
	})

	s.Require().NoError(err)
}

// TestUpdateInterestStatusFailure verifies a failed update surfaces the envelope.
func (s *LeadTestSuite) TestUpdateInterestStatusFailure() {
	s.Router.Post(interestPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no lead")
	})

	err := s.svc().UpdateInterestStatus(context.Background(), lead.UpdateInterestStatusRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
}

// TestRemoveFromSubsequence verifies the lead id is sent and the lead decodes.
func (s *LeadTestSuite) TestRemoveFromSubsequence() {
	s.Router.Post(subRemovePath, func(w http.ResponseWriter, req *http.Request) {
		var received lead.SubsequenceRemoveRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testID, received.ID)

		_, _ = w.Write([]byte(leadFixture))
	})

	got, err := s.svc().RemoveFromSubsequence(context.Background(), lead.SubsequenceRemoveRequest{ID: testID})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestRemoveFromSubsequenceFailure verifies a failure returns no lead.
func (s *LeadTestSuite) TestRemoveFromSubsequenceFailure() {
	s.Router.Post(subRemovePath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no lead")
	})

	got, err := s.svc().RemoveFromSubsequence(context.Background(), lead.SubsequenceRemoveRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestMoveToSubsequence verifies the ids are sent and the lead decodes.
func (s *LeadTestSuite) TestMoveToSubsequence() {
	s.Router.Post(subMovePath, func(w http.ResponseWriter, req *http.Request) {
		var received lead.SubsequenceMoveRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testID, received.ID)
		s.Equal("subseq-1", received.SubsequenceID)

		_, _ = w.Write([]byte(leadFixture))
	})

	got, err := s.svc().MoveToSubsequence(context.Background(), lead.SubsequenceMoveRequest{
		ID:            testID,
		SubsequenceID: "subseq-1",
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestMoveToSubsequenceFailure verifies a failure returns no lead.
func (s *LeadTestSuite) TestMoveToSubsequenceFailure() {
	s.Router.Post(subMovePath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no lead")
	})

	got, err := s.svc().MoveToSubsequence(context.Background(), lead.SubsequenceMoveRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestMove verifies the move body is sent and the background job decodes.
func (s *LeadTestSuite) TestMove() {
	s.Router.Post(movePath, func(w http.ResponseWriter, req *http.Request) {
		var received lead.MoveRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("campaign-2", received.ToCampaignID)
		s.Equal([]string{"l1"}, received.IDs)

		_, _ = w.Write([]byte(
			`{"id":"job-1","type":"move_leads","status":"pending","progress":0,` +
				`"entity_type":"campaign","workspace_id":"ws-1",` +
				`"created_at":"2026-08-01T10:00:00.000Z","updated_at":"2026-08-01T10:00:00.000Z",` +
				`"entity_id":null,"user_id":null}`,
		))
	})

	got, err := s.svc().Move(context.Background(), lead.MoveRequest{
		ToCampaignID: "campaign-2",
		IDs:          []string{"l1"},
	})

	s.Require().NoError(err)
	s.Equal("job-1", got.ID)
	s.Equal("pending", got.Status)
	s.Nil(got.EntityID)
}

// TestMoveFailure verifies a failed move returns no job.
func (s *LeadTestSuite) TestMoveFailure() {
	s.Router.Post(movePath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusForbidden, "Forbidden", "nope")
	})

	got, err := s.svc().Move(context.Background(), lead.MoveRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusForbidden)
	s.Nil(got)
}
