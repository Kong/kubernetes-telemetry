package forwarders

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bombsimon/logrusr/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/serializers"
	"github.com/kong/kubernetes-telemetry/pkg/telemetry"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

func TestTLSForwarder(t *testing.T) {
	const (
		rootPEM = `-----BEGIN CERTIFICATE-----
MIIDzDCCArSgAwIBAgIJAP5AVMhOiD+WMA0GCSqGSIb3DQEBCwUAMHExCzAJBgNV
BAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRYwFAYDVQQHDA1TYW4gRnJhbmNp
c2NvMQ0wCwYDVQQKDARLb25nMRgwFgYDVQQLDA9LdWJlcm5ldGVzIHRlYW0xDDAK
BgNVBAMMA0tHTzAeFw0yMjA4MjMxMzA3NTJaFw0zMjA4MjAxMzA3NTJaMHExCzAJ
BgNVBAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRYwFAYDVQQHDA1TYW4gRnJh
bmNpc2NvMQ0wCwYDVQQKDARLb25nMRgwFgYDVQQLDA9LdWJlcm5ldGVzIHRlYW0x
DDAKBgNVBAMMA0tHTzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMmb
QFXQoXIC56ZLMDL+qZAHfd92tBs0pxJv3ZKalIKInVy8YwnDmEINLkemsTlPgK32
NxI1UvW1nTo42CDp3GXE3dmKOxv5C1rWvR0EbnYyDK3idX4sS3nT3jqRJiSP0iJq
+r3LJjUg+LUVL+v0Yw41i1yhdEFyelTAKulykt9AVCDYA65j9cAhOjmFJjHTUR/J
/UveNARa/x9fu7rxfCBpMk+yPcXHBvJYm+GP4RF6izDRkJF+WBC9iPhri+feILk9
8W0VW77TC4tLBGUNBSy7JyRg6lK0VpCAZjPUG/mOnztn9E4YoVVm/qOoVLizGmOS
MGFAHoX68IOSLJkQHkkCAwEAAaNnMGUwCwYDVR0PBAQDAgQwMBMGA1UdJQQMMAoG
CCsGAQUFBwMBMEEGA1UdEQQ6MDiCCWxvY2FsaG9zdIIJMTI3LjAuMC4xgiB0ZXN0
LmdhdGV3YXktb3BlcmF0b3Iua29uZ2hxLmNvbTANBgkqhkiG9w0BAQsFAAOCAQEA
pseJcAOULR0NBziL9nvKjSSyhhUZCrj12QwXX/dFVUCu6RNGcvGTmv5tLh8W3CIs
vJMHKZbZ1+afRM+nHwMfKqqpkJU0TJaKBbf+Ky+Zrk9xu8TVO0cprB6RMUhQrdB9
10L6y6I0qseNMSWg/1wtcE3tiTFC7/Zc+ywK2wU8lwKfSgFBOxa+jgfMJQoYfCTs
Q5jdmdhJTzXrY8YscOFvqe6zH4fpH58HiG0v9Q5/YMb5kTwQsoAIR78TJN+6Klg8
kKklxFcttse5LppQ5mwfRFXFl3/UErOq3YED5cVgraCj0H2xe1kTxIPQR4ex/DgF
/zyvypUrraeZJTYrFthkgQ==
-----END CERTIFICATE-----`
		keyPEM = `-----BEGIN PRIVATE KEY-----
MIIEugIBADANBgkqhkiG9w0BAQEFAASCBKQwggSgAgEAAoIBAQDJm0BV0KFyAuem
SzAy/qmQB33fdrQbNKcSb92SmpSCiJ1cvGMJw5hCDS5HprE5T4Ct9jcSNVL1tZ06
ONgg6dxlxN3Zijsb+Qta1r0dBG52Mgyt4nV+LEt50946kSYkj9Iiavq9yyY1IPi1
FS/r9GMONYtcoXRBcnpUwCrpcpLfQFQg2AOuY/XAITo5hSYx01Efyf1L3jQEWv8f
X7u68XwgaTJPsj3FxwbyWJvhj+EReosw0ZCRflgQvYj4a4vn3iC5PfFtFVu+0wuL
SwRlDQUsuyckYOpStFaQgGYz1Bv5jp87Z/ROGKFVZv6jqFS4sxpjkjBhQB6F+vCD
kiyZEB5JAgMBAAECggEAS1F5A5Zh+lojePj2FNcXOfvShr2uI8vT7wtj1/VwLiQj
xhWLWoZ8R5DtDU+1Phf5lwQ5JtBNIgarqqi59fHoqQyXZUJDOvwbxeAb3s9dBUNF
gWDtTCn4OJdymqbHfTlN5BXbfzR6Hbcns18q/BfdOd2/Jugaqqi+ExOH9JcdT9Hq
2M1E4j+9OYQ4IreQsPzQzxmMAJnkDnDhXlG4Xv7RwCT8v1Uytkk965BD38oqp5p8
D2NrQZgY0nUK4oFiUEZV6kz1y4YfkNuEUoQ4porrk2m9nyCqbHVGINGVTlvmlfDC
qZt0Mt4+iDJwk+KKinKSJFrUYuCzZsq/bNiYYCPq0QKBgQDmPKtOGehjOneLIc/Z
ObovG15ftA2Ln1VGdNnoxAns0MUm+PI+w1H8Ox+8+yZAtBjKiQmUZ1zNdYsMG/LZ
T+O4HI3BudBMaAQbs+ve1LMZjwLAqejJQaPf6NxMSmi1D+wi2GWUR1VCTktXeMU3
VCb+QUQvTfd1tBgKkIlJXuQMNwKBgQDgKm+7++ew32Z/omWs5+WWfvAUQ9BxGdB7
fLGrJCptZz4HMezZ4D8LwLwo4oBvtLJkKiZPgfdxJqh/tb2HtR9GGesyqic1KG+c
AZuBoTXjb6i/YrCTBfFHNJacpt28oYnNyrVxe/fYP28KWRG47Es8ZBUd6l6NDfa9
uDsld7DpfwJ/Udc/DYQoFx2xYMOkHpNmm1gfM+XM6tS4e1MOIq+M16Fec3wKoETN
39skbQjZkCZ0qYoM3bPgSSh/RM6qhJThXZDI0xQ59u8ChtZuAceZ4nvzfojnNqMe
nXko1fWdQr9mMPy7Hvo8VFWAcpd7gy9mrPqGQkp0rGJYKWk3Y91XMwKBgBotOnEE
QJEJ9WkwKJlhVxEU76oeJSgf8JWLASBQD9hItxiV/ueOZS5VKmPH12G0AyTpOyIL
tj5zWjfXnDNNtkI0Yp+++OcfOrFICsW/cpCFiHoY5y+0APHktTXD0p7lajcq1bdT
16Rb+/aEYiprBXoe4cxlgvcLy2VqLxX3/SO3AoGAFyqu1k2B9hc+IWQYa03EWyx4
mM1T33W8gj2ADVwcLsfbK0inSk1jtGvd/ZksB7Hy9htBYOpU6a5HZpTZwm+Ek/T+
1VI1t2doPiYvTIo3Qy6JKeL6DmR73jZA41XtzrcQJ6Kfi5PFewqkIGe9Y6iSHRJX
f3cb9gYaLWdmvkx8p3g=
-----END PRIVATE KEY-----`
	)

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	require.True(t, ok)

	tlsCert, err := tls.X509KeyPair([]byte(rootPEM), []byte(keyPEM))
	require.NoError(t, err)

	listener, err := tls.Listen("tcp4", "localhost:0",
		&tls.Config{
			RootCAs:      roots,
			Certificates: []tls.Certificate{tlsCert},
			// Tried to actually verify the certificates in this test but was getting the following error
			// event though 127.0.0.1 and localhost were added to alt_names in openssl config:
			// failed to connect to reporting server: x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs
			InsecureSkipVerify: true, //nolint:gosec
		},
	)
	require.NoError(t, err)
	require.NotNil(t, listener)
	t.Logf("TLS server address %s", listener.Addr().String())

	var wg sync.WaitGroup
	go acceptLoop(t, listener, &wg,
		[]string{
			"<14>signal=test-signal;key=value;\n",
			"<14>signal=test-signal-2;key=value;\n",
		},
	)

	serializer := serializers.NewSemicolonDelimited()
	require.NotNil(t, serializer)

	log := logrusr.New(logrus.New())
	tf, err := NewTLSForwarder(listener.Addr().String(), log,
		func(c *tls.Config) {
			c.RootCAs = roots
			c.Certificates = []tls.Certificate{tlsCert}
			c.InsecureSkipVerify = true
		},
	)
	require.NoError(t, err)
	require.NotNil(t, tf)

	consumer := telemetry.NewConsumer(serializer, tf)

	m, err := telemetry.NewManager(
		"test-ping",
		telemetry.OptManagerPeriod(time.Hour),
		telemetry.OptManagerLogger(log),
	)
	require.NoError(t, err)

	w := telemetry.NewWorkflow("test1")
	p, err := provider.NewFixedValueProvider("test1-provider", types.ProviderReport{
		"key": "value",
	})
	require.NoError(t, err)
	w.AddProvider(p)
	m.AddWorkflow(w)
	require.NoError(t, m.AddConsumer(consumer))
	require.NoError(t, m.Start())

	wg.Add(1)
	require.NoError(t, m.TriggerExecute(context.Background(), "test-signal"))
	wg.Add(1)
	require.NoError(t, m.TriggerExecute(context.Background(), "test-signal-2"))
	wg.Wait()

	require.NoError(t, listener.Close())
	m.Stop()
}

func acceptLoop(t *testing.T, l net.Listener, wg *sync.WaitGroup, expectedData []string) {
	t.Log("server: accepting...")

	// Accept just once because TLSForwarder persists the connection
	conn, err := l.Accept()
	if err != nil {
		t.Logf("server: Accept() returned error: %v", err)
		return
	}
	handleClient(t, conn, expectedData, wg)
}

func handleClient(t *testing.T, conn net.Conn, expectedData []string, wg *sync.WaitGroup) {
	defer conn.Close() //nolint:gosec

	t.Logf("server: accepted from %s", conn.RemoteAddr())
	tlscon, ok := conn.(*tls.Conn)
	if ok {
		t.Log("server: connection ok")
		state := tlscon.ConnectionState()
		for _, v := range state.PeerCertificates {
			t.Log(x509.MarshalPKIXPublicKey(v.PublicKey))
		}
	}

	count := 0
	for ; count < len(expectedData); count++ {
		buf := make([]byte, 512)
		t.Log("server: conn: waiting")
		n, err := conn.Read(buf)
		if err != nil {
			t.Logf("server: conn: read: %s", err)
			break
		}

		assert.Equal(t, expectedData[count], string(buf[:n]))
		t.Logf("server: conn: echo %q\n", string(buf[:n]))

		n, err = conn.Write(buf[:n])
		t.Logf("server: conn: wrote %d bytes", n)
		if err != nil {
			t.Logf("server: write: %s", err)
			break
		}

		wg.Done()
	}

	assert.Equal(t, len(expectedData), count)
	t.Log("server: conn: closed")
}
