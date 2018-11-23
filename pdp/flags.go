package pdp

import "strings"

// FlagsType instance represents cutom flags type.
type FlagsType struct {
	n string
	k string
	f map[string]int
	b []string
	c int
}

// NewFlagsType function creates new custom type with given name. A value of
// the type can take any combination of listed flags (including empty set).
// It supports up to 64 flags and flag names should be unique for the type.
func NewFlagsType(name string, flags ...string) (Type, error) {
	key := strings.ToLower(name)
	if _, ok := BuiltinTypes[key]; ok {
		return nil, newDuplicatesBuiltinTypeError(name)
	}

	if len(flags) <= 0 {
		return nil, newNoFlagsDefinedError(name, len(flags))
	}

	if len(flags) > 64 {
		return nil, newTooManyFlagsDefinedError(name, len(flags))
	}

	c := 8
	if len(flags) > 8 {
		c = 16
	}
	if len(flags) > 16 {
		c = 32
	}
	if len(flags) > 32 {
		c = 64
	}

	f := make(map[string]int, len(flags))
	for i, s := range flags {
		n := strings.ToLower(s)
		flags[i] = n

		if j, ok := f[n]; ok {
			return nil, newDuplicateFlagName(name, s, i, j)
		}
		f[n] = i
	}

	return &FlagsType{
		n: name,
		k: strings.ToLower(name),
		f: f,
		b: flags,
		c: c,
	}, nil
}

// String method returns human readable type name.
func (t *FlagsType) String() string {
	return t.n
}

// GetKey method returns case insensitive (always lowercase) type key.
func (t *FlagsType) GetKey() string {
	return t.k
}

// Match checks equivalence of different flags types. Flags types match iff
// they are defined for the same number of flags.
func (t *FlagsType) Match(ot Type) bool {
	fot, ok := ot.(*FlagsType)
	if !ok {
		return false
	}

	if t == fot {
		return true
	}

	return len(t.b) == len(fot.b)
}

// Capacity gets number of bits required to represent any flags combination.
func (t *FlagsType) Capacity() int {
	return t.c
}

// GetFlagBit method returns bit number for given flag name. If there is no flag
// with the name it returns -1.
func (t *FlagsType) GetFlagBit(f string) int {
	if n, ok := t.f[f]; ok {
		return n
	}

	return -1
}
