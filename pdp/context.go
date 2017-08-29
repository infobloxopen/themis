package pdp

import (
	"fmt"
	"net"
	"strings"

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

var (
	effectNames = []string{
		"Deny",
		"Permit",
		"NotApplicable",
		"Indeterminate",
		"Indeterminate{D}",
		"Indeterminate{P}",
		"Indeterminate{DP}"}

	EffectIDs = map[string]int{
		"deny":   EffectDeny,
		"permit": EffectPermit}
)

type Context struct {
	a map[string]map[int]AttributeValue
	c *LocalContentStorage
}

func NewContext(c *LocalContentStorage, count int, f func(i int) (string, AttributeValue, error)) (*Context, error) {
	ctx := &Context{a: map[string]map[int]AttributeValue{}, c: c}

	for i := 0; i < count; i++ {
		ID, v, err := f(i)
		if err != nil {
			return nil, err
		}

		t := v.GetResultType()
		if m, ok := ctx.a[ID]; ok {
			if old, ok := m[t]; ok {
				return nil, newDuplicateAttributeValueError(ID, t, v, old)
			}

			m[t] = v
		} else {
			ctx.a[ID] = map[int]AttributeValue{t: v}
		}
	}

	return ctx, nil
}

func (c *Context) String() string {
	lines := []string{}
	if c.c != nil {
		if s := c.c.String(); len(s) > 0 {
			lines = append(lines, s)
		}
	}

	if len(c.a) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}

		lines = append(lines, "attributes:")
		for name, attrs := range c.a {
			for t, v := range attrs {
				lines = append(lines, fmt.Sprintf("- %s.(%s): %s", name, TypeNames[t], v.describe()))
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (c *Context) getAttribute(a Attribute) (AttributeValue, error) {
	t, ok := c.a[a.id]
	if !ok {
		return AttributeValue{}, bindError(newMissingAttributeError(), a.describe())
	}

	v, ok := t[a.t]
	if !ok {
		return AttributeValue{}, bindError(newMissingAttributeError(), a.describe())
	}

	return v, nil
}

func (c *Context) getContentItem(cID, iID string) (*ContentItem, error) {
	return c.c.Get(cID, iID)
}

func (c *Context) calculateBooleanExpression(e Expression) (bool, error) {
	v, err := e.calculate(c)
	if err != nil {
		return false, err
	}

	return v.boolean()
}

func (c *Context) calculateStringExpression(e Expression) (string, error) {
	v, err := e.calculate(c)
	if err != nil {
		return "", err
	}

	return v.str()
}

func (c *Context) calculateAddressExpression(e Expression) (net.IP, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.address()
}

func (c *Context) calculateDomainExpression(e Expression) (string, error) {
	v, err := e.calculate(c)
	if err != nil {
		return "", err
	}

	return v.domain()
}

func (c *Context) calculateNetworkExpression(e Expression) (*net.IPNet, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.network()
}

func (c *Context) calculateSetOfStringsExpression(e Expression) (*strtree.Tree, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfStrings()
}

func (c *Context) calculateSetOfNetworksExpression(e Expression) (*iptree.Tree, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfNetworks()
}

func (c *Context) calculateSetOfDomainsExpression(e Expression) (*domaintree.Node, error) {
	v, err := e.calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfDomains()
}

type Response struct {
	Effect      int
	status      boundError
	obligations []AttributeAssignmentExpression
}

func (r Response) Status() (int, []AttributeAssignmentExpression, error) {
	return r.Effect, r.obligations, r.status
}

type Evaluable interface {
	GetID() (string, bool)
	Calculate(ctx *Context) Response
	Append(path []string, v interface{}) (Evaluable, error)
	Delete(path []string) (Evaluable, error)
}
