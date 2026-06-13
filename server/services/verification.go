package services

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type verificationEntry struct {
	code      string
	expiresAt time.Time
}

type VerificationService struct {
	mu    sync.RWMutex
	codes map[string]*verificationEntry // key: "email:purpose"
	times map[string]time.Time          // key: "email:purpose" — 上次发送时间
}

func NewVerificationService() *VerificationService {
	v := &VerificationService{
		codes: make(map[string]*verificationEntry),
		times: make(map[string]time.Time),
	}
	// 定期清理过期验证码
	go v.cleanup()
	return v
}

// GenerateCode 生成 6 位随机验证码
func (v *VerificationService) GenerateCode(email, purpose string) string {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	key := email + ":" + purpose

	v.mu.Lock()
	v.codes[key] = &verificationEntry{
		code:      code,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	v.times[key] = time.Now()
	v.mu.Unlock()

	return code
}

// CheckRateLimit 检查是否在 60 秒限频内
func (v *VerificationService) CheckRateLimit(email, purpose string) bool {
	key := email + ":" + purpose
	v.mu.RLock()
	last, exists := v.times[key]
	v.mu.RUnlock()

	if exists && time.Since(last) < 60*time.Second {
		return false // 被限频
	}
	return true
}

// VerifyCode 验证验证码（不消耗，仅校验）
func (v *VerificationService) VerifyCode(email, code, purpose string) error {
	key := email + ":" + purpose
	v.mu.RLock()
	entry, exists := v.codes[key]
	v.mu.RUnlock()

	if !exists {
		return fmt.Errorf("验证码不存在或已过期")
	}
	if time.Now().After(entry.expiresAt) {
		return fmt.Errorf("验证码已过期")
	}
	if entry.code != code {
		return fmt.Errorf("验证码错误")
	}
	return nil
}

// UseCode 验证并消耗验证码（验证通过后删除）
func (v *VerificationService) UseCode(email, code, purpose string) error {
	key := email + ":" + purpose
	v.mu.Lock()
	defer v.mu.Unlock()

	entry, exists := v.codes[key]
	if !exists {
		return fmt.Errorf("验证码不存在或已过期")
	}
	if time.Now().After(entry.expiresAt) {
		delete(v.codes, key)
		return fmt.Errorf("验证码已过期")
	}
	if entry.code != code {
		return fmt.Errorf("验证码错误")
	}

	delete(v.codes, key)
	return nil
}

// cleanup 定期清理过期验证码
func (v *VerificationService) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		v.mu.Lock()
		now := time.Now()
		for key, entry := range v.codes {
			if now.After(entry.expiresAt) {
				delete(v.codes, key)
			}
		}
		v.mu.Unlock()
	}
}
