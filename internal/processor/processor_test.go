package processor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"banksalad-backend-task/internal/domain"
)

func TestCreditProcessor_FilterEligibleUsers(t *testing.T) {
	// Given: 신용점수 프로세서 생성
	processor := NewCreditProcessor()

	testCases := []struct {
		name          string
		users         []*domain.User
		expectedCount int
	}{
		{
			name:          "빈 사용자 목록",
			users:         nil,
			expectedCount: 0,
		},
		{
			name:          "신용점수 상승 사용자만 있는 경우",
			users:         createTestUsers([]bool{true, true, true}),
			expectedCount: 3,
		},
		{
			name:          "신용점수 하락 사용자만 있는 경우",
			users:         createTestUsers([]bool{false, false, false}),
			expectedCount: 0,
		},
		{
			name:          "혼합된 사용자 목록",
			users:         createTestUsers([]bool{true, false, true, false, true}),
			expectedCount: 3,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// When: 신용점수 상승 사용자 필터링 실행
			eligible := processor.FilterEligibleUsers(tc.users)

			// Then: 결과 검증
			if tc.expectedCount == 0 {
				assert.Nil(t, eligible)
				return
			}

			require.NotNil(t, eligible)
			assert.Len(t, eligible, tc.expectedCount)

			// 모든 반환된 사용자가 신용점수 상승 사용자인지 확인
			for i, user := range eligible {
				assert.True(t, user.IsEligibleForNotification(),
					"%d번째 사용자는 신용점수가 상승하지 않았음", i)
			}
		})
	}
}

func TestCreditProcessor_CountEligibleUsers(t *testing.T) {
	// Given: 신용점수 프로세서 생성
	processor := NewCreditProcessor()

	testCases := []struct {
		name          string
		users         []*domain.User
		expectedCount int
	}{
		{
			name:          "빈 사용자 목록 카운트",
			users:         nil,
			expectedCount: 0,
		},
		{
			name:          "신용점수 상승 사용자 5명",
			users:         createTestUsers([]bool{true, true, true, true, true}),
			expectedCount: 5,
		},
		{
			name:          "혼합된 사용자 목록 카운트",
			users:         createTestUsers([]bool{true, false, true, false, true, false}),
			expectedCount: 3,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// When: 신용점수 상승 사용자 카운트 실행
			count := processor.CountEligibleUsers(tc.users)

			// Then: 결과 검증
			assert.Equal(t, tc.expectedCount, count)
		})
	}
}

func TestDuplicateFilter_FilterDuplicates(t *testing.T) {
	// Given: 중복 필터 생성
	filter := NewDuplicateFilter()

	testCases := []struct {
		name          string
		users         []*domain.User
		expectedCount int
	}{
		{
			name:          "빈 사용자 목록",
			users:         nil,
			expectedCount: 0,
		},
		{
			name:          "중복 없는 사용자 목록",
			users:         createUniqueTestUsers(3),
			expectedCount: 3,
		},
		{
			name:          "완전 중복된 사용자 목록",
			users:         createDuplicateTestUsers("test@example.com", "010-1234-5678", 3),
			expectedCount: 1,
		},
		{
			name:          "일부 중복된 사용자 목록",
			users:         createMixedTestUsers(),
			expectedCount: 3,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// 각 테스트마다 필터 리셋
			filter.Reset()

			// When: 중복 제거 실행
			unique := filter.FilterDuplicates(tc.users)

			// Then: 결과 검증
			if tc.expectedCount == 0 {
				assert.Nil(t, unique)
				return
			}

			require.NotNil(t, unique)
			assert.Len(t, unique, tc.expectedCount)

			// 중복이 제거되었는지 확인 (뱅크샐러드 가이드 준수: map 순회 없음)
			emailSet := make(map[string]bool)
			for _, user := range unique {
				assert.False(t, emailSet[user.Email], "중복된 이메일이 발견됨: %s", user.Email)
				emailSet[user.Email] = true
			}
		})
	}
}

func TestDuplicateFilter_Reset(t *testing.T) {
	// Given: 중복 필터 생성 및 사용자 처리
	filter := NewDuplicateFilter()
	users := createUniqueTestUsers(3)
	filter.FilterDuplicates(users)

	// When: 필터 리셋 실행
	filter.Reset()

	// Then: 처리된 사용자 수가 0인지 확인
	assert.Equal(t, 0, filter.GetProcessedCount())
}

func TestDuplicateFilter_IsProcessed(t *testing.T) {
	// Given: 중복 필터 생성 및 사용자 생성
	filter := NewDuplicateFilter()
	user1, err := domain.NewUser("test1@example.com", "010-1111-1111", true)
	require.NoError(t, err)
	user2, err := domain.NewUser("test2@example.com", "010-2222-2222", true)
	require.NoError(t, err)

	// When: 첫 번째 사용자만 처리
	filter.FilterDuplicates([]*domain.User{user1})

	// Then: 처리 여부 확인
	assert.True(t, filter.IsProcessed(user1))
	assert.False(t, filter.IsProcessed(user2))
}

func TestDuplicateFilter_FilterDuplicates_WithDifferentStrategies(t *testing.T) {
	testCases := []struct {
		name          string
		strategy      domain.DuplicateStrategy
		users         []*domain.User
		expectedCount int
	}{
		{
			name:          "이메일 기준 중복 제거",
			strategy:      domain.ByEmail,
			users:         createSameEmailDifferentPhoneUsers(),
			expectedCount: 1,
		},
		{
			name:          "전화번호 기준 중복 제거",
			strategy:      domain.ByPhone,
			users:         createDifferentEmailSamePhoneUsers(),
			expectedCount: 1,
		},
		{
			name:          "이메일+전화번호 기준 중복 제거",
			strategy:      domain.ByBoth,
			users:         createSameEmailDifferentPhoneUsers(),
			expectedCount: 2, // 전화번호가 다르므로 별도 처리
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Given: 전략을 지정한 필터 생성
			filter := NewDuplicateFilterWithStrategy(tc.strategy)

			// When: 중복 제거 실행
			unique := filter.FilterDuplicates(tc.users)

			// Then: 결과 검증
			assert.Len(t, unique, tc.expectedCount)
			assert.Equal(t, tc.strategy, filter.GetStrategy())
		})
	}
}

// 테스트 헬퍼 함수들

func createTestUsers(creditUpStates []bool) []*domain.User {
	users := make([]*domain.User, len(creditUpStates))

	for i, creditUp := range creditUpStates {
		email := fmt.Sprintf("test%d@example.com", i)
		phone := fmt.Sprintf("010-1234-%04d", i)
		user, _ := domain.NewUser(email, phone, creditUp)
		users[i] = user
	}

	return users
}

func createUniqueTestUsers(count int) []*domain.User {
	users := make([]*domain.User, count)

	for i := 0; i < count; i++ {
		email := fmt.Sprintf("unique%d@example.com", i)
		phone := fmt.Sprintf("010-1111-%04d", i)
		user, _ := domain.NewUser(email, phone, true)
		users[i] = user
	}

	return users
}

func createDuplicateTestUsers(email, phone string, count int) []*domain.User {
	users := make([]*domain.User, count)

	for i := 0; i < count; i++ {
		user, _ := domain.NewUser(email, phone, true)
		users[i] = user
	}

	return users
}

func createMixedTestUsers() []*domain.User {
	// 3개의 고유한 사용자와 2개의 중복 사용자 생성
	user1, _ := domain.NewUser("user1@example.com", "010-1111-1111", true)
	user2, _ := domain.NewUser("user2@example.com", "010-2222-2222", true)
	user3, _ := domain.NewUser("user3@example.com", "010-3333-3333", true)
	user1Duplicate, _ := domain.NewUser("user1@example.com", "010-1111-1111", true)
	user2Duplicate, _ := domain.NewUser("user2@example.com", "010-2222-2222", true)

	return []*domain.User{user1, user2, user3, user1Duplicate, user2Duplicate}
}

func createSameEmailDifferentPhoneUsers() []*domain.User {
	// 같은 이메일, 다른 전화번호 (전화번호 변경 시나리오)
	user1, _ := domain.NewUser("user@example.com", "010-1111-1111", true)
	user2, _ := domain.NewUser("user@example.com", "010-2222-2222", true) // 전화번호 변경

	return []*domain.User{user1, user2}
}

func createDifferentEmailSamePhoneUsers() []*domain.User {
	// 다른 이메일, 같은 전화번호 (가족 공유폰 시나리오)
	user1, _ := domain.NewUser("father@example.com", "010-1111-1111", true)
	user2, _ := domain.NewUser("mother@example.com", "010-1111-1111", true) // 같은 전화번호

	return []*domain.User{user1, user2}
}
