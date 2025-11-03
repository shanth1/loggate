package generator

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/shanth1/loggate/cmd/loggen/internal/config"
	"github.com/shanth1/loggate/internal/core/domain"
)

type Generator struct {
	template *config.TemplateConfig
	faker    *gofakeit.Faker
}

func New(template *config.TemplateConfig) *Generator {
	return &Generator{
		template: template,
		faker:    gofakeit.New(0),
	}
}

func (g *Generator) Generate() domain.LogMessage {
	level := g.pickLevel()
	messageTemplate := g.pickMessage(level)
	message := g.faker.Generate(messageTemplate)

	return domain.LogMessage{
		Time:    time.Now(),
		Level:   level,
		App:     g.template.App,
		Service: g.template.Service,
		Message: message,
		Fields:  g.generateFields(),
	}
}

func (g *Generator) pickLevel() string {
	r := rand.Float64()
	var cumulative float64
	for level, prob := range g.template.Levels {
		cumulative += prob
		if r < cumulative {
			return level
		}
	}
	return "INFO" // Fallback
}

func (g *Generator) pickMessage(level string) string {
	messages, ok := g.template.Messages[level]
	if !ok || len(messages) == 0 {
		return "Default message for level " + level
	}
	return messages[rand.Intn(len(messages))]
}

func (g *Generator) generateFields() map[string]interface{} {
	if len(g.template.Fields) == 0 {
		return nil
	}
	fields := make(map[string]interface{}, len(g.template.Fields))
	for _, fieldCfg := range g.template.Fields {
		parts := strings.Split(fieldCfg.Type, ":")
		fakeFunc := parts[0]

		var val interface{}
		switch fakeFunc {
		case "uuid":
			val = g.faker.UUID()
		case "useragent":
			val = g.faker.UserAgent()
		case "price", "float":
			if len(parts) == 2 {
				params := strings.Split(parts[1], ",")
				var min, max float64
				fmt.Sscanf(params[0], "%f", &min)
				fmt.Sscanf(params[1], "%f", &max)
				val = g.faker.Price(min, max)
			} else {
				val = g.faker.Price(0, 1000)
			}
		case "number":
			if len(parts) == 2 {
				params := strings.Split(parts[1], ",")
				var min, max int
				fmt.Sscanf(params[0], "%d", &min)
				fmt.Sscanf(params[1], "%d", &max)
				val = g.faker.Number(min, max)
			} else {
				val = g.faker.Number(0, 10000)
			}
		case "productsku":
			val = fmt.Sprintf("SKU-%d", g.faker.Number(1000, 9999))
		default:
			val = g.faker.Word()
		}
		fields[fieldCfg.Key] = val
	}
	return fields
}
