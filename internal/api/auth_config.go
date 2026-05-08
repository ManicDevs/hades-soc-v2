package api

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	authConfigOnce sync.Once
	authConfigErr  error
	jwtSecret      []byte
	devPasswords   map[string]string
)

func initAuthConfig() error {
	authConfigOnce.Do(func() {
		secret := strings.TrimSpace(os.Getenv("HADES_JWT_SECRET"))
		if secret == "" {
			authConfigErr = fmt.Errorf("HADES_JWT_SECRET is required")
			return
		}
		jwtSecret = []byte(secret)

		devPasswords = map[string]string{}
		raw := strings.TrimSpace(os.Getenv("HADES_DEV_CREDENTIALS"))
		if raw == "" {
			return
		}
		for _, pair := range strings.Split(raw, ",") {
			parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				continue
			}
			devPasswords[parts[0]] = parts[1]
		}
	})

	return authConfigErr
}
