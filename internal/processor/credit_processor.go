package processor

import (
	"banksalad-backend-task/internal/domain"
)

type CreditProcessor struct{}

func NewCreditProcessor() *CreditProcessor {
	return &CreditProcessor{}
}

func (cp *CreditProcessor) FilterEligibleUsers(users []*domain.User) []*domain.User {
	if len(users) == 0 {
		return nil
	}

	// 신용점수 상승 사용자만 필터링하기 위한 슬라이스 생성
	eligible := make([]*domain.User, 0, len(users))

	for _, user := range users {
		if user.IsEligibleForNotification() {
			eligible = append(eligible, user)
		}
	}

	// 빈 결과인 경우 nil 반환
	if len(eligible) == 0 {
		return nil
	}

	return eligible
}

func (cp *CreditProcessor) CountEligibleUsers(users []*domain.User) int {
	if len(users) == 0 {
		return 0
	}

	count := 0
	for _, user := range users {
		if user.IsEligibleForNotification() {
			count++
		}
	}

	return count
}
