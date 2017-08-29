package pdp

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

func TestContext(t *testing.T) {
	ct := strtree.NewTree()
	ct.InplaceInsert("test-notag-content", &LocalContent{id: "test-notag-content"})
	var nullContent *LocalContent
	ct.InplaceInsert("test-null-content", nullContent)
	tag := uuid.New()
	ct.InplaceInsert("test-content", &LocalContent{id: "test-content", tag: &tag})
	lcs := &LocalContentStorage{r: ct}

	ctx, err := NewContext(lcs, 9, func(i int) (string, AttributeValue, error) {
		switch i {
		default:
			return "", undefinedValue, fmt.Errorf("No attribute for index: %d", i)

		case 0:
			return "b", MakeBooleanValue(true), nil

		case 1:
			return "s", MakeStringValue("test"), nil

		case 2:
			a := net.ParseIP("192.0.2.1")
			if a == nil {
				return "", undefinedValue, fmt.Errorf("Can't make IP address")
			}

			return "a", MakeAddressValue(a), nil

		case 3:
			_, n, err := net.ParseCIDR("192.0.2.0/24")
			if err != nil {
				return "", undefinedValue, fmt.Errorf("Can't make IP network: %s", err)
			}

			return "n", MakeNetworkValue(n), nil

		case 4:
			return "d", MakeDomainValue("example.com"), nil

		case 5:
			st := strtree.NewTree()
			st.InplaceInsert("1 - one", 1)
			st.InplaceInsert("2 - two", 2)
			st.InplaceInsert("3 - three", 3)
			return "ss", MakeSetOfStringsValue(st), nil

		case 6:
			nt := iptree.NewTree()
			_, n, err := net.ParseCIDR("192.0.2.0/28")
			if err != nil {
				return "", undefinedValue, fmt.Errorf("Can't make IP network: %s", err)
			}
			nt.InplaceInsertNet(n, 1)
			_, n, err = net.ParseCIDR("192.0.2.16/28")
			if err != nil {
				return "", undefinedValue, fmt.Errorf("Can't make IP network: %s", err)
			}
			nt.InplaceInsertNet(n, 2)
			_, n, err = net.ParseCIDR("192.0.2.32/28")
			if err != nil {
				return "", undefinedValue, fmt.Errorf("Can't make IP network: %s", err)
			}
			nt.InplaceInsertNet(n, 3)
			return "sn", MakeSetOfNetworksValue(nt), nil

		case 7:
			dt := &domaintree.Node{}
			dt.InplaceInsert("example.com", 1)
			dt.InplaceInsert("example.org", 2)
			dt.InplaceInsert("example.net", 3)
			return "sd", MakeSetOfDomainsValue(dt), nil

		case 8:
			return "ls", MakeListOfStringsValue([]string{"1 - one", "2 - two", "3 - three"}), nil
		}
	})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if len(ctx.String()) <= 0 {
		t.Error("Expected content description but got nothing")
	}
}
