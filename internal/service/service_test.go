package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"banksalad-backend-task/internal/domain"
)

func setupTestDir(t *testing.T) string {
	t.Helper()

	// 임시 디렉토리 생성
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("banksalad_test_%d", time.Now().UnixNano()))
	err := os.MkdirAll(filepath.Join(tmpDir, "files", "output"), 0755)
	require.NoError(t, err, "테스트 디렉토리 생성 실패")

	// 원래 작업 디렉토리 저장
	originalWd, _ := os.Getwd()

	// 임시 디렉토리로 이동
	err = os.Chdir(tmpDir)
	require.NoError(t, err, "작업 디렉토리 변경 실패")

	// 테스트 완료 후 정리
	t.Cleanup(func() {
		os.Chdir(originalWd)
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

type MockEmailClient struct {
	sentEmails []string
	shouldFail bool
}

func (m *MockEmailClient) Send(email string, message string) error {
	if m.shouldFail {
		return fmt.Errorf("이메일 전송 실패")
	}
	m.sentEmails = append(m.sentEmails, email)
	return nil
}

type MockSMSClient struct {
	sentSMS    []string
	shouldFail bool
}

func (m *MockSMSClient) Send(phoneNumber string, message string) error {
	if m.shouldFail {
		return fmt.Errorf("SMS 전송 실패")
	}
	m.sentSMS = append(m.sentSMS, phoneNumber)
	return nil
}

func TestRateLimiter_Wait(t *testing.T) {
	// Given: 초당 2개 토큰을 가진 속도 제한기 생성
	rateLimiter := NewRateLimiter(2, time.Second)
	t.Cleanup(func() {
		rateLimiter.Stop()
	})

	ctx := context.Background()

	testCases := []struct {
		name        string
		waitCount   int
		expectError bool
	}{
		{
			name:        "토큰 용량 내 요청",
			waitCount:   2,
			expectError: false,
		},
		{
			name:        "토큰 용량 초과 요청",
			waitCount:   3,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// When & Then: 토큰 요청 및 검증
			for i := 0; i < tc.waitCount; i++ {
				err := rateLimiter.Wait(ctx)

				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestEmailService_SendEmails_Unit(t *testing.T) {
	testCases := []struct {
		name            string
		users           []*domain.User
		shouldFail      bool
		expectedSuccess int
	}{
		{
			name:            "빈 사용자 목록",
			users:           nil,
			shouldFail:      false,
			expectedSuccess: 0,
		},
		{
			name:            "정상적인 이메일 전송",
			users:           createTestUsers(3),
			shouldFail:      false,
			expectedSuccess: 3,
		},
		{
			name:            "이메일 전송 실패",
			users:           createTestUsers(2),
			shouldFail:      true,
			expectedSuccess: 0, // 모든 전송이 실패
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Given: Mock 클라이언트를 사용한 이메일 서비스 생성
			mockClient := &MockEmailClient{shouldFail: tc.shouldFail}
			emailService := NewEmailServiceWithClient(mockClient)
			ctx := context.Background()

			// When: 이메일 전송 실행
			successCount, err := emailService.SendEmails(ctx, tc.users)

			// Then: 결과 검증
			// 컨텍스트 에러가 아닌 경우는 개별 전송 실패로 에러를 반환하지 않음
			if err != nil {
				assert.True(t, err == context.Canceled || err == context.DeadlineExceeded)
			}

			assert.Equal(t, tc.expectedSuccess, successCount)

			if !tc.shouldFail && len(tc.users) > 0 {
				assert.Len(t, mockClient.sentEmails, len(tc.users))
			}
		})
	}
}

func TestEmailService_SendEmails_Integration(t *testing.T) {
	// Given: 테스트 환경 설정
	setupTestDir(t)

	testCases := []struct {
		name  string
		users []*domain.User
	}{
		{
			name:  "빈 사용자 목록",
			users: nil,
		},
		{
			name:  "정상적인 이메일 전송",
			users: createTestUsers(3),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Given: 실제 이메일 서비스 생성
			emailService := NewEmailService()
			ctx := context.Background()

			// When: 이메일 전송 실행
			successCount, err := emailService.SendEmails(ctx, tc.users)

			// Then: 결과 검증
			if err != nil {
				assert.True(t, err == context.Canceled || err == context.DeadlineExceeded)
			}

			// 0.5% 에러율로 인해 일부 실패할 수 있음
			if len(tc.users) > 0 {
				assert.Greater(t, successCount, 0, "모든 이메일 전송이 실패함")
			}

			assert.LessOrEqual(t, successCount, len(tc.users))
		})
	}
}

func TestSMSService_SendSMS_Unit(t *testing.T) {
	testCases := []struct {
		name            string
		users           []*domain.User
		shouldFail      bool
		expectedSuccess int
	}{
		{
			name:            "빈 사용자 목록",
			users:           nil,
			shouldFail:      false,
			expectedSuccess: 0,
		},
		{
			name:            "정상적인 SMS 전송",
			users:           createTestUsers(2),
			shouldFail:      false,
			expectedSuccess: 2,
		},
		{
			name:            "SMS 전송 실패",
			users:           createTestUsers(1),
			shouldFail:      true,
			expectedSuccess: 0, // 모든 전송이 실패
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Given: Mock 클라이언트를 사용한 SMS 서비스 생성
			mockClient := &MockSMSClient{shouldFail: tc.shouldFail}
			smsService := NewSMSServiceWithClient(mockClient)
			t.Cleanup(func() {
				smsService.Stop()
			})
			ctx := context.Background()

			// When: SMS 전송 실행
			successCount, err := smsService.SendSMS(ctx, tc.users)

			// Then: 결과 검증
			if err != nil {
				assert.True(t, err == context.Canceled || err == context.DeadlineExceeded)
			}

			assert.Equal(t, tc.expectedSuccess, successCount)

			if !tc.shouldFail && len(tc.users) > 0 {
				assert.Len(t, mockClient.sentSMS, len(tc.users))
			}
		})
	}
}

func TestSMSService_SendSMS_Integration(t *testing.T) {
	// Given: 테스트 환경 설정
	setupTestDir(t)

	testCases := []struct {
		name  string
		users []*domain.User
	}{
		{
			name:  "빈 사용자 목록",
			users: nil,
		},
		{
			name:  "정상적인 SMS 전송",
			users: createTestUsers(2),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Given: 실제 SMS 서비스 생성
			smsService := NewSMSService()
			t.Cleanup(func() {
				smsService.Stop()
			})
			ctx := context.Background()

			// When: SMS 전송 실행
			successCount, err := smsService.SendSMS(ctx, tc.users)

			// Then: 결과 검증
			if err != nil {
				assert.True(t, err == context.Canceled || err == context.DeadlineExceeded)
			}

			// 0.5% 에러율로 인해 일부 실패할 수 있음
			if len(tc.users) > 0 {
				assert.Greater(t, successCount, 0, "모든 SMS 전송이 실패함")
			}

			assert.LessOrEqual(t, successCount, len(tc.users))
		})
	}
}

func TestNotificationManager_SendNotifications_Integration(t *testing.T) {
	// Given: 테스트 환경 설정
	setupTestDir(t)

	// Given: 실제 알림 매니저 생성
	manager := NewNotificationManager()
	users := createTestUsers(3)
	ctx := context.Background()

	// When: 알림 전송 실행
	emailSuccess, smsSuccess, err := manager.SendNotifications(ctx, users)

	// Then: 결과 검증
	if err != nil {
		assert.True(t, err == context.Canceled || err == context.DeadlineExceeded)
	}

	// 0.5% 에러율로 인해 일부 실패할 수 있지만, 모든 전송이 실패하면 안 됨
	if len(users) > 0 {
		assert.Greater(t, emailSuccess, 0, "모든 이메일 전송이 실패함")
		assert.Greater(t, smsSuccess, 0, "모든 SMS 전송이 실패함")
	}

	assert.LessOrEqual(t, emailSuccess, len(users))
	assert.LessOrEqual(t, smsSuccess, len(users))
}

func TestNotificationManager_SendNotifications_Unit(t *testing.T) {
	testCases := []struct {
		name            string
		users           []*domain.User
		emailShouldFail bool
		smsShouldFail   bool
		expectError     bool
	}{
		{
			name:            "빈 사용자 목록",
			users:           nil,
			emailShouldFail: false,
			smsShouldFail:   false,
			expectError:     false,
		},
		{
			name:            "정상적인 알림 전송",
			users:           createTestUsers(2),
			emailShouldFail: false,
			smsShouldFail:   false,
			expectError:     false,
		},
		{
			name:            "이메일 실패, SMS 성공",
			users:           createTestUsers(2),
			emailShouldFail: true,
			smsShouldFail:   false,
			expectError:     false, // 개별 실패는 에러로 반환하지 않음
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Given: Mock 서비스들을 사용한 알림 매니저 생성
			mockEmailClient := &MockEmailClient{shouldFail: tc.emailShouldFail}
			mockSMSClient := &MockSMSClient{shouldFail: tc.smsShouldFail}

			emailService := NewEmailServiceWithClient(mockEmailClient)
			smsService := NewSMSServiceWithClient(mockSMSClient)

			manager := &NotificationManager{
				emailService: emailService,
				smsService:   smsService,
			}

			t.Cleanup(func() {
				manager.Close()
			})

			ctx := context.Background()

			// When: 알림 전송 실행
			emailSuccess, smsSuccess, err := manager.SendNotifications(ctx, tc.users)

			// Then: 결과 검증
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 성공 수 검증
			expectedEmailSuccess := 0
			expectedSMSSuccess := 0
			if !tc.emailShouldFail {
				expectedEmailSuccess = len(tc.users)
			}
			if !tc.smsShouldFail {
				expectedSMSSuccess = len(tc.users)
			}

			assert.Equal(t, expectedEmailSuccess, emailSuccess)
			assert.Equal(t, expectedSMSSuccess, smsSuccess)
		})
	}
}

// 테스트 헬퍼 함수
func createTestUsers(count int) []*domain.User {
	users := make([]*domain.User, count)

	for i := 0; i < count; i++ {
		email := fmt.Sprintf("test%d@example.com", i)
		phone := fmt.Sprintf("010-1234-%04d", i)
		user, _ := domain.NewUser(email, phone, true)
		users[i] = user
	}

	return users
}
