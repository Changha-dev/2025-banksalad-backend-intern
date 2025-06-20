package processor

import (
	"banksalad-backend-task/internal/domain"
)

type DuplicateFilter struct {
	processed map[string]struct{}
	strategy  domain.DuplicateStrategy
}

func NewDuplicateFilter() *DuplicateFilter {
	return &DuplicateFilter{
		processed: make(map[string]struct{}),
		strategy:  domain.ByEmail, // 기본값
	}
}

// 전략을 지정하는 생성자
func NewDuplicateFilterWithStrategy(strategy domain.DuplicateStrategy) *DuplicateFilter {
	return &DuplicateFilter{
		processed: make(map[string]struct{}),
		strategy:  strategy,
	}
}

// 전략 변경 메서드
func (df *DuplicateFilter) SetStrategy(strategy domain.DuplicateStrategy) {
	df.strategy = strategy
}

// 현재 전략 조회
func (df *DuplicateFilter) GetStrategy() domain.DuplicateStrategy {
	return df.strategy
}

func (df *DuplicateFilter) FilterDuplicates(users []*domain.User) []*domain.User {
	if len(users) == 0 {
		return nil
	}

	unique := make([]*domain.User, 0, len(users))

	for _, user := range users {
		key := user.UniqueKeyByStrategy(df.strategy) // 전략 사용

		if _, exists := df.processed[key]; !exists {
			df.processed[key] = struct{}{}
			unique = append(unique, user)
		}
	}

	if len(unique) == 0 {
		return nil
	}

	return unique
}

func (df *DuplicateFilter) Reset() {
	df.processed = make(map[string]struct{})
}

func (df *DuplicateFilter) GetProcessedCount() int {
	return len(df.processed)
}

func (df *DuplicateFilter) IsProcessed(user *domain.User) bool {
	key := user.UniqueKeyByStrategy(df.strategy)
	_, exists := df.processed[key]
	return exists
}
