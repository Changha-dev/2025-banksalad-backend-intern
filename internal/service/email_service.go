package service

import (
	"context"
	"fmt"
	"sync"

	"banksalad-backend-task/clients"
	"banksalad-backend-task/internal/domain"
)

type EmailSender interface {
	Send(email string, message string) error
}

type EmailService interface {
	SendEmails(ctx context.Context, users []*domain.User) error
}

type emailService struct {
	client EmailSender // 인터페이스 타입으로 변경
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

func (es *emailService) SendEmails(ctx context.Context, users []*domain.User) error {
	if len(users) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(users))

	for _, user := range users {
		wg.Add(1)
		go func(u *domain.User) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				if err := es.client.Send(u.Email, "신용점수 상승 알림"); err != nil {
					errChan <- fmt.Errorf("이메일 전송 실패 %s: %w", u.Email, err)
				}
			}
		}(user)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
