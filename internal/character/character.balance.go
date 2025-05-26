package character

import (
	"sync"
	"time"
)

// BalanceType represents different types of character balance
type BalanceType string

const (
	AttackBalance   BalanceType = "attack"
	HealingBalance  BalanceType = "healing"
	MovementBalance BalanceType = "movement"
)

// Balance tracks cooldowns for different action types
type Balance struct {
	mu        sync.RWMutex
	balances  map[BalanceType]time.Time
	cooldowns map[BalanceType]time.Duration
}

func NewBalance() *Balance {
	return &Balance{
		balances:  make(map[BalanceType]time.Time),
		cooldowns: make(map[BalanceType]time.Duration),
	}
}

func (b *Balance) SetCooldown(balanceType BalanceType, duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cooldowns[balanceType] = duration
}

func (b *Balance) UseBalance(balanceType BalanceType, customDuration ...time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	duration := b.cooldowns[balanceType]
	if len(customDuration) > 0 {
		duration = customDuration[0]
	}

	b.balances[balanceType] = time.Now().Add(duration)
}

func (b *Balance) HasBalance(balanceType BalanceType) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	cooldownEnd, exists := b.balances[balanceType]
	if !exists {
		return true
	}

	return time.Now().After(cooldownEnd)
}

// TimeUntilBalance returns the time remaining until a balance is restored
func (b *Balance) TimeUntilBalance(balanceType BalanceType) time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()

	cooldownEnd, exists := b.balances[balanceType]
	if !exists {
		return 0
	}

	remaining := time.Until(cooldownEnd)
	if remaining < 0 {
		return 0
	}
	return remaining
}
