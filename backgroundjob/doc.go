// Package backgroundjob provides typed access to the Instantly.ai V2 Background
// Job API.
//
// It wraps the /api/v2/background-jobs endpoints, which list the long-running
// jobs a workspace has queued (lead moves, imports, exports, warmup updates, and
// more) and read a single job by id, optionally selecting which fields of its
// data payload to return.
//
//	svc := backgroundjob.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, backgroundjob.WithStatus("pending,in-progress"))
//	job, err := svc.Get(ctx, "job-1", backgroundjob.WithDataFields("success_count,failed_count"))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package backgroundjob
