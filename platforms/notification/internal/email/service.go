package email

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
)

type Service interface {
	SendEmail(ctx context.Context, req *SendEmailRequest) (*EmailMessage, error)
	SendTemplatedEmail(ctx context.Context, req *SendEmailRequest, templateName string, data interface{}) (*EmailMessage, error)
	BulkEmail(ctx context.Context, reqs []*SendEmailRequest) ([]*EmailMessage, error)
	GetEmail(ctx context.Context, id string) (*EmailMessage, error)
	TrackBounce(ctx context.Context, id string) error
	TrackOpen(ctx context.Context, id string) error
}

type service struct {
	repo    Repository
	pub     events.Publisher
	from    string
	tmplEng *TemplateEngine
}

func NewService(repo Repository, pub events.Publisher, from string) Service {
	return &service{
		repo:    repo,
		pub:     pub,
		from:    from,
		tmplEng: NewTemplateEngine(),
	}
}

func (s *service) SendEmail(ctx context.Context, req *SendEmailRequest) (*EmailMessage, error) {
	msg := &EmailMessage{
		To:        req.To,
		CC:        req.CC,
		BCC:       req.BCC,
		From:      s.from,
		ReplyTo:   req.ReplyTo,
		Subject:   req.Subject,
		PlainText: req.PlainText,
		HTML:      req.HTML,
		Status:    EmailStatusSent,
		Attachments: req.Attachments,
		Metadata:  req.Metadata,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, err
	}

	s.pub.Publish(ctx, events.EventNotificationSent, &events.NotificationSentEvent{
		NotificationID: msg.ID,
		Channel:        "email",
		Type:           "email",
		SentAt:         time.Now(),
	})

	return msg, nil
}

func (s *service) SendTemplatedEmail(ctx context.Context, req *SendEmailRequest, templateName string, data interface{}) (*EmailMessage, error) {
	if req.HTML == "" {
		html, err := s.tmplEng.Render(templateName, data)
		if err != nil {
			return nil, err
		}
		req.HTML = html
	}

	if req.PlainText == "" {
		plainText, err := s.tmplEng.Render(templateName+"_text", data)
		if err == nil && plainText != "" {
			req.PlainText = plainText
		}
	}

	return s.SendEmail(ctx, req)
}

func (s *service) BulkEmail(ctx context.Context, reqs []*SendEmailRequest) ([]*EmailMessage, error) {
	results := make([]*EmailMessage, 0, len(reqs))
	for _, req := range reqs {
		msg, err := s.SendEmail(ctx, req)
		if err != nil {
			continue
		}
		results = append(results, msg)
	}
	return results, nil
}

func (s *service) GetEmail(ctx context.Context, id string) (*EmailMessage, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) TrackBounce(ctx context.Context, id string) error {
	if err := s.repo.UpdateStatus(ctx, id, EmailStatusBounced); err != nil {
		return err
	}

	msg, _ := s.repo.GetByID(ctx, id)
	if msg != nil && len(msg.To) > 0 {
		s.pub.Publish(ctx, events.EventEmailBounced, &events.EmailBouncedEvent{
			EmailID:  id,
			To:       msg.To[0],
			Reason:   "bounced",
			BouncedAt: time.Now(),
		})
	}

	return nil
}

func (s *service) TrackOpen(ctx context.Context, id string) error {
	if err := s.repo.UpdateStatus(ctx, id, EmailStatusOpened); err != nil {
		return err
	}

	msg, _ := s.repo.GetByID(ctx, id)
	if msg != nil && len(msg.To) > 0 {
		s.pub.Publish(ctx, events.EventEmailOpened, &events.EmailOpenedEvent{
			EmailID:  id,
			To:       msg.To[0],
			OpenedAt: time.Now(),
		})
	}

	return nil
}

func (s *service) RegisterTemplate(name, tmpl string) error {
	return s.tmplEng.Register(name, tmpl)
}
