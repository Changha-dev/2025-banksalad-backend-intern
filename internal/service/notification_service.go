package service

import (
	"context"
	"fmt"
	"sync"

	"banksalad-backend-task/internal/domain"
)

type NotificationService interface {
	SendNotifications(ctx context.Context, users []*domain.User) error
}

type NotificationManager struct {
	emailService EmailService
	smsService   SMSService
}

func NewNotificationManager() *NotificationManager {
	emailService := NewEmailService()
	smsService := NewSMSService()

	return &NotificationManager{
		emailService: emailService,
		smsService:   smsService,
	}
}

func (nm *NotificationManager) SendNotifications(ctx context.Context, users []*domain.User) error {
	if len(users) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// 이메일 알림 병렬 전송
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := nm.emailService.SendEmails(ctx, users); err != nil {
			errChan <- fmt.Errorf("이메일 알림 전송 실패: %w", err)
		}
	}()

	// SMS 알림 속도 제한 전송
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := nm.smsService.SendSMS(ctx, users); err != nil {
			errChan <- fmt.Errorf("SMS 알림 전송 실패: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// 에러 확인
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
