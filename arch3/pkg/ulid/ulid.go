package ulid

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// New 生成一个新的 ULID
// 使用 crypto/rand 作为熵源，保证密码学安全
// 使用 Monotonic 保证同一毫秒内 ID 单调递增
func New() (string, error) {
	entropy := ulid.Monotonic(rand.Reader, 0)
	timestamp := ulid.Timestamp(time.Now())

	id, err := ulid.New(timestamp, entropy)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
