package shared

import (
	"path/filepath"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/teadove/goteleout/internal/utils"
)

const (
	defaultEnvPrefix = "fun_"
	defaultEnvFile   = ".env"
)

type telegram struct {
	AppID           int    `env:"app_id,required"`
	AppHash         string `env:"app_hash,required"`
	PhoneNumber     string `env:"phone_number,required"`
	SessionPath     string `env:"session_storage_path"  envDefault:"telegram-session.json"`
	SessionFullPath string
}

type storage struct {
	Filename string `env:"filename" envDefault:"storage.json"`
}

type Settings struct {
	LogErrorToSelf  bool     `env:"log_error_to_self" envDefault:"false"`
	Telegram        telegram `envPrefix:"telegram__"`
	Storage         storage  `envPrefix:"storage__"`
	FileStoragePath string   `                       env:"file_storage_path" envDefault:"~/.config/fun-telegram/"`
	LogLevel        string   `                       env:"log_level"         envDefault:"debug"`
}

func MustNewSettings() Settings {
	var settings Settings
	_ = godotenv.Load(defaultEnvFile)

	err := env.Parse(&settings, env.Options{Prefix: defaultEnvPrefix})
	utils.Check(err)

	realPath, err := homedir.Expand(settings.FileStoragePath)
	utils.Check(err)
	settings.Telegram.SessionFullPath = filepath.Join(realPath, settings.Telegram.SessionPath)
	return settings
}
