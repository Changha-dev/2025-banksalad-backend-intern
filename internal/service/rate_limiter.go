package service

import (
	"context"
	"time"
)

type RateLimiter struct {
	ticker   *time.Ticker
	tokens   chan struct{}
	done     chan struct{}
	capacity int
}

func NewRateLimiter(rate int, duration time.Duration) *RateLimiter {
	interval := duration / time.Duration(rate)

	rl := &RateLimiter{
		ticker:   time.NewTicker(interval),
		tokens:   make(chan struct{}, rate),
		done:     make(chan struct{}),
		capacity: rate,
	}

	// 초기 토큰 채우기
	for i := 0; i < rate; i++ {
		rl.tokens <- struct{}{}
	}

	// 토큰 보충 시작
	go rl.replenish()

	return rl
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *RateLimiter) replenish() {
	for {
		select {
		case <-rl.ticker.C:
			select {
			case rl.tokens <- struct{}{}:
				// 토큰 추가 성공
			default:
				// 토큰 버킷이 가득 참
			}
		case <-rl.done:
			return
		}
	}
}

func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
	close(rl.done)
}

func (rl *RateLimiter) GetCapacity() int {
	return rl.capacity
}
