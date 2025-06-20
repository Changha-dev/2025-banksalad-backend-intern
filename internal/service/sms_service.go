package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"banksalad-backend-task/clients"
	"banksalad-backend-task/internal/domain"
)

type SMSSender interface {
	Send(phoneNumber string, message string) error
}

type SMSService interface {
	SendSMS(ctx context.Context, users []*domain.User) (int, error)
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

func (ss *smsService) SendSMS(ctx context.Context, users []*domain.User) (int, error) {
	if len(users) == 0 {
		return 0, nil
	}

	successCount := 0
	failureCount := 0

	for _, user := range users {
		select {
		case <-ctx.Done():
			return successCount, ctx.Err()
		default:
			// 속도 제한 대기
			if err := ss.rateLimiter.Wait(ctx); err != nil {
				return successCount, fmt.Errorf("속도 제한 대기 중 오류: %w", err)
			}

			if err := ss.client.Send(user.PhoneNumber, "신용점수 상승 알림"); err != nil {
				// 에러를 로그로 기록하고 계속 진행
				log.Printf("SMS 전송 실패 (계속 진행): %s - %v", user.PhoneNumber, err)
				failureCount++
			} else {
				successCount++
			}
		}
	}

	log.Printf("✓ SMS 전송 완료: %d/%d명 성공, %d명 실패", successCount, len(users), failureCount)
	return successCount, nil
}

func (ss *smsService) Stop() {
	ss.rateLimiter.Stop()
}
