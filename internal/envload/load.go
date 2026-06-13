package envload

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var once sync.Once

// Load reads optional .env files. Existing environment variables are not overwritten.
func Load() {
	once.Do(func() {
		if dir, err := os.UserConfigDir(); err == nil {
			_ = godotenv.Load(filepath.Join(dir, "theirtime", ".env"))
		}
		_ = godotenv.Load(".env")
	})
}
