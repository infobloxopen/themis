package policy

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

var errInvalidValue = errors.New("invalid attribute value")

const (
	attrNameDomainName   = "domain_name"
	attrNameDNSQtype     = "dns_qtype"
	attrNameSourceIP     = "source_ip"
	attrNameAddress      = "address"
	attrNamePolicyAction = "policy_action"
	attrNameRedirectTo   = "redirect_to"
	attrNameLog          = "log"
)

const (
	attrIndexDomainName = iota
	attrIndexDNSQtype
	attrIndexSourceIP
	attrIndexAddress
	attrIndexPolicyAction
	attrIndexRedirectTo
	attrIndexLog

	attrIndexCount
)

const (
	actionInvalid = iota
	actionDrop
	actionAllow
	actionBlock
	actionRedirect
	actionRefuse

	actionsTotal
)

type attrConf struct {
	name  string
	index int
	value pdp.AttributeValue
}

const maxDnstapLists = 3

const (
	attrListTypeVal1 = iota
	attrListTypeVal2
	attrListTypeDefDecision
	attrListTypeMetrics

	attrListTypeDnstap //must be last
)

type attrsConfig struct {
	attrInds  map[string]int
	confLists [attrListTypeDnstap + maxDnstapLists][]attrConf
}

func newAttrsConfig() *attrsConfig {
	return &attrsConfig{
		attrInds: map[string]int{
			attrNameDomainName:   attrIndexDomainName,
			attrNameDNSQtype:     attrIndexDNSQtype,
			attrNameSourceIP:     attrIndexSourceIP,
			attrNameAddress:      attrIndexAddress,
			attrNamePolicyAction: attrIndexPolicyAction,
			attrNameRedirectTo:   attrIndexRedirectTo,
			attrNameLog:          attrIndexLog,
		},
	}
}

func (ac *attrsConfig) provideIndex(name string) int {
	ind, ok := ac.attrInds[name]
	if !ok {
		ind = len(ac.attrInds)
		ac.attrInds[name] = ind
	}
	return ind
}

func (ac *attrsConfig) parseAttrList(listType int, argList ...string) error {
	list := make([]attrConf, 0, len(argList))
	for _, arg := range argList {
		name, val, err := parseAttrArg(arg)
		if err != nil {
			return err
		}
		ind := ac.provideIndex(name)
		list = append(list, attrConf{name: name, index: ind, value: val})
	}
	ac.confLists[listType] = list
	return nil
}

func parseAttrArg(arg string) (name string, val pdp.AttributeValue, e error) {
	tokens := strings.SplitN(arg, "=", 2)
	name = tokens[0]
	val = pdp.UndefinedValue
	if len(tokens) < 2 {
		return
	}

	if ip := net.ParseIP(tokens[1]); ip != nil {
		val = pdp.MakeAddressValue(ip)
		return
	}

	if num, err := strconv.ParseInt(tokens[1], 0, 64); err == nil {
		val = pdp.MakeIntegerValue(num)
		return
	}

	if str, err := strconv.Unquote(tokens[1]); err == nil {
		val = pdp.MakeStringValue(str)
		return
	}

	return "", pdp.UndefinedValue, errInvalidValue
}

func (ac *attrsConfig) hasMetrics() bool {
	return len(ac.confLists[attrListTypeMetrics]) > 0
}

func serializeOrPanic(a pdp.AttributeAssignment) string {
	v, err := a.GetValue()
	if err != nil {
		panic(err)
	}

	s, err := v.Serialize()
	if err != nil {
		panic(err)
	}

	return s
}
