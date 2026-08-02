package lead_test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mrz1836/go-instantly"
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
