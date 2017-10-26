package test

import (
	"fmt"
	"io"
	"os"
	"strings"

	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
	"github.com/infobloxopen/themis/pepcli/requests"
)

const (
	Name        = "test"
	Description = "evaluates given requests on PDP server"
)

func Exec(addr, in, out string, n int, v interface{}) error {
	reqs, err := requests.Load(in)
	if err != nil {
		return fmt.Errorf("can't load requests from \"%s\"", in)
	}

	if n < 1 {
		n = len(reqs)
	}

	f := os.Stdout
	if len(out) > 0 {
		f, err = os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	c := pep.NewBalancedClient([]string{addr}, 0, 0)
	err = c.Connect()
	if err != nil {
		return fmt.Errorf("can't connect to %s: %s", addr, err)
	}
	defer c.Close()

	for i := 0; i < n; i++ {
		idx := i % len(reqs)
		req := reqs[idx]

		res, err := c.Validate(&req)
		if err != nil {
			return fmt.Errorf("can't send request %d (%d): %s", idx, i, err)
		}

		err = dump(res, f)
		if err != nil {
			return fmt.Errorf("can't dump response for reqiest %d (%d): %s", idx, i, err)
		}
	}

	return nil
}

func dump(r *pdp.Response, f io.Writer) error {
	lines := []string{fmt.Sprintf("- effect: %s", pdp.EffectName(r.Effect))}
	if len(r.Reason) > 0 {
		lines = append(lines, fmt.Sprintf("  reason: %q", r.Reason))
	}

	if len(r.Obligations) > 0 {
		lines = append(lines, "  obligation:")
		for _, attr := range r.Obligations {
			lines = append(lines, fmt.Sprintf("    - id: %q", attr.Id))
			lines = append(lines, fmt.Sprintf("      type: %q", attr.Type))
			lines = append(lines, fmt.Sprintf("      value: %q", attr.Value))
			lines = append(lines, "")
		}
	} else {
		lines = append(lines, "")
	}

	_, err := fmt.Fprintf(f, "%s\n", strings.Join(lines, "\n"))
	return err
}
