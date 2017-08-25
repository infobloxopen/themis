package pdp

import (
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

func TestLocalContentStorage(t *testing.T) {
	tag := uuid.New()

	sTree := strtree.NewTree()
	sm := MakeContentStringMap(sTree)
	ssmc := MakeContentMappingItem("str-str-map", TypeString, []int{TypeString, TypeString}, sm)

	sTree = strtree.NewTree()
	sTree.InplaceInsert("1-first", "first")
	sTree.InplaceInsert("2-second", "second")
	sTree.InplaceInsert("3-third", "third")

	sm = MakeContentStringMap(sTree)
	ksm := MakeContentMappingItem("key-str-map", TypeString, []int{TypeString}, sm)

	sTree = strtree.NewTree()
	sm = MakeContentStringMap(sTree)
	snmc := MakeContentMappingItem("str-net-map", TypeString, []int{TypeString, TypeNetwork}, sm)

	nTree := iptree.NewTree()
	_, n, err := net.ParseCIDR("192.0.2.16/28")
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}
	nTree.InplaceInsertNet(n, "first")
	_, n, err = net.ParseCIDR("192.0.2.32/28")
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}
	nTree.InplaceInsertNet(n, "second")
	_, n, err = net.ParseCIDR("2001:db8::/32")
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}
	nTree.InplaceInsertNet(n, "third")

	nm := MakeContentNetworkMap(nTree)
	knm := MakeContentMappingItem("key-net-map", TypeString, []int{TypeNetwork}, nm)

	sTree = strtree.NewTree()
	sm = MakeContentStringMap(sTree)
	sdmc := MakeContentMappingItem("str-dom-map", TypeString, []int{TypeString, TypeDomain}, sm)

	dTree := &domaintree.Node{}
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dm := MakeContentDomainMap(dTree)
	kdm := MakeContentMappingItem("key-dom-map", TypeString, []int{TypeDomain}, dm)

	sTree = strtree.NewTree()
	sTree.InplaceInsert("1-first", "first")
	sTree.InplaceInsert("2-second", "second")
	sTree.InplaceInsert("3-third", "third")

	sm = MakeContentStringMap(sTree)
	smc := MakeContentMappingItem("str-map", TypeString, []int{TypeString}, sm)

	nTree = iptree.NewTree()
	_, n, err = net.ParseCIDR("192.0.2.16/28")
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}
	nTree.InplaceInsertNet(n, "first")
	_, n, err = net.ParseCIDR("192.0.2.32/28")
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}
	nTree.InplaceInsertNet(n, "second")
	_, n, err = net.ParseCIDR("2001:db8::/32")
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}
	nTree.InplaceInsertNet(n, "third")

	nm = MakeContentNetworkMap(nTree)
	nmc := MakeContentMappingItem("net-map", TypeString, []int{TypeNetwork}, nm)

	dTree = &domaintree.Node{}
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dm = MakeContentDomainMap(dTree)
	dmc := MakeContentMappingItem("dom-map", TypeString, []int{TypeDomain}, dm)

	dTree = &domaintree.Node{}
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dm = MakeContentDomainMap(dTree)
	dmcAdd := MakeContentMappingItem("dom-map-add", TypeString, []int{TypeDomain}, dm)

	dTree = &domaintree.Node{}
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dm = MakeContentDomainMap(dTree)
	dmcDel := MakeContentMappingItem("dom-map-del", TypeString, []int{TypeDomain}, dm)

	items := []*ContentItem{ssmc, snmc, sdmc, smc, nmc, dmc, dmcDel}
	s := NewLocalContentStorage([]*LocalContent{NewLocalContent("first", &tag, items)})

	dTree = &domaintree.Node{}
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dm = MakeContentDomainMap(dTree)
	dmc = MakeContentMappingItem("dom-map", TypeString, []int{TypeDomain}, dm)
	s = s.Add(NewLocalContent("second", &tag, []*ContentItem{dmc}))

	newTag := uuid.New()
	u := NewContentUpdate("first", tag, newTag)
	u.Append(UOAdd, []string{"str-str-map", "key"}, ksm)
	u.Append(UOAdd, []string{"str-net-map", "key"}, knm)
	u.Append(UOAdd, []string{"str-dom-map", "key"}, kdm)

	u.Append(UOAdd, []string{"str-str-map", "key", "4-fourth"}, MakeContentValueItem("4-fourth", TypeString, "fourth"))
	u.Append(UODelete, []string{"str-str-map", "key", "3-third"}, nil)

	u.Append(UOAdd, []string{"str-net-map", "key", "192.0.2.48/28"},
		MakeContentValueItem("192.0.2.48/28", TypeString, "fourth"))
	u.Append(UODelete, []string{"str-net-map", "key", "2001:db8::/32"}, nil)

	u.Append(UOAdd, []string{"str-dom-map", "key", "example.gov"},
		MakeContentValueItem("example.gov", TypeString, "fourth"))
	u.Append(UODelete, []string{"str-dom-map", "key", "example.net"}, nil)

	u.Append(UOAdd, []string{"str-map", "4-fourth"}, MakeContentValueItem("4-fourth", TypeString, "fourth"))
	u.Append(UOAdd, []string{"net-map", "2001:db8:1000::/40"},
		MakeContentValueItem("2001:db8:1000::/40", TypeString, "fourth"))
	u.Append(UOAdd, []string{"net-map", "2001:db8:1000::1"},
		MakeContentValueItem("2001:db8:1000::1", TypeString, "fifth"))
	u.Append(UODelete, []string{"str-map", "3-third"}, nil)

	u.Append(UOAdd, []string{"dom-map-add"}, dmcAdd)
	u.Append(UODelete, []string{"dom-map-del"}, nil)

	tr, err := s.NewTransaction("first", &tag)
	if err != nil {
		t.Fatalf("Expected no error but got %T (%s)", err, err)
	}

	err = tr.Apply(u)
	if err != nil {
		t.Fatalf("Expected no error but got %T (%s)", err, err)
	}

	s, err = tr.Commit(s)
	if err != nil {
		t.Fatalf("Expected no error but got %T (%s)", err, err)
	}

	c, err := s.Get("first", "str-str-map")
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		path := []Expression{
			MakeStringValue("key"),
			MakeStringValue("4-fourth")}
		v, err := c.Get(path, nil)
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		} else {
			e := "fourth"
			s, err := v.Serialize()
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else if s != e {
				t.Errorf("Expected %q but got %q", e, s)
			}
		}
	}

	c, err = s.Get("first", "str-net-map")
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		a, err := MakeValueFromString(TypeAddress, "192.0.2.49")
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		} else {
			path := []Expression{MakeStringValue("key"), a}
			v, err := c.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else {
				e := "fourth"
				s, err := v.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got %T (%s)", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		n, err := MakeValueFromString(TypeNetwork, "192.0.2.50/31")
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		} else {
			path := []Expression{MakeStringValue("key"), n}
			v, err := c.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else {
				e := "fourth"
				s, err := v.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got %T (%s)", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}
	}

	c, err = s.Get("first", "str-dom-map")
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		path := []Expression{
			MakeStringValue("key"),
			MakeDomainValue("example.gov")}
		v, err := c.Get(path, nil)
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		} else {
			e := "fourth"
			s, err := v.Serialize()
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else if s != e {
				t.Errorf("Expected %q but got %q", e, s)
			}
		}
	}
}
