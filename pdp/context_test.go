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
			return "", UndefinedValue, fmt.Errorf("no attribute for index: %d", i)

		case 0:
			return "b", MakeBooleanValue(true), nil

		case 1:
			return "s", MakeStringValue("test"), nil

		case 2:
			return "a", MakeAddressValue(net.ParseIP("192.0.2.1")), nil

		case 3:
			return "n", MakeNetworkValue(makeTestNetwork("192.0.2.0/24")), nil

		case 4:
			return "d", MakeDomainValue(makeTestDomain("example.com")), nil

		case 5:
			st := strtree.NewTree()
			st.InplaceInsert("1 - one", 1)
			st.InplaceInsert("2 - two", 2)
			st.InplaceInsert("3 - three", 3)
			return "ss", MakeSetOfStringsValue(st), nil

		case 6:
			nt := iptree.NewTree()
			nt.InplaceInsertNet(makeTestNetwork("192.0.2.0/28"), 1)
			nt.InplaceInsertNet(makeTestNetwork("192.0.2.16/28"), 2)
			nt.InplaceInsertNet(makeTestNetwork("192.0.2.32/28"), 3)
			return "sn", MakeSetOfNetworksValue(nt), nil

		case 7:
			dt := &domaintree.Node{}
			dt.InplaceInsert(makeTestDomain("example.com"), 1)
			dt.InplaceInsert(makeTestDomain("example.org"), 2)
			dt.InplaceInsert(makeTestDomain("example.net"), 3)
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
