package domain

import "time"

type LogMessage struct {
	Time    time.Time              `json:"time"`
	Level   string                 `json:"level"`
	App     string                 `json:"app"`
	Service string                 `json:"service"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}
