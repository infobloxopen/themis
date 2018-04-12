package pdp

import "testing"

func TestBuiltinTypes(t *testing.T) {
	for k, bt := range BuiltinTypes {
		if k != bt.GetKey() {
			t.Errorf("expected %q key but got %q", k, bt.GetKey())
		}

		if len(bt.String()) <= 0 {
			t.Errorf("exepcted some human readable name for type %q but got empty string", k)
		}
	}
}

func TestSignature(t *testing.T) {
	sign := MakeSignature(
		TypeUndefined,
		TypeBoolean,
		TypeString,
		TypeInteger,
		TypeFloat,
		TypeAddress,
		TypeNetwork,
		TypeDomain,
		TypeSetOfStrings,
		TypeSetOfNetworks,
		TypeSetOfDomains,
		TypeListOfStrings,
	)

	e := "\"Undefined\"/" +
		"\"Boolean\"/" +
		"\"String\"/" +
		"\"Integer\"/" +
		"\"Float\"/" +
		"\"Address\"/" +
		"\"Network\"/" +
		"\"Domain\"/" +
		"\"Set of Strings\"/" +
		"\"Set of Networks\"/" +
		"\"Set of Domains\"/" +
		"\"List of Strings\""
	s := sign.String()
	if s != e {
		t.Errorf("expected %s signature but got %s", e, s)
	}

	sign = MakeSignature()
	e = "empty"
	s = sign.String()
	if s != e {
		t.Errorf("expected %s signature but got %s", e, s)
	}
}

func TestTypeSet(t *testing.T) {
	set := makeTypeSet(
		TypeUndefined,
		TypeBoolean,
		TypeString,
		TypeAddress,
		TypeNetwork,
		TypeDomain,
		TypeSetOfStrings,
		TypeSetOfNetworks,
		TypeSetOfDomains,
		TypeListOfStrings,
	)

	e := "\"Address\", " +
		"\"Boolean\", " +
		"\"Domain\", " +
		"\"List of Strings\", " +
		"\"Network\", " +
		"\"Set of Domains\", " +
		"\"Set of Networks\", " +
		"\"Set of Strings\", " +
		"\"String\", " +
		"\"Undefined\""
	s := set.String()
	if s != e {
		t.Errorf("expected %s signature but got %s", e, s)
	}

	if !set.Contains(TypeAddress) {
		t.Errorf("expected %q in the set but it isn't here", TypeAddress)
	}

	if set.Contains(TypeInteger) {
		t.Errorf("expected %q to be not in the set but it is here", TypeInteger)
	}

	set = makeTypeSet()
	e = "empty"
	s = set.String()
	if s != e {
		t.Errorf("expected %s signature but got %s", e, s)
	}
}
