package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type LogMessage struct {
	Time    time.Time
	Level   string
	App     string
	Service string
	Message string
	Fields  map[string]interface{}
}

func (m *LogMessage) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if t, ok := raw["time"]; ok {
		if timeStr, ok := t.(string); ok {
			parsedTime, err := time.Parse(time.RFC3339Nano, timeStr)
			if err != nil {
				parsedTime, _ = time.Parse(time.RFC3339, timeStr)
			}
			m.Time = parsedTime
		}
		delete(raw, "time")
	}

	if lvl, ok := raw["level"]; ok {
		m.Level = fmt.Sprintf("%v", lvl)
		delete(raw, "level")
	}

	if app, ok := raw["app"]; ok {
		m.App = fmt.Sprintf("%v", app)
		delete(raw, "app")
	}

	if srv, ok := raw["service"]; ok {
		m.Service = fmt.Sprintf("%v", srv)
		delete(raw, "service")
	}

	if msg, ok := raw["message"]; ok {
		m.Message = fmt.Sprintf("%v", msg)
		delete(raw, "message")
	}

	m.Fields = raw

	return nil
}
