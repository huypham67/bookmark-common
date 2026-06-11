package common

import "github.com/rs/zerolog/log"

// ExitOnError exits the application with a fatal log message if the error is not nil.
func ExitOnError(err error, msg string) {
	if err != nil {
		log.Fatal().
			Err(err).
			Msg(msg)
	}
}
