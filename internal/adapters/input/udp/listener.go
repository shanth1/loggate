package udp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/shanth1/gotools/log"
	"github.com/shanth1/loggate/internal/core/domain"
	"github.com/shanth1/loggate/internal/core/ports"
)

type Listener struct {
	conn     *net.UDPConn
	ingester ports.LogIngester
}

func New(address string, ingester ports.LogIngester) (*Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return &Listener{conn: conn, ingester: ingester}, nil
}

func (l *Listener) Start(ctx context.Context) {
	logger := log.FromContext(ctx)

	logger.Info().Msg(fmt.Sprintf("udp listener started on %s", l.conn.LocalAddr()))

	defer l.conn.Close()

	buffer := make([]byte, 65535)

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("shutting down UDP listener...")
			return
		default:
			l.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := l.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				logger.Error().Err(err).Msg("reading from udp")
				continue
			}

			var rawData map[string]interface{}
			if err := json.Unmarshal(buffer[:n], &rawData); err != nil {
				logger.Warn().Err(err).Str("raw_log", string(buffer[:n])).Msg("failed to unmarshal raw log data")
				continue
			}

			msg := domain.LogMessage{
				Fields: make(map[string]interface{}),
			}

			if t, ok := rawData["time"]; ok {
				if timeStr, ok := t.(string); ok {
					parsedTime, err := time.Parse(time.RFC3339Nano, timeStr)
					if err != nil {
						parsedTime, _ = time.Parse(time.RFC3339, timeStr)
					}
					msg.Time = parsedTime
				}
				delete(rawData, "time")
			}
			if msg.Time.IsZero() {
				msg.Time = time.Now()
			}

			if lvl, ok := rawData["level"]; ok {
				msg.Level = fmt.Sprintf("%v", lvl)
				delete(rawData, "level")
			}
			msg.Level = strings.ToLower(msg.Level)

			if app, ok := rawData["app"]; ok {
				msg.App = fmt.Sprintf("%v", app)
				delete(rawData, "app")
			}

			if srv, ok := rawData["service"]; ok {
				msg.Service = fmt.Sprintf("%v", srv)
				delete(rawData, "service")
			}

			if m, ok := rawData["message"]; ok {
				msg.Message = fmt.Sprintf("%v", m)
				delete(rawData, "message")
			}

			for key, value := range rawData {
				msg.Fields[key] = value
			}

			if msg.App == "" {
				logger.Warn().Str("source", addr.String()).Str("log_content", string(buffer[:n])).Msg("received log with missing 'app' field, dropping")
				continue
			}
			if msg.Service == "" {
				logger.Warn().Str("source", addr.String()).Str("log_content", string(buffer[:n])).Msg("received log with missing 'service' field, dropping")
				continue
			}

			l.ingester.Ingest(ctx, msg)
		}
	}
}
