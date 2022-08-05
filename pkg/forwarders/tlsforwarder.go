package forwarders

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout  = time.Second * 30
	defaultDeadline = time.Minute
)

var tlsConf = tls.Config{
	MinVersion: tls.VersionTLS13,
	MaxVersion: tls.VersionTLS13,
}

// TODO: Address logging levels and library to be used.
// See: https://github.com/Kong/kubernetes-ingress-controller/issues/1893
const (
	logrusrDiff = 4

	// InfoLevel is the converted logging level from logrus to go-logr for
	// information level logging. Note that the logrusr middleware technically
	// flattens all levels prior to this level into this level as well.
	InfoLevel = int(logrus.InfoLevel) - logrusrDiff

	// DebugLevel is the converted logging level from logrus to go-logr for
	// debug level logging.
	DebugLevel = int(logrus.DebugLevel) - logrusrDiff

	// WarnLevel is the converted logrus level to go-logr for warnings.
	WarnLevel = int(logrus.WarnLevel) - logrusrDiff
)

type tlsForwarder struct {
	log  logr.Logger
	conn *tls.Conn
}

// NewTLSForwarder creates a TLS forwarder which forwards received serialized reports
// to a TLS endpoint specified by the provided address.
func NewTLSForwarder(address string, log logr.Logger) *tlsForwarder {
	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: defaultTimeout,
		},
		"tcp",
		address,
		&tlsConf,
	)
	if err != nil {
		log.V(DebugLevel).Info("failed to connect to reporting server", "error", err)
		return nil
	}

	return &tlsForwarder{
		log:  log,
		conn: conn,
	}
}

func (tf *tlsForwarder) Name() string {
	return "TLSForwarder"
}

func (tf *tlsForwarder) Forward(payload []byte) error {
	err := tf.conn.SetDeadline(time.Now().Add(defaultDeadline))
	if err != nil {
		return fmt.Errorf("failed to set report connection deadline: %w", err)
	}

	_, err = tf.conn.Write(payload)
	if err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}
	return nil
}
