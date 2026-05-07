package auditlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const dirEnvVar = "MCP_GITLAB_API_LOG_DIR"

type entry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Payload   any    `json:"payload"`
}

func Write(eventType string, payload any) (string, error) {
	id := uuid.NewString()
	logDir := stringsOrDefault(os.Getenv(dirEnvVar), "logs")

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(entry{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      eventType,
		Payload:   payload,
	}, "", "  ")
	if err != nil {
		return "", err
	}

	path := filepath.Join(logDir, id+".json")
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return "", err
	}

	return path, nil
}

func stringsOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
