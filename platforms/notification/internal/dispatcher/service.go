package dispatcher

import (
	"context"
	"math"
	"time"

	"github.com/tikiclone/tiki/platforms/notification/internal/email"
	"github.com/tikiclone/tiki/platforms/notification/internal/inapp"
	"github.com/tikiclone/tiki/platforms/notification/internal/notifier"
	"github.com/tikiclone/tiki/platforms/notification/internal/preferences"
	"github.com/tikiclone/tiki/platforms/notification/internal/push"
	"github.com/tikiclone/tiki/platforms/notification/internal/sms"
)

type Service interface {
	Dispatch(ctx context.Context, job *DispatchJob) (*DispatchResult, error)
	BatchDispatch(ctx context.Context, jobs []*DispatchJob) ([]*DispatchResult, error)
	RetryFailed(ctx context.Context) ([]*DispatchResult, error)
	CreateJob(ctx context.Context, req *notifier.SendNotificationRequest) (*DispatchJob, error)
}

type service struct {
	repo        Repository
	notifSvc    notifier.Service
	pushSvc     push.Service
	emailSvc    email.Service
	smsSvc      sms.Service
	inappSvc    inapp.Service
	prefSvc     preferences.Service
	retryPolicy RetryPolicy
}

func NewService(
	repo Repository,
	notifSvc notifier.Service,
	pushSvc push.Service,
	emailSvc email.Service,
	smsSvc sms.Service,
	inappSvc inapp.Service,
	prefSvc preferences.Service,
) Service {
	return &service{
		repo:        repo,
		notifSvc:    notifSvc,
		pushSvc:     pushSvc,
		emailSvc:    emailSvc,
		smsSvc:      smsSvc,
		inappSvc:    inappSvc,
		prefSvc:     prefSvc,
		retryPolicy: DefaultRetryPolicy(),
	}
}

func (s *service) CreateJob(ctx context.Context, req *notifier.SendNotificationRequest) (*DispatchJob, error) {
	job := &DispatchJob{
		UserID:     req.UserID,
		Channel:    req.Channel,
		Type:       req.Type,
		Title:      req.Title,
		Body:       req.Body,
		Data:       req.Data,
		Priority:   req.Priority,
		MaxRetries: s.retryPolicy.MaxRetries,
		Status:     "pending",
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	result, err := s.Dispatch(ctx, job)
	if err != nil {
		s.repo.UpdateStatus(ctx, job.ID, "failed", err.Error())
		return job, err
	}

	if result.Success {
		s.repo.UpdateStatus(ctx, job.ID, "completed", "")
	} else {
		s.repo.UpdateStatus(ctx, job.ID, "failed", result.Error)
	}

	return job, nil
}

func (s *service) Dispatch(ctx context.Context, job *DispatchJob) (*DispatchResult, error) {
	ok, err := s.prefSvc.ShouldSend(ctx, job.UserID, string(job.Channel), string(job.Type))
	if err != nil {
		return &DispatchResult{JobID: job.ID, Success: false, Channel: string(job.Channel), Error: err.Error(), Attempt: job.RetryCount + 1}, err
	}
	if !ok {
		return &DispatchResult{JobID: job.ID, Success: false, Channel: string(job.Channel), Error: "blocked by preferences", Attempt: job.RetryCount + 1}, nil
	}

	result := &DispatchResult{
		JobID:   job.ID,
		Channel: string(job.Channel),
		Attempt: job.RetryCount + 1,
	}

	switch job.Channel {
	case notifier.ChannelPush:
		pushReq := &push.PushNotificationRequest{
			UserID: job.UserID,
			Title:  job.Title,
			Body:   job.Body,
		}
		pushResult, err := s.pushSvc.SendPush(ctx, pushReq)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = pushResult.Success
			result.Error = pushResult.Error
		}

	case notifier.ChannelEmail:
		emailReq := &email.SendEmailRequest{
			To:      []string{job.UserID},
			Subject: job.Title,
			HTML:    job.Body,
		}
		_, err := s.emailSvc.SendEmail(ctx, emailReq)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
		}

	case notifier.ChannelSMS:
		smsReq := &sms.SendSMSRequest{
			To:   job.UserID,
			Body: job.Title + ": " + job.Body,
		}
		_, err := s.smsSvc.SendSMS(ctx, smsReq)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
		}

	case notifier.ChannelInApp:
		inappReq := &inapp.SendInAppRequest{
			UserID:   job.UserID,
			Category: inapp.CategorySystem,
			Title:    job.Title,
			Body:     job.Body,
		}
		_, err := s.inappSvc.SendInApp(ctx, inappReq)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
		}

	default:
		result.Success = false
		result.Error = "unknown channel"
	}

	if !result.Success && job.RetryCount < job.MaxRetries {
		s.repo.IncrementRetry(ctx, job.ID)
	}

	return result, nil
}

func (s *service) BatchDispatch(ctx context.Context, jobs []*DispatchJob) ([]*DispatchResult, error) {
	results := make([]*DispatchResult, 0, len(jobs))
	for _, job := range jobs {
		result, err := s.Dispatch(ctx, job)
		if err != nil {
			results = append(results, &DispatchResult{
				JobID:   job.ID,
				Success: false,
				Channel: string(job.Channel),
				Error:   err.Error(),
				Attempt: job.RetryCount + 1,
			})
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *service) RetryFailed(ctx context.Context) ([]*DispatchResult, error) {
	failed, err := s.repo.GetFailed(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*DispatchResult, 0, len(failed))
	for _, job := range failed {
		if job.RetryCount >= job.MaxRetries {
			continue
		}

		delay := s.retryPolicy.BaseDelay * time.Duration(math.Pow(s.retryPolicy.BackoffMultiplier, float64(job.RetryCount)))
		if delay > s.retryPolicy.MaxDelay {
			delay = s.retryPolicy.MaxDelay
		}

		time.Sleep(delay)

		result, err := s.Dispatch(ctx, job)
		if err != nil {
			result = &DispatchResult{
				JobID:   job.ID,
				Success: false,
				Channel: string(job.Channel),
				Error:   err.Error(),
				Attempt: job.RetryCount + 1,
			}
		}

		if result.Success {
			s.repo.UpdateStatus(ctx, job.ID, "completed", "")
		}

		results = append(results, result)
	}

	return results, nil
}
