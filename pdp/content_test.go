package pdp

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/domaintree16"
	"github.com/infobloxopen/go-trees/domaintree32"
	"github.com/infobloxopen/go-trees/domaintree64"
	"github.com/infobloxopen/go-trees/domaintree8"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

func TestLocalContentStorage(t *testing.T) {
	tag := uuid.New()

	ssmc := MakeContentMappingItem(
		"str-str-map",
		TypeString,
		MakeSignature(TypeString, TypeString),
		MakeContentStringMap(strtree.NewTree()),
	)

	sTree := strtree.NewTree()
	sTree.InplaceInsert("1-first", "first")
	sTree.InplaceInsert("2-second", "second")
	sTree.InplaceInsert("3-third", "third")

	ksm := MakeContentMappingItem(
		"key-str-map",
		TypeString,
		MakeSignature(TypeString),
		MakeContentStringMap(sTree),
	)

	snmc := MakeContentMappingItem(
		"str-net-map",
		TypeString,
		MakeSignature(TypeString, TypeNetwork),
		MakeContentStringMap(strtree.NewTree()),
	)

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

	knm := MakeContentMappingItem(
		"key-net-map",
		TypeString,
		MakeSignature(TypeNetwork),
		MakeContentNetworkMap(nTree),
	)

	sdmc := MakeContentMappingItem(
		"str-dom-map",
		TypeString,
		MakeSignature(TypeString, TypeDomain),
		MakeContentStringMap(strtree.NewTree()),
	)

	dTree := new(domaintree.Node)
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	kdm := MakeContentMappingItem(
		"key-dom-map",
		TypeString,
		MakeSignature(TypeDomain),
		MakeContentDomainMap(dTree),
	)

	sTree = strtree.NewTree()
	sTree.InplaceInsert("1-first", "first")
	sTree.InplaceInsert("2-second", "second")
	sTree.InplaceInsert("3-third", "third")

	smc := MakeContentMappingItem(
		"str-map",
		TypeString,
		MakeSignature(TypeString),
		MakeContentStringMap(sTree),
	)

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

	nmc := MakeContentMappingItem(
		"net-map",
		TypeString,
		MakeSignature(TypeNetwork),
		MakeContentNetworkMap(nTree),
	)

	dTree = new(domaintree.Node)
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dmc := MakeContentMappingItem(
		"dom-map",
		TypeString,
		MakeSignature(TypeDomain),
		MakeContentDomainMap(dTree),
	)

	dTree = &domaintree.Node{}
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dmcAdd := MakeContentMappingItem(
		"dom-map-add",
		TypeString,
		MakeSignature(TypeDomain),
		MakeContentDomainMap(dTree),
	)

	dTree = new(domaintree.Node)
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dmcDel := MakeContentMappingItem(
		"dom-map-del",
		TypeString,
		MakeSignature(TypeDomain),
		MakeContentDomainMap(dTree),
	)

	dTree8 := &domaintree8.Node{}
	dTree8.InplaceInsert("example.com", 1)
	dTree8.InplaceInsert("example.net", 3)
	dTree8.InplaceInsert("example.org", 5)

	ft8, err := NewFlagsType("8flags",
		"f00", "f01", "f02", "f03", "f04", "f05", "f06", "f07",
	)
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}

	dm8c := MakeContentMappingItem(
		"dom8-map",
		ft8,
		MakeSignature(TypeDomain),
		MakeContentDomainFlags8Map(dTree8),
	)

	dTree16 := &domaintree16.Node{}
	dTree16.InplaceInsert("example.com", 1)
	dTree16.InplaceInsert("example.net", 3)
	dTree16.InplaceInsert("example.org", 5)

	ft16, err := NewFlagsType("16flags",
		"f00", "f01", "f02", "f03", "f04", "f05", "f06", "f07",
		"f10", "f11", "f12", "f13", "f14", "f15", "f16", "f17",
	)
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}

	dm16c := MakeContentMappingItem(
		"dom16-map",
		ft16,
		MakeSignature(TypeDomain),
		MakeContentDomainFlags16Map(dTree16),
	)

	dTree32 := &domaintree32.Node{}
	dTree32.InplaceInsert("example.com", 1)
	dTree32.InplaceInsert("example.net", 3)
	dTree32.InplaceInsert("example.org", 5)

	ft32, err := NewFlagsType("32flags",
		"f00", "f01", "f02", "f03", "f04", "f05", "f06", "f07",
		"f10", "f11", "f12", "f13", "f14", "f15", "f16", "f17",
		"f20", "f21", "f22", "f23", "f24", "f25", "f26", "f27",
		"f30", "f31", "f32", "f33", "f34", "f35", "f36", "f37",
	)
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}

	dm32c := MakeContentMappingItem(
		"dom32-map",
		ft32,
		MakeSignature(TypeDomain),
		MakeContentDomainFlags32Map(dTree32),
	)

	dTree64 := &domaintree64.Node{}
	dTree64.InplaceInsert("example.com", 1)
	dTree64.InplaceInsert("example.net", 3)
	dTree64.InplaceInsert("example.org", 5)

	ft64, err := NewFlagsType("64flags",
		"f00", "f01", "f02", "f03", "f04", "f05", "f06", "f07",
		"f10", "f11", "f12", "f13", "f14", "f15", "f16", "f17",
		"f20", "f21", "f22", "f23", "f24", "f25", "f26", "f27",
		"f30", "f31", "f32", "f33", "f34", "f35", "f36", "f37",
		"f40", "f41", "f42", "f43", "f44", "f45", "f46", "f47",
		"f50", "f51", "f52", "f53", "f54", "f55", "f56", "f57",
		"f60", "f61", "f62", "f63", "f64", "f65", "f66", "f67",
		"f70", "f71", "f72", "f73", "f74", "f75", "f76", "f77",
	)
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}

	dm64c := MakeContentMappingItem(
		"dom64-map",
		ft64,
		MakeSignature(TypeDomain),
		MakeContentDomainFlags64Map(dTree64),
	)

	items := []*ContentItem{ssmc, snmc, sdmc, smc, nmc, dmc, dmcDel, dm8c, dm16c, dm32c, dm64c}
	s := NewLocalContentStorage([]*LocalContent{NewLocalContent("first", &tag, items)})

	dTree = &domaintree.Node{}
	dTree.InplaceInsert("example.com", "first")
	dTree.InplaceInsert("example.net", "second")
	dTree.InplaceInsert("example.org", "third")

	dmc = MakeContentMappingItem(
		"dom-map",
		TypeString,
		MakeSignature(TypeDomain),
		MakeContentDomainMap(dTree),
	)
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

	u.Append(UOAdd, []string{"dom8-map", "example.gov"},
		MakeContentValueItem("example.gov", ft8, uint8(10)))
	u.Append(UODelete, []string{"dom8-map", "example.net"}, nil)

	u.Append(UOAdd, []string{"dom16-map", "example.gov"},
		MakeContentValueItem("example.gov", ft16, uint16(10)))
	u.Append(UODelete, []string{"dom16-map", "example.net"}, nil)

	u.Append(UOAdd, []string{"dom32-map", "example.gov"},
		MakeContentValueItem("example.gov", ft32, uint32(10)))
	u.Append(UODelete, []string{"dom32-map", "example.net"}, nil)

	u.Append(UOAdd, []string{"dom64-map", "example.gov"},
		MakeContentValueItem("example.gov", ft64, uint64(10)))
	u.Append(UODelete, []string{"dom64-map", "example.net"}, nil)

	u.Append(UOAdd, []string{"dom-map-add"}, dmcAdd)
	u.Append(UODelete, []string{"dom-map-del"}, nil)

	eUpd := fmt.Sprintf("content update: %s - %s\n"+
		"content: \"first\"\n"+
		"commands:\n"+
		"- Add path (\"str-str-map\"/\"key\")\n"+
		"- Add path (\"str-net-map\"/\"key\")\n"+
		"- Add path (\"str-dom-map\"/\"key\")\n"+
		"- Add path (\"str-str-map\"/\"key\"/\"4-fourth\")\n"+
		"- Delete path (\"str-str-map\"/\"key\"/\"3-third\")\n"+
		"- Add path (\"str-net-map\"/\"key\"/\"192.0.2.48/28\")\n"+
		"- Delete path (\"str-net-map\"/\"key\"/\"2001:db8::/32\")\n"+
		"- Add path (\"str-dom-map\"/\"key\"/\"example.gov\")\n"+
		"- Delete path (\"str-dom-map\"/\"key\"/\"example.net\")\n"+
		"- Add path (\"str-map\"/\"4-fourth\")\n"+
		"- Add path (\"net-map\"/\"2001:db8:1000::/40\")\n"+
		"- Add path (\"net-map\"/\"2001:db8:1000::1\")\n"+
		"- Delete path (\"str-map\"/\"3-third\")\n"+
		"- Add path (\"dom8-map\"/\"example.gov\")\n"+
		"- Delete path (\"dom8-map\"/\"example.net\")\n"+
		"- Add path (\"dom16-map\"/\"example.gov\")\n"+
		"- Delete path (\"dom16-map\"/\"example.net\")\n"+
		"- Add path (\"dom32-map\"/\"example.gov\")\n"+
		"- Delete path (\"dom32-map\"/\"example.net\")\n"+
		"- Add path (\"dom64-map\"/\"example.gov\")\n"+
		"- Delete path (\"dom64-map\"/\"example.net\")\n"+
		"- Add path (\"dom-map-add\")\n"+
		"- Delete path (\"dom-map-del\")", tag.String(), newTag.String())
	sUpd := u.String()
	if eUpd != sUpd {
		t.Errorf("Expected:\n%s\n\nbut got:\n%s\n\n", eUpd, sUpd)
	}

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
			MakeDomainValue(domain.WireNameLower("\x07example\x03gov\x00"))}
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

	c, err = s.Get("first", "dom8-map")
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		path := []Expression{
			MakeDomainValue(domain.WireNameLower("\x07example\x03gov\x00")),
		}
		v, err := c.Get(path, nil)
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		} else {
			e := "\"f01\",\"f03\""
			s, err := v.Serialize()
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else if s != e {
				t.Errorf("Expected %q but got %q", e, s)
			}
		}
	}

	c, err = s.Get("first", "dom16-map")
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		path := []Expression{
			MakeDomainValue(domain.WireNameLower("\x07example\x03gov\x00")),
		}
		v, err := c.Get(path, nil)
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		} else {
			e := "\"f01\",\"f03\""
			s, err := v.Serialize()
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else if s != e {
				t.Errorf("Expected %q but got %q", e, s)
			}
		}
	}

	c, err = s.Get("first", "dom32-map")
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		path := []Expression{
			MakeDomainValue(domain.WireNameLower("\x07example\x03gov\x00")),
		}
		v, err := c.Get(path, nil)
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		} else {
			e := "\"f01\",\"f03\""
			s, err := v.Serialize()
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else if s != e {
				t.Errorf("Expected %q but got %q", e, s)
			}
		}
	}

	c, err = s.Get("first", "dom64-map")
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		path := []Expression{
			MakeDomainValue(domain.WireNameLower("\x07example\x03gov\x00")),
		}
		v, err := c.Get(path, nil)
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		} else {
			e := "\"f01\",\"f03\""
			s, err := v.Serialize()
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else if s != e {
				t.Errorf("Expected %q but got %q", e, s)
			}
		}
	}
}
