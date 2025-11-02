// file: internal/adapters/input/udp_listener/listener.go
package udp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
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
	logger := log.FromCtx(ctx)

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

			var msg domain.LogMessage
			if err := json.Unmarshal(buffer[:n], &msg); err != nil {
				logger.Warn().Err(err).Msg("failed to unmarshal log")
				continue
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
