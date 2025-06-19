package domain

import (
	"strings"
	"testing"
)

func TestNewUser(t *testing.T) {
	testCases := []struct {
		name        string
		email       string
		phoneNumber string
		creditUp    bool
		expectError bool
	}{
		{
			name:        "valid user with credit up",
			email:       "Duser780641_29@example.fake",
			phoneNumber: "000-0420-2932",
			creditUp:    true,
			expectError: false,
		},
		{
			name:        "valid user with credit down",
			email:       "Duser206226_26@example.fake",
			phoneNumber: "000-1815-2005",
			creditUp:    false,
			expectError: false,
		},
		{
			name:        "empty email",
			email:       "",
			phoneNumber: "000-1815-2005",
			creditUp:    true,
			expectError: true,
		},
		{
			name:        "empty phone number",
			email:       "Duser206226_26@example.fake",
			phoneNumber: "",
			creditUp:    true,
			expectError: true,
		},
		{
			name:        "whitespace trimming",
			email:       "  Duser206226_26@example.fake  ",
			phoneNumber: "  000-1815-2005  ",
			creditUp:    true,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Given: 테스트 데이터 준비
			email := tc.email
			phoneNumber := tc.phoneNumber
			creditUp := tc.creditUp

			// When: 사용자 생성 실행
			user, err := NewUser(email, phoneNumber, creditUp)

			// Then: 결과 검증
			if tc.expectError {
				// 에러가 예상되는 경우
				if err == nil {
					t.Error("expected error but got none")
				}
				if user != nil {
					t.Error("expected nil user but got one")
				}
				return
			}

			// 성공이 예상되는 경우
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("expected user but got nil")
				return
			}

			// 필드 값 검증
			expectedEmail := strings.TrimSpace(tc.email)
			expectedPhone := strings.TrimSpace(tc.phoneNumber)

			if user.Email != expectedEmail {
				t.Errorf("expected email %s, got %s", expectedEmail, user.Email)
			}

			if user.PhoneNumber != expectedPhone {
				t.Errorf("expected phone %s, got %s", expectedPhone, user.PhoneNumber)
			}

			if user.CreditUp != tc.creditUp {
				t.Errorf("expected credit up %v, got %v", tc.creditUp, user.CreditUp)
			}
		})
	}
}

func TestUser_IsEligibleForNotification(t *testing.T) {
	testCases := []struct {
		name     string
		creditUp bool
		expected bool
	}{
		{
			name:     "credit up - eligible",
			creditUp: true,
			expected: true,
		},
		{
			name:     "credit down - not eligible",
			creditUp: false,
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Given: 사용자 객체 준비
			user := &User{
				Email:       "Duser206226_26@example.fake",
				PhoneNumber: "000-1815-2005",
				CreditUp:    tc.creditUp,
			}

			// When: 알림 대상 여부 확인
			result := user.IsEligibleForNotification()

			// Then: 결과 검증
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestUser_UniqueKey(t *testing.T) {
	// Given: 사용자 객체 준비
	user := &User{
		Email:       "Duser206226_26@example.fake",
		PhoneNumber: "000-1815-2005",
		CreditUp:    true,
	}
	expectedKey := "Duser206226_26@example.fake"

	// When: 고유 키 조회
	actualKey := user.UniqueKey()

	// Then: 결과 검증
	if actualKey != expectedKey {
		t.Errorf("expected %s, got %s", expectedKey, actualKey)
	}
}
