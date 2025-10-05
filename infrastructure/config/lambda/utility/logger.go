package utility

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// Channel enum
type Priority uint

const (
	INFO Priority = iota
	WARN
	ERROR
	DEBUG
)

func (priority Priority) Log(msg string, key string, err any, isProdLog bool) {
	if !strings.Contains(strings.ToLower(LoadENV("ENV")), "prod") || isProdLog {
		switch priority {
		case INFO:
			log.Printf("ℹ️ INFO: %s %s:%s", msg, key, stringifyOneLine(err))
		case WARN:
			log.Printf("⚠️ WARN: %s %s:%s", msg, key, stringifyOneLine(err))
		case ERROR:
			log.Printf("❌ ERROR: %s %s:%s", msg, key, stringifyOneLine(err))
		case DEBUG:
			log.Printf("🔍 DEBUG: %s %s:%s", msg, key, stringifyOneLine(err))
		default:
			log.Println("❌ ERROR: %s %s:%s", msg, key, stringifyOneLine(err))
		}
	}
}

func stringifyOneLine(v any) string {
	if v == nil {
		return "<nil>"
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(bytes)
}
