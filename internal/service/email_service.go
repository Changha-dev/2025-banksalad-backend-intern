package service

import (
	"context"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"banksalad-backend-task/clients"
	"banksalad-backend-task/internal/domain"
)

type EmailSender interface {
	Send(email string, message string) error
}

type EmailService interface {
	SendEmails(ctx context.Context, users []*domain.User) (int, error)
}

type emailService struct {
	client EmailSender
}

func NewEmailService() EmailService {
	client := clients.NewEmailClient()
	return &emailService{
		client: client,
	}
}

func NewEmailServiceWithClient(client EmailSender) EmailService {
	return &emailService{
		client: client,
	}
}

func (es *emailService) SendEmails(ctx context.Context, users []*domain.User) (int, error) {
	if len(users) == 0 {
		return 0, nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(users))
	successCount := int64(0)
	failureCount := int64(0)

	for _, user := range users {
		wg.Add(1)
		go func(u *domain.User) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.WithField("panic", r).Error("recovered from panic")
					atomic.AddInt64(&failureCount, 1)
				}
			}()

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				if err := es.client.Send(u.Email, "신용점수 상승 알림"); err != nil {
					log.WithError(err).WithField("email", u.Email).Error("이메일 전송 실패 (계속 진행)")
					atomic.AddInt64(&failureCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(user)
	}

	wg.Wait()
	close(errChan)

	// 컨텍스트 취소 에러만 반환, 개별 전송 실패는 무시
	for err := range errChan {
		if err == context.Canceled || err == context.DeadlineExceeded {
			return int(successCount), err
		}
	}

	log.WithFields(log.Fields{
		"success": successCount,
		"total":   len(users),
		"failure": failureCount,
	}).Info("이메일 전송 완료")

	return int(successCount), nil
}
