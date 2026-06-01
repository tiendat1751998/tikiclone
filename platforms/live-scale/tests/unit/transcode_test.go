package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/transcoding"
)

func TestTranscodeCreateJob(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	job, err := svc.CreateJob(context.Background(), "stream-001", "rtmp://origin/live/stream-001", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != transcoding.JobPending {
		t.Errorf("expected pending, got %v", job.Status)
	}
	if job.StreamID != "stream-001" {
		t.Errorf("expected stream-001, got %s", job.StreamID)
	}
}

func TestTranscodeCreateJobWithProfiles(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	job, err := svc.CreateJob(context.Background(), "stream-002", "rtmp://origin/live/stream-002",
		[]transcoding.VideoProfile{transcoding.Profile480p, transcoding.Profile720p, transcoding.Profile1080p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(job.Profiles) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(job.Profiles))
	}
}

func TestTranscodeCreateJobInvalid(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	_, err := svc.CreateJob(context.Background(), "", "", nil)
	if err != transcoding.ErrInvalidJobData {
		t.Errorf("expected ErrInvalidJobData, got %v", err)
	}
}

func TestTranscodeCreateJobUnsupportedProfile(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	_, err := svc.CreateJob(context.Background(), "stream-003", "rtmp://origin/live/stream-003",
		[]transcoding.VideoProfile{"4k"})
	if err == nil {
		t.Fatal("expected error for unsupported profile")
	}
}

func TestTranscodeGetJobStatus(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	job, _ := svc.CreateJob(context.Background(), "stream-004", "rtmp://origin/live/stream-004", nil)
	got, err := svc.GetJobStatus(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != job.ID {
		t.Errorf("expected job ID %s, got %s", job.ID, got.ID)
	}
}

func TestTranscodeGetJobNotFound(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	_, err := svc.GetJobStatus(context.Background(), "nonexistent")
	if err != transcoding.ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}

func TestTranscodeStartJob(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	job, _ := svc.CreateJob(context.Background(), "stream-005", "rtmp://origin/live/stream-005", nil)
	if err := svc.StartJob(context.Background(), job.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetJobStatus(context.Background(), job.ID)
	if got.Status != transcoding.JobProcessing {
		t.Errorf("expected processing, got %v", got.Status)
	}
}

func TestTranscodeCompleteJob(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	job, _ := svc.CreateJob(context.Background(), "stream-006", "rtmp://origin/live/stream-006", nil)
	svc.StartJob(context.Background(), job.ID)
	outputs := []transcoding.Output{
		{Profile: transcoding.Profile720p, URL: "https://cdn.example.com/stream-006_720p.m3u8", SizeBytes: 1024000, DurationMs: 60000},
	}
	if err := svc.CompleteJob(context.Background(), job.ID, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetJobStatus(context.Background(), job.ID)
	if got.Status != transcoding.JobCompleted {
		t.Errorf("expected completed, got %v", got.Status)
	}
	if len(got.Outputs) != 1 {
		t.Errorf("expected 1 output, got %d", len(got.Outputs))
	}
}

func TestTranscodeCancelJob(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	job, _ := svc.CreateJob(context.Background(), "stream-007", "rtmp://origin/live/stream-007", nil)
	if err := svc.CancelJob(context.Background(), job.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetJobStatus(context.Background(), job.ID)
	if got.Status != transcoding.JobFailed {
		t.Errorf("expected failed (cancelled), got %v", got.Status)
	}
}

func TestTranscodeFailJob(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	job, _ := svc.CreateJob(context.Background(), "stream-008", "rtmp://origin/live/stream-008", nil)
	svc.StartJob(context.Background(), job.ID)
	if err := svc.FailJob(context.Background(), job.ID, "encoding error"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetJobStatus(context.Background(), job.ID)
	if got.Status != transcoding.JobFailed {
		t.Errorf("expected failed, got %v", got.Status)
	}
	if got.Error != "encoding error" {
		t.Errorf("expected 'encoding error', got %s", got.Error)
	}
}

func TestTranscodeListJobs(t *testing.T) {
	svc := transcoding.NewService(transcoding.NewInMemoryRepository())
	svc.CreateJob(context.Background(), "stream-009", "rtmp://origin/live/stream-009", nil)
	svc.CreateJob(context.Background(), "stream-009", "rtmp://origin/live/stream-009-2", nil)
	jobs, err := svc.ListJobs(context.Background(), "stream-009")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestTranscodeSupportedProfiles(t *testing.T) {
	if _, ok := transcoding.SupportedProfiles[transcoding.Profile480p]; !ok {
		t.Error("480p profile not found")
	}
	if _, ok := transcoding.SupportedProfiles[transcoding.Profile720p]; !ok {
		t.Error("720p profile not found")
	}
	if _, ok := transcoding.SupportedProfiles[transcoding.Profile1080p]; !ok {
		t.Error("1080p profile not found")
	}
	p480 := transcoding.SupportedProfiles[transcoding.Profile480p]
	if p480.Width != 854 || p480.Height != 480 || p480.BitrateKbps != 1500 {
		t.Errorf("480p config mismatch: %+v", p480)
	}
}
