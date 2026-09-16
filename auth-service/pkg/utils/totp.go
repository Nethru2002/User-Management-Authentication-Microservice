package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

type TOTPManager struct {
	Issuer string
	Period int64
	Digits int
}

func NewTOTPManager(issuer string) *TOTPManager {
	return &TOTPManager{
		Issuer: issuer,
		Period: 30,
		Digits: 6,
	}
}

func (t *TOTPManager) GenerateSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

func (t *TOTPManager) GenerateCode(secret string, timestamp time.Time) (string, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", err
	}

	counter := uint64(timestamp.Unix() / t.Period)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	truncatedHash := binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff
	code := truncatedHash % uint32(math.Pow10(t.Digits))

	return fmt.Sprintf(fmt.Sprintf("%%0%dd", t.Digits), code), nil
}

// ValidateCode checks against current, previous, and next intervals to tolerate clock drift
func (t *TOTPManager) ValidateCode(secret, code string) bool {
	now := time.Now()
	for _, offset := range []int64{-1, 0, 1} {
		testTime := now.Add(time.Duration(offset*t.Period) * time.Second)
		expected, err := t.GenerateCode(secret, testTime)
		if err == nil && hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}

func (t *TOTPManager) ProvisioningURI(email, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		t.Issuer, email, secret, t.Issuer,
	)
}