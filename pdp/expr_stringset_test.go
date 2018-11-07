package pdp

import (
	"fmt"
	"testing"

	"github.com/infobloxopen/go-trees/strtree"
)

func TestSetOfStringsIntersect(t *testing.T) {
	ctx, err := NewContext(nil, 0, nil)
	if err != nil {
		t.Fatalf("Expected context but got error %s", err)
	}

	testCases := []struct {
		a, b, c *strtree.Tree
	}{
		{
			a: newStrTree("foo", "bar", "doo"),
			b: newStrTree("boo", "mar", "aoo"),
			c: newStrTree(),
		},
		{
			a: newStrTree("foo", "bar"),
			b: newStrTree("boo", "mar", "foo"),
			c: newStrTree("foo"),
		},
		{
			a: newStrTree("foo", "bar", "boo"),
			b: newStrTree("boo", "mar", "foo"),
			c: newStrTree("boo", "foo"),
		},
		{
			a: newStrTree("1", "2", "3"),
			b: newStrTree("4", "5", "6", "1"),
			c: newStrTree("1"),
		},
		{
			a: newStrTree("foo", "bar", "boo", "goo", "mar", "aoo", "noo", "moo"),
			b: newStrTree("mar", "boo", "foo", "bar", "zoo", "yoo", "aoo"),
			c: newStrTree("mar", "boo", "foo", "bar", "aoo"),
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Set of Strings Intersect %v + %v", tc.a, tc.b), func(t *testing.T) {
			a := MakeSetOfStringsValue(tc.a)
			b := MakeSetOfStringsValue(tc.b)
			e := makeFunctionSetOfStringsIntersect(a, b)

			v, err := e.Calculate(ctx)
			if err != nil {
				t.Errorf("Expect Calculate() returns no error, but got '%s'", err)
				return
			}

			res, err := v.setOfStrings()
			if err != nil {
				t.Errorf("Expect integer result with no error, but got '%s'", err)
			} else if !compareEnumeratedSets(tc.c, res, t) {
				t.Errorf("Expect result '%v', but got '%v'", tc.c, res)
			}
		})
	}
}

func TestSetOfStringsLen(t *testing.T) {
	ctx, err := NewContext(nil, 0, nil)
	if err != nil {
		t.Fatalf("Expected context but got error %s", err)
	}

	testCases := []struct {
		a *strtree.Tree
		b int64
	}{
		{
			a: newStrTree(),
			b: 0,
		},
		{
			a: newStrTree("foo", "bar"),
			b: 2,
		},
		{
			a: newStrTree("foo", "bar", "boo"),
			b: 3,
		},
		{
			a: newStrTree("foo", "bar", "boo", "goo", "mar", "aoo", "noo", "moo"),
			b: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Set of Strings Len %v", tc.a), func(t *testing.T) {
			a := MakeSetOfStringsValue(tc.a)
			e := makeFunctionSetOfStringsLen(a)

			v, err := e.Calculate(ctx)
			if err != nil {
				t.Errorf("Expect Calculate() returns no error, but got '%s'", err)
				return
			}

			res, err := v.integer()
			if err != nil {
				t.Errorf("Expect integer result with no error, but got '%s'", err)
			} else if res != tc.b {
				t.Errorf("Expect result '%v', but got '%v'", tc.b, res)
			}
		})
	}
}

func compareEnumeratedSets(a, b *strtree.Tree, t *testing.T) bool {
	first := a.Enumerate()
	second := b.Enumerate()
	f, fok := <-first
	s, sok := <-second
	for fok && sok {
		if f.Key != s.Key {
			return false
		}
		f, fok = <-first
		s, sok = <-second
	}
	return fok == sok
}
