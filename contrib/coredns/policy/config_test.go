package policy

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/infobloxopen/themis/pdp"
)

func TestPolicyConfigParse(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		err   error

		endpoints    []string
		options      map[uint16][]*edns0Opt
		debugSuffix  *string
		streams      *int
		hotSpot      *bool
		attrs        *attrsConfig
		debugID      *string
		passthrough  []string
		connTimeout  *time.Duration
		autoReqSize  *bool
		maxReqSize   *int
		autoResAttrs *bool
		maxResAttrs  *int
		cacheTTL     *time.Duration
		cacheLimit   *int
	}{
		{
			desc: "MissingPolicySection",
			input: `.:53 {
						log stdout
					}`,
			err: errors.New("Policy setup called without keyword 'policy' in Corefile"),
		},
		{
			desc: "InvalidOption",
			input: `.:53 {
						policy {
							error option
						}
					}`,
			err: errors.New("invalid policy plugin option"),
		},
		{
			desc: "NoEndpointArguemnts",
			input: `.:53 {
						policy {
							endpoint
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "SingleEntryEndpoint",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
						}
					}`,
			endpoints: []string{"10.2.4.1:5555"},
		},
		{
			desc: "TwoEntriesEndpoint",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555 10.2.4.2:5555
						}
					}`,
			endpoints: []string{"10.2.4.1:5555", "10.2.4.2:5555"},
		},
		{
			desc: "InvalidEDNS0Size",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex wrong_size 0 32
						}
					}`,
			err: errors.New("Could not parse EDNS0 data size"),
		},
		{
			desc: "EDNS0Hex",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex 32 0 32
						}
					}`,
			options: map[uint16][]*edns0Opt{
				0xfff0: {
					&edns0Opt{
						name:     "uid",
						dataType: typeEDNS0Hex,
						size:     32,
						start:    0,
						end:      32},
				},
			},
		},
		{
			desc: "InvalidEDNS0Code",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 wrong_hex uid hex
						}
					}`,
			err: errors.New("Could not parse EDNS0 code"),
		},
		{
			desc: "InvalidEDNS0StartIndex",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex 32 wrong_offset 32
						}
					}`,
			err: errors.New("Could not parse EDNS0 start index"),
		},
		{
			desc: "InvalidEDNS0EndIndex",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex 32 0 wrong_size
						}
					}`,
			err: errors.New("Could not parse EDNS0 end index"),
		},
		{
			desc: "EDNS0Hex2",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex 32 0 16
							edns0 0xfff0 id hex 32 16 32
						}
					}`,
			options: map[uint16][]*edns0Opt{
				0xfff0: {
					&edns0Opt{
						name:     "uid",
						dataType: typeEDNS0Hex,
						size:     32,
						start:    0,
						end:      16},
					&edns0Opt{
						name:     "id",
						dataType: typeEDNS0Hex,
						size:     32,
						start:    16,
						end:      32},
				},
			},
		},
		{
			desc: "InvalidEDNS0StartEndPair",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex 32 16 15
						}
					}`,
			err: errors.New("End index should be > start index"),
		},
		{
			desc: "InvalidEDNS0SizeEndPair",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex 32 0 33
						}
					}`,
			err: errors.New("End index should be <= size"),
		},
		{
			desc: "NotEnoughEDNS0Arguments",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff1
						}
					}`,
			err: errors.New("Invalid edns0 directive"),
		},
		{
			desc: "InvalidEDNS0Type",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff1 guid bin
						}
					}`,
			err: errors.New("Could not add EDNS0"),
		},
		{
			desc: "NoDebugQuerySuffixArguments",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_query_suffix
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "DebugQuerySuffix",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_query_suffix debug.local.
						}
					}`,
			debugSuffix: newStringPtr("debug.local."),
		},
		{
			desc: "PDPClientStreams",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams 10
						}
					}`,
			streams: newIntPtr(10),
		},
		{
			desc: "InvalidPDPClientStreams",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams Ten
						}
					}`,
			err: errors.New("Could not parse number of streams"),
		},
		{
			desc: "NoPDPClientStreamsArguments",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "NegativePDPClientStreams",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams -1
						}
					}`,
			err: errors.New("Expected at least one stream got -1"),
		},
		{
			desc: "PDPClientStreamsWithRoundRobin",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams 10 Round-Robin
						}
					}`,
			streams: newIntPtr(10),
			hotSpot: newBoolPtr(false),
		},
		{
			desc: "PDPClientStreamsWithHotSpot",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams 10 Hot-Spot
						}
					}`,
			streams: newIntPtr(10),
			hotSpot: newBoolPtr(true),
		},
		{
			desc: "InvalidPDPClientStreamsBalancer",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams 10 Unknown-Balancer
						}
					}`,
			err: errors.New("Expected round-robin or hot-spot balancing but got Unknown-Balancer"),
		},
		{
			desc: "ComplexAttributeConfig",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex 32 0 16
							edns0 0xfff1 id
							validation1 type="query" domain_name uid
							validation2 type="response" address pid
							default_decision policy_action=2
							metrics id policy_action
							dnstap 0 src="policy_plugin"
							dnstap 2 policy_id policy_plugin src="policy_plugin"
						}
					}`,
			attrs: &attrsConfig{
				attrInds: map[string]int{
					attrNameDomainName:   attrIndexDomainName,
					attrNameDNSQtype:     attrIndexDNSQtype,
					attrNameSourceIP:     attrIndexSourceIP,
					attrNameAddress:      attrIndexAddress,
					attrNamePolicyAction: attrIndexPolicyAction,
					attrNameRedirectTo:   attrIndexRedirectTo,
					attrNameLog:          attrIndexLog,
					"uid":                attrIndexCount + 0,
					"id":                 attrIndexCount + 1,
					"type":               attrIndexCount + 2,
					"pid":                attrIndexCount + 3,
					"src":                attrIndexCount + 4,
					"policy_id":          attrIndexCount + 5,
					"policy_plugin":      attrIndexCount + 6,
				},
				confLists: [attrListTypeDnstap + maxDnstapLists][]attrConf{
					{
						attrConf{"type", attrIndexCount + 2, pdp.MakeStringValue("query")},
						attrConf{"domain_name", attrIndexDomainName, pdp.UndefinedValue},
						attrConf{"uid", attrIndexCount + 0, pdp.UndefinedValue},
					},
					{
						attrConf{"type", attrIndexCount + 2, pdp.MakeStringValue("response")},
						attrConf{"address", attrIndexAddress, pdp.UndefinedValue},
						attrConf{"pid", attrIndexCount + 3, pdp.UndefinedValue},
					},
					{
						attrConf{"policy_action", attrIndexPolicyAction, pdp.MakeIntegerValue(2)},
					},
					{
						attrConf{"id", attrIndexCount + 1, pdp.UndefinedValue},
						attrConf{"policy_action", attrIndexPolicyAction, pdp.UndefinedValue},
					},
					{
						attrConf{"src", attrIndexCount + 4, pdp.MakeStringValue("policy_plugin")},
					},
					{},
					{
						attrConf{"policy_id", attrIndexCount + 5, pdp.UndefinedValue},
						attrConf{"policy_plugin", attrIndexCount + 6, pdp.UndefinedValue},
						attrConf{"src", attrIndexCount + 4, pdp.MakeStringValue("policy_plugin")},
					},
				},
			},
		},
		{
			desc: "BadDNStapArgument",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							dnstap wrong_number
						}
					}`,
			err: errors.New("invalid syntax"),
		},
		{
			desc: "BadDNStapArgumentRange",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							dnstap 10
						}
					}`,
			err: errors.New("Incorrect dnstap log level"),
		},
		{
			desc: "NoDNStapAttributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							dnstap 0
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "NoValidation1Attributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							validation1
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "NoValidation2Attributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							validation2
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "NoDefaultDecisionAttributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							default_decision
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "NoMetricsAttributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							metrics
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "BadAttributeValue",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							metrics id=bad_value
						}
					}`,
			err: errors.New("invalid attribute value"),
		},
		{
			desc: "DebugID",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_id corednsinstance
						}
					}`,
			debugID: newStringPtr("corednsinstance"),
		},
		{
			desc: "NoDebugIDArguments",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_id
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "Passthrough",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							passthrough google.com. facebook.org.
						}
					}`,
			passthrough: []string{
				"google.com.",
				"facebook.org.",
			},
		},
		{
			desc: "NoPassthroughArguments",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							passthrough
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "NoConnectionTimeoutArguments",
			input: `.:53 {
						policy {
							connection_timeout
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "NoConnectionTimeout",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							connection_timeout no
						}
					}`,
			connTimeout: newDurationPtr(-1),
		},
		{
			desc: "ConnectionTimeout",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							connection_timeout 500ms
						}
					}`,
			connTimeout: newDurationPtr(500 * time.Millisecond),
		},
		{
			desc: "InvalidConnectionTimeout",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							connection_timeout invalid
						}
					}`,
			err: errors.New("Could not parse timeout: time: invalid duration \"invalid\""),
		},
		{
			desc: "Log",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							log
						}
					}`,
		},
		{
			desc: "TrailingLogArgument",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							log stdout
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "MaxRequestSize",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_request_size 128
						}
					}`,
			autoReqSize: newBoolPtr(false),
			maxReqSize:  newIntPtr(128),
		},
		{
			desc: "MaxRequestSize",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_request_size auto
						}
					}`,
			autoReqSize: newBoolPtr(true),
			maxReqSize:  newIntPtr(-1),
		},
		{
			desc: "MaxRequestSize",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_request_size auto 128
						}
					}`,
			autoReqSize: newBoolPtr(true),
			maxReqSize:  newIntPtr(128),
		},
		{
			desc: "NoMaxRequestSizeArguments",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_request_size
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "InvalidMaxRequestSize",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_request_size test
						}
					}`,
			err: errors.New("Could not parse PDP request size limit"),
		},
		{
			desc: "OverflowMaxRequestSize",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_request_size 2147483648
						}
					}`,
			err: errors.New("Size limit 2147483648 (> 2147483647) for PDP request is too high"),
		},
		{
			desc: "MaxResponseAttributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_response_attributes 128
						}
					}`,
			autoResAttrs: newBoolPtr(false),
			maxResAttrs:  newIntPtr(128),
		},
		{
			desc: "MaxResponseAttributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_response_attributes auto
						}
					}`,
			autoResAttrs: newBoolPtr(true),
			maxResAttrs:  newIntPtr(64),
		},
		{
			desc: "NoMaxResponseAttributesArguments",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_response_attributes
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "InvalidMaxResponseAttributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_response_attributes invalid
						}
					}`,
			err: errors.New("Could not parse PDP response attributes limit"),
		},
		{
			desc: "OverflowMaxResponseAttributes",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							max_response_attributes 2147483648
						}
					}`,
			err: errors.New("Attributes limit 2147483648 (> 2147483647) for PDP response is too high"),
		},
		{
			desc: "NoDecisionCache",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
						}
					}`,
			cacheTTL: newDurationPtr(0),
		},
		{
			desc: "DecisionCache",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							cache
						}
					}`,
			cacheTTL:   newDurationPtr(10 * time.Minute),
			cacheLimit: newIntPtr(0),
		},
		{
			desc: "DecisionCacheWithTTL",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							cache 15s
						}
					}`,
			cacheTTL:   newDurationPtr(15 * time.Second),
			cacheLimit: newIntPtr(0),
		},
		{
			desc: "DecisionCacheWithTTLAndLimit",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							cache 15s 128
						}
					}`,
			cacheTTL:   newDurationPtr(15 * time.Second),
			cacheLimit: newIntPtr(128),
		},
		{
			desc: "TooManyCacheArguments",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							cache too many of them
						}
					}`,
			err: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			desc: "InvalidCacheTTL",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							cache invalid
						}
					}`,
			err: errors.New("Could not parse decision cache TTL"),
		},
		{
			desc: "WrongCacheTTL",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							cache -15s
						}
					}`,
			err: errors.New("Can't set decision cache TTL to"),
		},
		{
			desc: "InvalidCacheLimit",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							cache 15s invalid
						}
					}`,
			err: errors.New("Could not parse decision cache limit"),
		},
		{
			desc: "OverflowCacheLimit",
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							cache 15s 2147483648
						}
					}`,
			err: errors.New("Cache limit 2147483648 (> 2147483647) is too high"),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			c := caddy.NewTestController("dns", test.input)
			mw, err := policyParse(c)
			if err != nil {
				if test.err != nil {
					if !strings.Contains(err.Error(), test.err.Error()) {
						t.Errorf("Expected error '%v' but got '%v'\n", test.err, err)
					}
				} else {
					t.Errorf("Expected no error but got '%v'\n", err)
				}
			} else {
				if test.err != nil {
					t.Errorf("Expected error '%v' but got 'nil'\n", test.err)
				} else {
					if test.endpoints != nil {
						if len(test.endpoints) != len(mw.conf.endpoints) {
							t.Errorf("Expected endpoints %v but got %v\n", test.endpoints, mw.conf.endpoints)
						} else {
							for i := 0; i < len(test.endpoints); i++ {
								if test.endpoints[i] != mw.conf.endpoints[i] {
									t.Errorf("Expected endpoint '%s' but got '%s'\n",
										test.endpoints[i], mw.conf.endpoints[i])
								}
							}
						}
					}

					if test.options != nil {
						for k, testOpts := range test.options {
							if mwOpts, ok := mw.conf.options[k]; ok {
								if len(testOpts) != len(mwOpts) {
									t.Errorf("Expected %d EDNS0 options for 0x%04x but got %d",
										len(testOpts), k, len(mwOpts))
								} else {
									for i, testOpt := range testOpts {
										mwOpt := mwOpts[i]
										if testOpt.name != mwOpt.name ||
											testOpt.dataType != mwOpt.dataType ||
											testOpt.size != mwOpt.size ||
											testOpt.start != mwOpt.start ||
											testOpt.end != mwOpt.end {
											t.Errorf("Expected EDNS0 option:\n\t\"%#v\""+
												"\nfor 0x%04x at %d but got:\n\t\"%#v\"",
												*testOpt, k, i, *mwOpt)
										}
									}
								}
							} else {
								t.Errorf("Expected EDNS0 options 0x%04x but got nothing", k)
							}
						}

						for k := range mw.conf.options {
							if _, ok := test.options[k]; !ok {
								t.Errorf("Got unexpected options 0x%04x", k)
							}
						}
					}

					if test.debugSuffix != nil && *test.debugSuffix != mw.conf.debugSuffix {
						t.Errorf("Expected debug suffix %q but got %q", *test.debugSuffix, mw.conf.debugSuffix)
					}

					if test.streams != nil && *test.streams != mw.conf.streams {
						t.Errorf("Expected %d streams but got %d", *test.streams, mw.conf.streams)
					}

					if test.hotSpot != nil && *test.hotSpot != mw.conf.hotSpot {
						t.Errorf("Expected hotSpot=%v but got %v", *test.hotSpot, mw.conf.hotSpot)
					}

					if test.attrs != nil {
						if len(test.attrs.attrInds) != len(mw.conf.attrs.attrInds) {
							t.Errorf("Unexpected count of attributes in %+v: expected %d, but got %d",
								mw.conf.attrs.attrInds, len(test.attrs.attrInds), len(mw.conf.attrs.attrInds))
						}
						for an, ai := range test.attrs.attrInds {
							if mw.conf.attrs.attrInds[an] != ai {
								t.Errorf("Unexpected index of attribute %s: expected %d, but got %d",
									an, ai, mw.conf.attrs.attrInds[an])
							}
						}
						for lt, l := range test.attrs.confLists {
							for i, ac := range l {
								if ac.name != mw.conf.attrs.confLists[lt][i].name {
									t.Errorf("Unexpected name of %d attribute in conf list %d: expected %s, but got %s",
										i, lt, ac.name, mw.conf.attrs.confLists[lt][i].name)
								}
								if ac.index != mw.conf.attrs.confLists[lt][i].index {
									t.Errorf("Unexpected index of %d attribute in conf list %d: expected %d, but got %d",
										i, lt, ac.index, mw.conf.attrs.confLists[lt][i].index)
								}
								testutil.AssertValue(t, lt*10+i, ac.value, mw.conf.attrs.confLists[lt][i].value)
							}
						}
					}

					if test.debugID != nil && *test.debugID != mw.conf.debugID {
						t.Errorf("Expected debug id %q but got %q", *test.debugID, mw.conf.debugID)
					}

					if test.passthrough != nil {
						if len(test.passthrough) != len(mw.conf.passthrough) {
							t.Errorf("Expected %d passthrough suffixes but got %d",
								len(test.passthrough), len(mw.conf.passthrough))
						} else {
							for i, s := range test.passthrough {
								if s != mw.conf.passthrough[i] {
									t.Errorf("Expected %q passthrough suffix at %d but got %q",
										s, i, mw.conf.passthrough[i])
								}
							}
						}
					}

					if test.connTimeout != nil && *test.connTimeout != mw.conf.connTimeout {
						t.Errorf("Expected connection timeout %s but got %s", *test.connTimeout, mw.conf.connTimeout)
					}

					if test.autoReqSize != nil && *test.autoReqSize != mw.conf.autoReqSize {
						t.Errorf("Expected automatic request size %v but got %v",
							*test.autoReqSize, mw.conf.autoReqSize)
					}

					if test.maxReqSize != nil && *test.maxReqSize != mw.conf.maxReqSize {
						t.Errorf("Expected request size limit %d but got %d", *test.maxReqSize, mw.conf.maxReqSize)
					}

					if test.autoResAttrs != nil && *test.autoResAttrs != mw.conf.autoResAttrs {
						t.Errorf("Expected automatic response attributes %v but got %v",
							*test.autoResAttrs, mw.conf.autoResAttrs)
					}

					if test.maxResAttrs != nil && *test.maxResAttrs != mw.conf.maxResAttrs {
						t.Errorf("Expected response attributes limit %d but got %d",
							*test.maxResAttrs, mw.conf.maxResAttrs)
					}

					if test.cacheTTL != nil && *test.cacheTTL != mw.conf.cacheTTL {
						t.Errorf("Expected cache TTL %s but got %s", *test.cacheTTL, mw.conf.cacheTTL)
					}

					if test.cacheLimit != nil && *test.cacheLimit != mw.conf.cacheLimit {
						t.Errorf("Expected cache limit %d but got %d", *test.cacheLimit, mw.conf.cacheLimit)
					}
				}
			}
		})
	}
}

func newStringPtr(s string) *string {
	return &s
}

func newIntPtr(n int) *int {
	return &n
}

func newBoolPtr(b bool) *bool {
	return &b
}

func newDurationPtr(d time.Duration) *time.Duration {
	return &d
}
