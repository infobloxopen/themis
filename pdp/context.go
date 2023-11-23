// Package pdp implements Policy Decision Point (PDP). It is responsible for
// making authorization decisions based on policies it has.
package pdp

import (
	"fmt"
	"net"
	"strings"

	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

// Effect* constants define possible consequences of decision evaluation.
const (
	// EffectDeny indicates that request is denied.
	EffectDeny = iota
	// EffectPermit indicates that request is permitted.
	EffectPermit

	// EffectNotApplicable indicates that policies don't contain any policy
	// and rule applicable to the request.
	EffectNotApplicable

	// EffectIndeterminate indicates that evaluation can't be done for
	// the request. For example required attribute is missing.
	EffectIndeterminate
	// EffectIndeterminateD indicates that evaluation can't be done for
	// the request but if it could effect would be EffectDeny.
	EffectIndeterminateD
	// EffectIndeterminateD indicates that evaluation can't be done for
	// the request but if it could effect would be EffectPermit.
	EffectIndeterminateP
	// EffectIndeterminateD indicates that evaluation can't be done for
	// the request but if it could effect would be only EffectDeny or
	// EffectPermit.
	EffectIndeterminateDP

	effectTotalCount
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

	// EffectIDs maps all possible values of rule's effect to its id.
	EffectIDs = map[string]int{
		"deny":   EffectDeny,
		"permit": EffectPermit}
)

// Context represents request context. The context contains all information
// needed to evaluate request.
type Context struct {
	a map[string]interface{}
	c *LocalContentStorage
}

// EffectNameFromEnum returns human readable name for Effect enum
func EffectNameFromEnum(effectEnum int) string {
	if effectEnum < 0 || effectEnum >= effectTotalCount {
		return fmt.Sprintf("<unknown %d>", effectEnum)
	}

	return effectNames[effectEnum]
}

// NewContext creates new instance of context. It requires pointer to local
// content storage and request attributes. The storage can be nil only
// if there is no policies or rules require it (otherwise evaluation may
// crash reaching it). Context collects input attributes by calling "f"
// function. The function is called "count" times and on each call it gets
// incrementing number starting from 0. The function should return attribute
// name and value. If "f" function returns error NewContext stops iterations
// and returns the same error. All pairs of attribute name and type should be
// unique.
func NewContext(c *LocalContentStorage, count int, f func(i int) (string, AttributeValue, error)) (*Context, error) {
	ctx := &Context{a: make(map[string]interface{}, count), c: c}

	for i := 0; i < count; i++ {
		ID, av, err := f(i)
		if err != nil {
			return nil, err
		}

		err = ctx.putAttribute(ID, av)
		if err != nil {
			return nil, err
		}
	}

	return ctx, nil
}

// NewContextFromBytes creates new instance of context. It requires a pointer
// to local content storage and a request represented as a byte sequence.
// Requirements for the storage are the same as NewContext function has.
// The request umarshaled to sequence of attributes as descirbed by
// (Un)MarshalRequest* functions help.
func NewContextFromBytes(c *LocalContentStorage, b []byte, currentContext *Context) (*Context, error) {
	off, err := checkRequestVersion(b)
	if err != nil {
		return nil, err
	}

	count, n, err := getRequestAttributeCount(b[off:])
	if err != nil {
		return nil, err
	}
	off += n

	var previousContext *Context

	// Create a new context when we call function for the first time in for loop
	if currentContext == nil {
		currentContext = &Context{a: make(map[string]interface{}, count), c: c}
	} else {
		// Keep track of old context while we fill in a new context for the current iteration
		previousContext = currentContext
		currentContext = &Context{a: make(map[string]interface{}, count), c: c}
	}

	attributeKeyValPairs, err := getAllRequestAttributes(b, off, count)
	if err != nil {
		return nil, err
	}

	var newCnameOrAddrFound bool
	var currCtxContainsAddr bool
	var currCtxContainsCName bool

	for _, pai := range attributeKeyValPairs {
		id := pai.ID
		typ := pai.Value.GetResultType()

		// Non-CNAME, non-Address types need to be in every context
		if typ != TypeAddress || typ != TypeDomain {
			if err := currentContext.putAttribute(id, pai.Value); err != nil {
				return nil, err
			}
		}

		if typ == TypeAddress {
			// Check if Address was added to previous context and if current context already has an address
			var addrExists bool
			if previousContext != nil {
				_, addrExists = previousContext.a[id]
			}

			if !addrExists && !currCtxContainsAddr {
				if err := currentContext.putAttribute(id, pai.Value); err != nil {
					return nil, err
				}
				currCtxContainsAddr = true
				newCnameOrAddrFound = true
			}
		}

		if typ == TypeDomain {
			// Check if CNAME was added to previous context and if current context already has an CNAME
			var cnameExists bool
			if previousContext != nil {
				_, cnameExists = previousContext.a[id]
			}

			if !cnameExists && !currCtxContainsCName {
				if err := currentContext.putAttribute(id, pai.Value); err != nil {
					return nil, err
				}
				currCtxContainsCName = true
				newCnameOrAddrFound = true
			}
		}
	}

	// If no new CNAME or Address was added to the current context, return nil
	if !newCnameOrAddrFound {
		return nil, nil
	}

	return currentContext, nil
}

// String implements Stringer interface.
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
			switch v := attrs.(type) {
			default:
				panic(fmt.Errorf("expected AttributeValue or map[int]AttributeValue but got: %T, %#v", v, v))

			case AttributeValue:
				lines = append(lines, fmt.Sprintf("- %s.(%s): %s", name, v.t, v.describe()))

			case map[Type]AttributeValue:
				for t, av := range v {
					lines = append(lines, fmt.Sprintf("- %s.(%s): %s", name, t, av.describe()))
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (c *Context) putAttribute(ID string, a AttributeValue) error {
	if v, ok := c.a[ID]; ok {
		switch v := v.(type) {
		default:
			panic(fmt.Errorf("expected AttributeValue or map[int]AttributeValue but got: %T, %#v", v, v))

		case AttributeValue:
			t := a.GetResultType()
			if v.t == t {
				return newDuplicateAttributeValueError(ID, t, a, v)
			}

			m := make(map[Type]AttributeValue, 2)
			m[v.t] = v
			m[t] = a

		case map[Type]AttributeValue:
			t := a.GetResultType()
			if old, ok := v[t]; ok {
				return newDuplicateAttributeValueError(ID, t, a, old)
			}

			v[t] = a
		}
	} else {
		c.a[ID] = a
	}

	return nil
}

func (c *Context) getAttribute(a Attribute) (AttributeValue, error) {
	v, ok := c.a[a.id]
	if !ok {
		return UndefinedValue, bindError(newMissingAttributeError(), a.describe())
	}

	switch v := v.(type) {
	case AttributeValue:
		if v.t != a.t {
			return UndefinedValue, bindError(newMissingAttributeError(), a.describe())
		}

		return v, nil

	case map[Type]AttributeValue:
		av, ok := v[a.t]
		if !ok {
			return UndefinedValue, bindError(newMissingAttributeError(), a.describe())
		}

		return av, nil
	}

	panic(fmt.Errorf("expected AttributeValue or map[int]AttributeValue but got: %T, %#v", v, v))
}

// GetContentItem returns content item value
func (c *Context) GetContentItem(cID, iID string) (*ContentItem, error) {
	return c.c.Get(cID, iID)
}

func (c *Context) calculateBooleanExpression(e Expression) (bool, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return false, err
	}

	return v.boolean()
}

func (c *Context) calculateStringExpression(e Expression) (string, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return "", err
	}

	return v.str()
}

func (c *Context) calculateIntegerExpression(e Expression) (int64, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.integer()
}

func (c *Context) calculateFloatExpression(e Expression) (float64, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.float()
}

func (c *Context) calculateFloatOrIntegerExpression(e Expression) (float64, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	if v.GetResultType() == TypeInteger {
		intVal, err := v.integer()
		if err != nil {
			return 0, err
		}

		return float64(intVal), nil
	}
	return v.float()
}

func (c *Context) calculateAddressExpression(e Expression) (net.IP, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.address()
}

func (c *Context) calculateDomainExpression(e Expression) (domain.Name, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return domain.Name{}, err
	}

	return v.domain()
}

func (c *Context) calculateNetworkExpression(e Expression) (*net.IPNet, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.network()
}

func (c *Context) calculateSetOfStringsExpression(e Expression) (*strtree.Tree, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfStrings()
}

func (c *Context) calculateSetOfNetworksExpression(e Expression) (*iptree.Tree, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfNetworks()
}

func (c *Context) calculateSetOfDomainsExpression(e Expression) (*domaintree.Node, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfDomains()
}

func (c *Context) calculateListOfStringsExpression(e Expression) ([]string, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.listOfStrings()
}

func (c *Context) calculateFlags8Expression(e Expression) (uint8, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.flags8()
}

func (c *Context) calculateFlags16Expression(e Expression) (uint16, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.flags16()
}

func (c *Context) calculateFlags32Expression(e Expression) (uint32, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.flags32()
}

func (c *Context) calculateFlags64Expression(e Expression) (uint64, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.flags64()
}

// Response represent result of policies evaluation.
type Response struct {
	// Effect is resulting effect.
	Effect int
	// Status contains an error if any occurs on response evaluation.
	Status error
	// Obligations constain set of obligations collected during evaluation.
	Obligations []AttributeAssignment
}

// calcValues returns calculated attributes and errors
func (r Response) calcValues(ctx *Context) ([]AttributeAssignment, []error) {
	var errs []error

	if r.Status != nil {
		errs = append(errs, newPolicyCalculationError(r.Status))
	}

	a := make([]AttributeAssignment, 0, len(r.Obligations))
	for _, e := range r.Obligations {
		v, err := e.e.Calculate(ctx)
		if err != nil {
			errs = append(errs, newObligationCalculationError(e.a, err))
			continue
		}

		a = append(a, MakeAttributeAssignment(e.a, v))
	}

	return a, errs
}

// Marshal encodes response as a sequence of bytes.
func (r Response) Marshal(ctx *Context) ([]byte, error) {
	a, errs := r.calcValues(ctx)

	return marshalResponse(r.Effect, a, errs...)
}

// MarshalWithAllocator encodes response as a sequence of bytes. It uses given
// allocator to create required response buffer. The allocator is expected
// to take number of bytes required and return slice of that length.
func (r Response) MarshalWithAllocator(f func(n int) ([]byte, error), ctx *Context) ([]byte, error) {
	a, errs := r.calcValues(ctx)

	return marshalResponseWithAllocator(f, r.Effect, a, errs...)
}

// MarshalToBuffer fills given byte array with marshalled representation
// of the response. The method returns number of bytes filled or error.
func (r Response) MarshalToBuffer(b []byte, ctx *Context) (int, error) {
	a, errs := r.calcValues(ctx)

	return marshalResponseToBuffer(b, r.Effect, a, errs...)
}

// Evaluable interface defines abstract PDP's entity which can be evaluated
// for given context (policy set or policy).
type Evaluable interface {
	GetID() (string, bool)
	Calculate(ctx *Context) Response
	Append(path []string, v interface{}) (Evaluable, error)
	Delete(path []string) (Evaluable, error)

	getOrder() int
	setOrder(ord int)
	describe() string
}
