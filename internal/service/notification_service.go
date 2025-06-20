package service

import (
	"context"
	"sync"

	"banksalad-backend-task/internal/domain"
)

type NotificationManager struct {
	emailService EmailService
	smsService   SMSService
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		emailService: NewEmailService(),
		smsService:   NewSMSService(),
	}
}

func (nm *NotificationManager) SendNotifications(ctx context.Context, users []*domain.User) (int, int, error) {
	if len(users) == 0 {
		return 0, 0, nil
	}

	var wg sync.WaitGroup
	var emailSuccess, smsSuccess int
	var emailErr, smsErr error

	// 이메일 전송
	wg.Add(1)
	go func() {
		defer wg.Done()
		emailSuccess, emailErr = nm.emailService.SendEmails(ctx, users)
	}()

	// SMS 전송
	wg.Add(1)
	go func() {
		defer wg.Done()
		smsSuccess, smsErr = nm.smsService.SendSMS(ctx, users)
	}()

	wg.Wait()

	// 에러가 있으면 첫 번째 에러 반환
	if emailErr != nil {
		return emailSuccess, smsSuccess, emailErr
	}
	if smsErr != nil {
		return emailSuccess, smsSuccess, smsErr
	}

	return emailSuccess, smsSuccess, nil
}

func (nm *NotificationManager) Close() error {
	if nm.smsService != nil {
		nm.smsService.Stop()
	}
	return nil
}
