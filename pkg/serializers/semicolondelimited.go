package serializers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

type semicolonDelimited struct {
	signal string
}

// NewSemicolonDelimited creates a new serializer that will serialize telemetry
// reports into a semicolon delimited format.
func NewSemicolonDelimited(signal string) semicolonDelimited {
	return semicolonDelimited{
		signal: signal,
	}
}

func (s semicolonDelimited) Serialize(report types.Report) ([]byte, error) {
	out := make([]string, 0, len(report))
	for _, v := range report {
		out = append(out, serializeReport(v))
	}

	// Should this prefix go to TLSForwarder instead?
	prefix := "<14>signal=" + s.signal + ";"

	sort.Strings(out)
	return []byte(prefix + strings.Join(out, "") + "\n"), nil
}

func serializeReport(report provider.Report) string {
	var out []string
	for k, v := range report {
		switch vv := v.(type) {
		case provider.Report:
			out = append(out, serializeReport(vv))
		default:
			out = append(out, fmt.Sprintf("%v=%v;", k, v))
		}
	}

	sort.Strings(out)
	return strings.Join(out, "")
}
