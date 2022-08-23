package forwarders

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
)

const (
	defaultTimeout  = time.Second * 30
	defaultDeadline = time.Minute
)

var defaultTLSConf = tls.Config{
	MinVersion: tls.VersionTLS13,
	MaxVersion: tls.VersionTLS13,
}

type tlsForwarder struct {
	logger logr.Logger
	conn   *tls.Conn
}

// TLSOpt defines an option type that manipulates *tls.Config.
type TLSOpt func(*tls.Config)

// NewTLSForwarder creates a TLS forwarder which forwards received serialized reports
// to a TLS endpoint specified by the provided address.
func NewTLSForwarder(address string, logger logr.Logger, tlsOpts ...TLSOpt) (*tlsForwarder, error) {
	tlsConf := defaultTLSConf.Clone()
	for _, opt := range tlsOpts {
		opt(tlsConf)
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: defaultTimeout,
		},
		"tcp",
		address,
		tlsConf,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to reporting server: %w", err)
	}

	return &tlsForwarder{
		logger: logger,
		conn:   conn,
	}, nil
}

// Name returns the name of the forwarder.
func (tf *tlsForwarder) Name() string {
	return "TLSForwarder"
}

// Forward forwards the received payload to the configured TLS endpoint.
func (tf *tlsForwarder) Forward(ctx context.Context, payload []byte) error {
	var deadline time.Time
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	} else {
		deadline = time.Now().Add(defaultDeadline)
	}

	err := tf.conn.SetDeadline(deadline)
	if err != nil {
		return fmt.Errorf("failed to set report connection deadline: %w", err)
	}

	_, err = tf.conn.Write(payload)
	if err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}
	return nil
}
