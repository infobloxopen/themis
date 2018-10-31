package integrationTests

import (
	"fmt"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pdp/ast"
)

type policyFormat int

const (
	YAML policyFormat = iota
	JSON
)

var policyFormatString = map[policyFormat]string{
	YAML: "YAML",
	JSON: "JSON",
}

func (f policyFormat) String() string {
	return policyFormatString[f]
}

type testCase struct {
	attrs              []pdp.AttributeAssignment
	expected           int
	expectedObligation string
	expectedError      string
}

type testSuite struct {
	policies map[policyFormat]string
	testSet  []testCase
}

func init() {
	log.SetLevel(log.ErrorLevel)
}

func loadPolicy(pf policyFormat, ps string) *pdp.PolicyStorage {
	var parser ast.Parser
	switch pf {
	case YAML:
		parser = ast.NewYAMLParser()
	case JSON:
		parser = ast.NewJSONParser()
	}
	policyStorage, err := parser.Unmarshal(strings.NewReader(ps), nil)
	if err != nil {
		panic(fmt.Errorf("expected no error while parsing policies but got: %s", err))
	}

	return policyStorage
}

func createContext(req []pdp.AttributeAssignment) (*pdp.Context, error) {
	ctx, err := pdp.NewContext(nil, len(req), func(i int) (string, pdp.AttributeValue, error) {
		a := req[i]

		v, err := a.GetValue()
		if err != nil {
			return "", pdp.UndefinedValue, fmt.Errorf("error getting attribute value: %s", err)
		}

		return a.GetID(), v, nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create context: %s", err)
	}

	return ctx, nil
}

func serializeAssignments(attrs []pdp.AttributeAssignment) (string, error) {
	s := make([]string, len(attrs))

	ctx, err := pdp.NewContext(nil, 0, nil)
	if err != nil {
		return "", fmt.Errorf("can't create empty context")
	}

	for i, a := range attrs {
		id, t, v, err := a.Serialize(ctx)
		if err != nil {
			return "", fmt.Errorf("can't serialize attribute %q (%d): %s", id, i, err)
		}

		s[i] = fmt.Sprintf("%s.(%s)=%q", id, t, v)
	}

	return strings.Join(s, ","), nil
}

func validateTestSuite(ts testSuite, t *testing.T) {
	for i, tc := range ts.testSet {
		desc, err := serializeAssignments(tc.attrs)
		if err != nil {
			t.Fatalf("can't create descripiton for case %d: %s", i, err)
		}

		t.Run(desc, func(t *testing.T) {
			ctx, err := createContext(tc.attrs)
			if err != nil {
				t.Fatalf("Expected no error while creating context but got: %s", err)
			}

			for pf, ps := range ts.policies {
				t.Run(fmt.Sprintf("Policy Format: %s", pf), func(t *testing.T) {
					p := loadPolicy(pf, ps)
					r := p.Root().Calculate(ctx)
					err := r.Status
					if err != nil {
						if tc.expectedError == "" {
							t.Fatalf("Expected no error while evaluating policy but got: %s", err)
						} else if !strings.Contains(err.Error(), tc.expectedError) {
							t.Fatalf("Expected error while evaluating policy '%s', but got '%s'", tc.expectedError, err)
						}
					}

					if r.Effect != tc.expected {
						t.Fatalf("Expected result of policy evaluation %s, but got %s",
							pdp.EffectNameFromEnum(tc.expected), pdp.EffectNameFromEnum(r.Effect))
					}
					if tc.expectedObligation != "" {
						obLen := len(r.Obligations)
						if obLen != 1 {
							t.Fatalf("Expected result of policy evaluation include 1 obligation, but got %d", obLen)
						}
						_, _, obligationRes, err := r.Obligations[0].Serialize(ctx)
						if err != nil {
							t.Fatalf("Expected when serializing obligation, but got %s", err)
						}
						if obligationRes != tc.expectedObligation {
							t.Fatalf("Expected obligation of policy evaluation %s, but got %s", tc.expectedObligation, obligationRes)
						}
					}
				})
			}
		})
	}
}
