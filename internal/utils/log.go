package utils

import "github.com/rs/zerolog/log"

func LogInterface(v any) {
	log.Info().Interface(GetType(v), v).Str("status", "logging.interface").Send()
}
