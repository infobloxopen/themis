package policy

import (
	"net"
	"strconv"
	"testing"

	"github.com/infobloxopen/go-trees/domain"
	"github.com/miekg/dns"

	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/infobloxopen/themis/pdp"
)

func TestNewAttrHolder(t *testing.T) {
	ac := newAttrsConfig()

	// large buffer
	ah := newAttrHolder(make([]pdp.AttributeAssignment, 100), ac)
	if len(ah.attrs) != len(ac.attrInds) {
		t.Errorf("Unexpected attrHolder size, expected %d, actual %d", len(ac.attrInds), len(ah.attrs))
	}

	// too small buffer
	ah = newAttrHolder(make([]pdp.AttributeAssignment, 2), ac)
	if len(ah.attrs) != len(ac.attrInds) {
		t.Errorf("Unexpected attrHolder size, expected %d, actual %d", len(ac.attrInds), len(ah.attrs))
	}

	// no buffer
	ah = newAttrHolder(nil, ac)
	if len(ah.attrs) != len(ac.attrInds) {
		t.Errorf("Unexpected attrHolder size, expected %d, actual %d", len(ac.attrInds), len(ah.attrs))
	}

	for i, a := range ah.attrs {
		if a != emptyAttr {
			t.Errorf("Attribute %d is not empty: %v", i, a)
		}
	}
}

func TestAddDnsQuery(t *testing.T) {
	cfg := newConfig()
	cfg.parseEDNS0("0xfffc", attrNameSourceIP, "address")
	cfg.parseEDNS0("0xfffd", "low", "hex", "16", "0", "8")
	cfg.parseEDNS0("0xfffd", "high", "hex", "16", "8", "16")
	cfg.parseEDNS0("0xfffe", "byte", "bytes")

	m := testutil.MakeTestDNSMsgWithEdns0("example.com", dns.TypeA, dns.ClassINET,
		testutil.NewEdns0(
			testutil.NewEdns0Local(0xfffd,
				[]byte{
					0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
					0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
				},
			),
			testutil.NewEdns0Local(0xfffd, []byte{}),
			testutil.NewEdns0Local(0xfffe, []byte("test")),
			testutil.NewEdns0Local(0xfffc, []byte(net.ParseIP("2001:db8::1"))),
		),
	)
	w := testutil.NewTestAddressedNonwriter("192.0.2.1")

	ah := newAttrHolder(nil, cfg.attrs)
	ah.addDnsQuery(w, m, cfg.options)
	testutil.AssertAttrList(t, ah.attrs,
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		pdp.MakeStringAssignment(attrNameDNSQtype, strconv.FormatUint(uint64(dns.TypeA), 16)),
		pdp.MakeAddressAssignment(attrNameSourceIP, net.ParseIP("2001:db8::1")),
		emptyAttr, emptyAttr, emptyAttr, emptyAttr,
		pdp.MakeStringAssignment("low", "0001020304050607"),
		pdp.MakeStringAssignment("high", "08090a0b0c0d0e0f"),
		pdp.MakeStringAssignment("byte", "test"),
	)

	m = testutil.MakeTestDNSMsgWithEdns0("example.com", dns.TypeA, dns.ClassINET,
		testutil.NewEdns0(
			testutil.NewEdns0Local(0xfffd,
				[]byte{
					0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
					0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
				},
			),
			testutil.NewEdns0Local(0xfffd, []byte{}),
			testutil.NewEdns0Local(0xfffe, []byte("test")),
		),
	)

	ah = newAttrHolder(nil, cfg.attrs)
	ah.addDnsQuery(w, m, cfg.options)
	testutil.AssertAttrList(t, ah.attrs,
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		pdp.MakeStringAssignment(attrNameDNSQtype, strconv.FormatUint(uint64(dns.TypeA), 16)),
		pdp.MakeAddressAssignment(attrNameSourceIP, net.ParseIP("192.0.2.1")),
		emptyAttr, emptyAttr, emptyAttr, emptyAttr,
		pdp.MakeStringAssignment("low", "0001020304050607"),
		pdp.MakeStringAssignment("high", "08090a0b0c0d0e0f"),
		pdp.MakeStringAssignment("byte", "test"),
	)

	m = testutil.MakeTestDNSMsg("...", dns.TypeA, dns.ClassINET)
	testutil.AssertPanicWithError(t, "addDnsQuery(invalidDomainName)", func() {
		ah.addDnsQuery(w, m, nil)
	}, "Can't treat %q as domain name: %s", "...", domain.ErrEmptyLabel)
}

func TestAddAddressAttr(t *testing.T) {
	ah := newAttrHolder(nil, newAttrsConfig())

	ah.addAddressAttr(net.ParseIP("221.222.223.224"))
	testutil.AssertAttrList(t, ah.attrs,
		emptyAttr, emptyAttr, emptyAttr,
		pdp.MakeAddressAssignment(attrNameAddress, net.ParseIP("221.222.223.224")),
		emptyAttr, emptyAttr, emptyAttr,
	)
}

func TestAddAttrList(t *testing.T) {
	ah := newAttrHolder(nil, newAttrsConfig())
	ah.addAttrList([]pdp.AttributeAssignment{
		pdp.MakeAddressAssignment(attrNameAddress, net.ParseIP("192.0.2.1")),
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		pdp.MakeStringAssignment(attrNameRedirectTo, "test.com"),
		pdp.MakeIntegerAssignment(attrNamePolicyAction, actionBlock),
		pdp.MakeStringAssignment("reason", "tryagain"),
		pdp.MakeAddressAssignment("router", net.ParseIP("2001:aaa8::1111")),
		pdp.MakeIntegerAssignment("level", 5),
		pdp.MakeDomainAssignment("proxy", testutil.MakeTestDomain(dns.Fqdn("proxy.net"))),
	})
	testutil.AssertAttrList(t, ah.attrs,
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		emptyAttr, emptyAttr,
		pdp.MakeAddressAssignment(attrNameAddress, net.ParseIP("192.0.2.1")),
		pdp.MakeIntegerAssignment(attrNamePolicyAction, actionBlock),
		pdp.MakeStringAssignment(attrNameRedirectTo, "test.com"),
		emptyAttr,
	)

	cfg := newAttrsConfig()
	cfg.parseAttrList(attrListTypeVal2, "proxy")
	cfg.parseAttrList(attrListTypeMetrics, "level")
	ah = newAttrHolder(nil, cfg)
	ah.addAttrList([]pdp.AttributeAssignment{
		pdp.MakeAddressAssignment(attrNameAddress, net.ParseIP("192.0.2.1")),
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		pdp.MakeStringAssignment(attrNameRedirectTo, "test.com"),
		pdp.MakeIntegerAssignment(attrNamePolicyAction, actionBlock),
		pdp.MakeStringAssignment("reason", "tryagain"),
		pdp.MakeAddressAssignment("router", net.ParseIP("2001:aaa8::1111")),
		pdp.MakeIntegerAssignment("level", 5),
		pdp.MakeDomainAssignment("proxy", testutil.MakeTestDomain(dns.Fqdn("proxy.net"))),
	})
	testutil.AssertAttrList(t, ah.attrs,
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		emptyAttr, emptyAttr,
		pdp.MakeAddressAssignment(attrNameAddress, net.ParseIP("192.0.2.1")),
		pdp.MakeIntegerAssignment(attrNamePolicyAction, actionBlock),
		pdp.MakeStringAssignment(attrNameRedirectTo, "test.com"),
		emptyAttr,
		pdp.MakeDomainAssignment("proxy", testutil.MakeTestDomain(dns.Fqdn("proxy.net"))),
		pdp.MakeIntegerAssignment("level", 5),
	)
}

func TestAttrList(t *testing.T) {
	cfg := newAttrsConfig()
	cfg.parseAttrList(attrListTypeDefDecision, "dd1", "dd2", attrNameDomainName)
	cfg.parseAttrList(attrListTypeVal1, "v11", "v", "v13", attrNameDomainName)
	cfg.parseAttrList(attrListTypeVal2, "v21", "v", "v23", "v24", attrNamePolicyAction)
	cfg.parseAttrList(attrListTypeMetrics, "m1", "m3", attrNameDomainName)
	cfg.parseAttrList(attrListTypeDnstap, "d01", "d02", "d", attrNamePolicyAction)
	cfg.parseAttrList(attrListTypeDnstap+1, "d11", "d12", "d13", "d", attrNameDomainName)

	ah := newAttrHolder(nil, cfg)
	ah.addAttrList([]pdp.AttributeAssignment{
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		pdp.MakeStringAssignment("dd1", "dd1val"),
		pdp.MakeStringAssignment("v11", "v11val"),
		pdp.MakeStringAssignment("v", "vval"),
		pdp.MakeStringAssignment("v23", "v23val"),
		pdp.MakeStringAssignment("v24", "v24val"),
		pdp.MakeStringAssignment("m1", "m1val"),
		pdp.MakeStringAssignment("d", "dval"),
		pdp.MakeStringAssignment("d02", "d02val"),
		pdp.MakeStringAssignment("d13", "d13val"),
	})

	testutil.AssertAttrList(t, ah.attrList(nil, attrListTypeDefDecision),
		pdp.MakeStringAssignment("dd1", "dd1val"),
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
	)

	testutil.AssertAttrList(t, ah.attrList(nil, attrListTypeVal1),
		pdp.MakeStringAssignment("v11", "v11val"),
		pdp.MakeStringAssignment("v", "vval"),
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
	)

	testutil.AssertAttrList(t, ah.attrList(nil, attrListTypeVal2),
		pdp.MakeStringAssignment("v", "vval"),
		pdp.MakeStringAssignment("v23", "v23val"),
		pdp.MakeStringAssignment("v24", "v24val"),
	)

	testutil.AssertAttrList(t, ah.attrList(nil, attrListTypeMetrics),
		pdp.MakeStringAssignment("m1", "m1val"),
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
	)

	testutil.AssertAttrList(t, ah.attrList(nil, attrListTypeDnstap),
		pdp.MakeStringAssignment("d02", "d02val"),
		pdp.MakeStringAssignment("d", "dval"),
	)

	testutil.AssertAttrList(t, ah.attrList(nil, attrListTypeDnstap+1),
		pdp.MakeStringAssignment("d13", "d13val"),
		pdp.MakeStringAssignment("d", "dval"),
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
	)
}

func TestResetAttrList(t *testing.T) {
	cfg := newAttrsConfig()
	cfg.parseAttrList(attrListTypeDefDecision, "dds=\"DD1val\"", "dd0", "dda=192.168.0.1", "ddi=15")
	ah := newAttrHolder(nil, cfg)

	ah.resetAttrList(attrListTypeDefDecision)
	// check all attributes
	testutil.AssertAttrList(t, ah.attrs,
		emptyAttr, emptyAttr, emptyAttr, emptyAttr, emptyAttr, emptyAttr, emptyAttr,
		pdp.MakeStringAssignment("dds", "DD1val"),
		emptyAttr,
		pdp.MakeAddressAssignment("dda", net.ParseIP("192.168.0.1")),
		pdp.MakeIntegerAssignment("ddi", 15),
	)

	ah.addAttrList([]pdp.AttributeAssignment{
		pdp.MakeStringAssignment("dd0", "DD0val"),
		pdp.MakeIntegerAssignment("ddi", 2),
		pdp.MakeStringAssignment("dda", "InconsistentType"),
	})
	// attrList calls resetAttrList internally
	// check attrListTypeDefDecision attributes
	testutil.AssertAttrList(t, ah.attrList(nil, attrListTypeDefDecision),
		pdp.MakeStringAssignment("dds", "DD1val"),
		pdp.MakeStringAssignment("dd0", "DD0val"),
		pdp.MakeAddressAssignment("dda", net.ParseIP("192.168.0.1")),
		pdp.MakeIntegerAssignment("ddi", 15),
	)
}

func TestRedirectValue(t *testing.T) {
	ah := newAttrHolder(nil, newAttrsConfig())

	// empty attribute
	rv := ah.redirectValue()
	if rv != "" {
		t.Errorf("Unexpected redirect value: %s", rv)
	}

	// wrong type
	ah.attrs[attrIndexRedirectTo] = pdp.MakeIntegerAssignment(attrNameRedirectTo, 15)
	rv = ah.redirectValue()
	if rv != "" {
		t.Errorf("Unexpected redirect value: %s", rv)
	}

	// correct value
	ah.attrs[attrIndexRedirectTo] = pdp.MakeStringAssignment(attrNameRedirectTo, "test.com")
	rv = ah.redirectValue()
	if rv != "test.com" {
		t.Errorf("Unexpected redirect value: %s", rv)
	}
}

func TestActionValue(t *testing.T) {
	ah := newAttrHolder(nil, newAttrsConfig())

	// empty attribute
	av := ah.actionValue()
	if av != actionInvalid {
		t.Errorf("Unexpected action value: %d", av)
	}

	// wrong type
	ah.attrs[attrIndexPolicyAction] = pdp.MakeStringAssignment(attrNamePolicyAction, "block")
	av = ah.actionValue()
	if av != actionInvalid {
		t.Errorf("Unexpected action value: %d", av)
	}

	// wrong value range
	ah.attrs[attrIndexPolicyAction] = pdp.MakeIntegerAssignment(attrNamePolicyAction, 12345)
	av = ah.actionValue()
	if av != actionInvalid {
		t.Errorf("Unexpected action value: %d", av)
	}

	// correct value
	ah.attrs[attrIndexPolicyAction] = pdp.MakeIntegerAssignment(attrNamePolicyAction, 4)
	av = ah.actionValue()
	if av != 4 {
		t.Errorf("Unexpected action value: %d", av)
	}
}

func TestLogValue(t *testing.T) {
	ah := newAttrHolder(nil, newAttrsConfig())

	// empty attribute
	lv := ah.logValue()
	if lv != 0 {
		t.Errorf("Unexpected log value: %d", lv)
	}

	// wrong type
	ah.attrs[attrIndexLog] = pdp.MakeStringAssignment(attrNameLog, "brief")
	lv = ah.logValue()
	if lv != 0 {
		t.Errorf("Unexpected log value: %d", lv)
	}

	// wrong value range
	ah.attrs[attrIndexLog] = pdp.MakeIntegerAssignment(attrNameLog, maxDnstapLists+1)
	lv = ah.logValue()
	if lv != 0 {
		t.Errorf("Unexpected log value: %d", lv)
	}

	// correct value
	ah.attrs[attrIndexLog] = pdp.MakeIntegerAssignment(attrNameLog, maxDnstapLists-1)
	lv = ah.logValue()
	if lv != maxDnstapLists-1 {
		t.Errorf("Unexpected log value: %d", lv)
	}
}

func TestResetAttribute(t *testing.T) {
	ah := newAttrHolder(nil, newAttrsConfig())
	ah.attrs[attrIndexDNSQtype] = pdp.MakeStringAssignment("test", "testval")

	ah.resetAttribute(attrIndexDNSQtype)

	testutil.AssertAttr(t, 0, ah.attrs[attrIndexDNSQtype], emptyAttr)
}

func TestDnstapList(t *testing.T) {
	cfg := newAttrsConfig()
	cfg.parseAttrList(attrListTypeDnstap, "d0", "d")
	cfg.parseAttrList(attrListTypeDnstap+2, "d2", "d")
	ah := newAttrHolder(nil, cfg)
	ah.addAttrList([]pdp.AttributeAssignment{
		pdp.MakeStringAssignment("d", "Dval"),
		pdp.MakeStringAssignment("d1", "D1val"),
		pdp.MakeStringAssignment("d2", "D2val"),
		pdp.MakeStringAssignment("d0", "D0val"),
	})

	// empty/wrong log attribute
	testutil.AssertDnstapList(t, ah.dnstapList(),
		&pb.DnstapAttribute{"d0", "D0val"},
		&pb.DnstapAttribute{"d", "Dval"},
	)

	// log=1 - no attributes
	ah.addAttrList([]pdp.AttributeAssignment{pdp.MakeIntegerAssignment(attrNameLog, 1)})
	testutil.AssertDnstapList(t, ah.dnstapList())

	// log=2
	ah.addAttrList([]pdp.AttributeAssignment{pdp.MakeIntegerAssignment(attrNameLog, 2)})
	testutil.AssertDnstapList(t, ah.dnstapList(),
		&pb.DnstapAttribute{"d2", "D2val"},
		&pb.DnstapAttribute{"d", "Dval"},
	)

	// log=3 (out of range)
	ah.addAttrList([]pdp.AttributeAssignment{pdp.MakeIntegerAssignment(attrNameLog, 3)})
	testutil.AssertDnstapList(t, ah.dnstapList(),
		&pb.DnstapAttribute{"d0", "D0val"},
		&pb.DnstapAttribute{"d", "Dval"},
	)
}
