package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name:        "신용점수 상승 사용자 정상 생성",
			email:       "Duser780641_29@example.fake",
			phoneNumber: "000-0420-2932",
			creditUp:    true,
			expectError: false,
		},
		{
			name:        "신용점수 하락 사용자 정상 생성",
			email:       "Duser206226_26@example.fake",
			phoneNumber: "000-1815-2005",
			creditUp:    false,
			expectError: false,
		},
		{
			name:        "빈 이메일",
			email:       "",
			phoneNumber: "000-1815-2005",
			creditUp:    true,
			expectError: true,
		},
		{
			name:        "빈 전화번호",
			email:       "Duser206226_26@example.fake",
			phoneNumber: "",
			creditUp:    true,
			expectError: true,
		},
		{
			name:        "공백 제거 처리",
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
				assert.Error(t, err)
				assert.Nil(t, user)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)

			// 필드 값 검증
			expectedEmail := strings.TrimSpace(tc.email)
			expectedPhone := strings.TrimSpace(tc.phoneNumber)

			assert.Equal(t, expectedEmail, user.Email)
			assert.Equal(t, expectedPhone, user.PhoneNumber)
			assert.Equal(t, tc.creditUp, user.CreditUp)
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
			name:     "신용점수 상승 - 알림 대상",
			creditUp: true,
			expected: true,
		},
		{
			name:     "신용점수 하락 - 알림 비대상",
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
			assert.Equal(t, tc.expected, result)
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
	assert.Equal(t, expectedKey, actualKey)
}

func TestUser_UniqueKeyByStrategy(t *testing.T) {
	// Given: 사용자 객체 준비
	user := &User{
		Email:       "user@example.com",
		PhoneNumber: "010-1234-5678",
		CreditUp:    true,
	}

	testCases := []struct {
		name     string
		strategy DuplicateStrategy
		expected string
	}{
		{
			name:     "이메일 기준 전략",
			strategy: ByEmail,
			expected: "user@example.com",
		},
		{
			name:     "전화번호 기준 전략",
			strategy: ByPhone,
			expected: "010-1234-5678",
		},
		{
			name:     "이메일+전화번호 기준 전략",
			strategy: ByBoth,
			expected: "user@example.com|010-1234-5678",
		},
		{
			name:     "잘못된 전략 (이메일 기본값)",
			strategy: DuplicateStrategy(999),
			expected: "user@example.com",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// When: 전략별 고유 키 조회
			actualKey := user.UniqueKeyByStrategy(tc.strategy)

			// Then: 결과 검증
			assert.Equal(t, tc.expected, actualKey)
		})
	}
}

func TestDuplicateStrategy_String(t *testing.T) {
	testCases := []struct {
		name     string
		strategy DuplicateStrategy
		expected string
	}{
		{
			name:     "이메일 전략 문자열 표현",
			strategy: ByEmail,
			expected: "ByEmail",
		},
		{
			name:     "전화번호 전략 문자열 표현",
			strategy: ByPhone,
			expected: "ByPhone",
		},
		{
			name:     "이메일+전화번호 전략 문자열 표현",
			strategy: ByBoth,
			expected: "ByBoth",
		},
		{
			name:     "알 수 없는 전략 문자열 표현",
			strategy: DuplicateStrategy(999),
			expected: "Unknown",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// When: String 메서드 호출
			result := tc.strategy.String()

			// Then: 결과 검증
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestUser_UniqueKeyByStrategy_RealWorldScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		user1       *User
		user2       *User
		strategy    DuplicateStrategy
		shouldMatch bool
		description string
	}{
		{
			name: "같은 이메일, 다른 전화번호 - 이메일 기준",
			user1: &User{
				Email:       "user@example.com",
				PhoneNumber: "010-1111-1111",
				CreditUp:    true,
			},
			user2: &User{
				Email:       "user@example.com",
				PhoneNumber: "010-2222-2222",
				CreditUp:    true,
			},
			strategy:    ByEmail,
			shouldMatch: true,
			description: "이메일 기준에서는 같은 사용자로 간주",
		},
		{
			name: "다른 이메일, 같은 전화번호 - 전화번호 기준",
			user1: &User{
				Email:       "user1@example.com",
				PhoneNumber: "010-1111-1111",
				CreditUp:    true,
			},
			user2: &User{
				Email:       "user2@example.com",
				PhoneNumber: "010-1111-1111",
				CreditUp:    true,
			},
			strategy:    ByPhone,
			shouldMatch: true,
			description: "전화번호 기준에서는 같은 사용자로 간주",
		},
		{
			name: "같은 이메일, 다른 전화번호 - 이메일+전화번호 기준",
			user1: &User{
				Email:       "user@example.com",
				PhoneNumber: "010-1111-1111",
				CreditUp:    true,
			},
			user2: &User{
				Email:       "user@example.com",
				PhoneNumber: "010-2222-2222",
				CreditUp:    true,
			},
			strategy:    ByBoth,
			shouldMatch: false,
			description: "이메일+전화번호 기준에서는 다른 사용자로 간주",
		},
		{
			name: "완전히 동일한 사용자 - 이메일+전화번호 기준",
			user1: &User{
				Email:       "user@example.com",
				PhoneNumber: "010-1111-1111",
				CreditUp:    true,
			},
			user2: &User{
				Email:       "user@example.com",
				PhoneNumber: "010-1111-1111",
				CreditUp:    true,
			},
			strategy:    ByBoth,
			shouldMatch: true,
			description: "모든 정보가 동일한 경우",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// When: 각 사용자의 고유 키 생성
			key1 := tc.user1.UniqueKeyByStrategy(tc.strategy)
			key2 := tc.user2.UniqueKeyByStrategy(tc.strategy)

			// Then: 매칭 여부 검증
			isMatch := key1 == key2
			assert.Equal(t, tc.shouldMatch, isMatch, tc.description)

			// 로그 출력 (디버깅용)
			t.Logf("%s: key1=%s, key2=%s, match=%v", tc.description, key1, key2, isMatch)
		})
	}
}
