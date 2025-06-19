package service

import (
	"context"
	"fmt"
	"time"

	"banksalad-backend-task/clients"
	"banksalad-backend-task/internal/domain"
)

type SMSSender interface {
	Send(phoneNumber string, message string) error
}

type SMSService interface {
	SendSMS(ctx context.Context, users []*domain.User) error
	Stop()
}

type smsService struct {
	client      SMSSender
	rateLimiter *RateLimiter
}

func NewSMSService() SMSService {
	client := clients.NewSmsClient()
	rateLimiter := NewRateLimiter(100, time.Second)

	return &smsService{
		client:      client,
		rateLimiter: rateLimiter,
	}
}

func NewSMSServiceWithClient(client SMSSender) SMSService {
	rateLimiter := NewRateLimiter(100, time.Second)
	return &smsService{
		client:      client,
		rateLimiter: rateLimiter,
	}
}

func (ss *smsService) SendSMS(ctx context.Context, users []*domain.User) error {
	if len(users) == 0 {
		return nil
	}

	for _, user := range users {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := ss.rateLimiter.Wait(ctx); err != nil {
				return fmt.Errorf("속도 제한 대기 중 오류: %w", err)
			}

			if err := ss.client.Send(user.PhoneNumber, "신용점수 상승 알림"); err != nil {
				return fmt.Errorf("SMS 전송 실패 %s: %w", user.PhoneNumber, err)
			}
		}
	}

	return nil
}

func (ss *smsService) Stop() {
	ss.rateLimiter.Stop()
}
