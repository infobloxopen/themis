package pdp

import (
	"net"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

const (
	EffectDeny = iota
	EffectPermit

	EffectNotApplicable

	EffectIndeterminate
	EffectIndeterminateD
	EffectIndeterminateP
	EffectIndeterminateDP
)

var effectNames = []string{
	"Deny",
	"Permit",
	"NotApplicable",
	"Indeterminate",
	"Indeterminate{D}",
	"Indeterminate{P}",
	"Indeterminate{DP}"}

type Context struct {
	a map[string]map[int]attributeValue
	c *strtree.Tree
}

func (c *Context) getAttribute(a attribute) (attributeValue, error) {
	t, ok := c.a[a.id]
	if !ok {
		return attributeValue{}, a.newMissingError()
	}

	v, ok := t[a.t]
	if !ok {
		return attributeValue{}, a.newMissingError()
	}

	return v, nil
}

func (c *Context) calculateBooleanExpression(e expression) (bool, error) {
	v, err := e.calculate(c)
	if err != nil {
		return false, err
	}

	return v.boolean()
}

func (c *Context) calculateStringExpression(e expression) (string, error) {
	v, err := e.calculate(c)
	if err != nil {
		return "", err
	}

	return v.str()
}

func (c *Context) calculateAddressExpression(e expression) (net.IP, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.address()
}

func (c *Context) calculateDomainExpression(e expression) (string, error) {
	v, err := e.calculate(c)
	if err != nil {
		return "", err
	}

	return v.domain()
}

func (c *Context) calculateNetworkExpression(e expression) (*net.IPNet, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.network()
}

func (c *Context) calculateSetOfNetworksExpression(e expression) (*iptree.Tree, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfNetworks()
}

func (c *Context) calculateSetOfDomainsExpression(e expression) (*domaintree.Node, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfDomains()
}

type Response struct {
	Effect      int
	status      boundError
	obligations []attributeAssignmentExpression
}

type Evaluable interface {
	GetID() string
	Calculate(ctx *Context) Response
}
