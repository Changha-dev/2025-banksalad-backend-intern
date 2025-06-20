package processor

import (
	"banksalad-backend-task/internal/domain"
)

type DuplicateFilter struct {
	processed map[string]struct{}
}

func NewDuplicateFilter() *DuplicateFilter {
	return &DuplicateFilter{
		processed: make(map[string]struct{}),
	}
}

func (df *DuplicateFilter) FilterDuplicates(users []*domain.User) []*domain.User {
	if len(users) == 0 {
		return nil
	}

	unique := make([]*domain.User, 0, len(users))

	for _, user := range users {
		key := user.UniqueKey()

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
	_, exists := df.processed[user.UniqueKey()]
	return exists
}
