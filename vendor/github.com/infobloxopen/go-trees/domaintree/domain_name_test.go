package domaintree

import (
	"fmt"
	"testing"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/infobloxopen/go-trees/dltree"
)

func TestSplit(t *testing.T) {
	dn := ""
	labels := split(dn)
	if len(labels) != 0 {
		t.Errorf("Expected zero labels for empty domain name %q but got %d", dn, len(labels))
	}

	dn = "."
	labels = split(dn)
	if len(labels) != 0 {
		t.Errorf("Expected zero labels for root fqdn %q but got %d", dn, len(labels))
	}

	dn = "www\\.test.com"
	labels = split(dn)
	assertDomainName(labels, []string{
		"com",
		"www\\.test",
	}, dn, t)

	dn = "www.test.com."
	labels = split(dn)
	assertDomainName(labels, []string{
		"com",
		"test",
		"www",
	}, dn, t)
}

func TestGetLabelsCount(t *testing.T) {
	dn := ""
	c := getLabelsCount(dn)
	if c != 0 {
		t.Errorf("Expected zero labels for empty domain name %q but got %d", dn, c)
	}

	dn = "."
	c = getLabelsCount(dn)
	if c != 0 {
		t.Errorf("Expected zero labels for root fqdn %q but got %d", dn, c)
	}

	dn = "www\\.test.com"
	c = getLabelsCount(dn)
	if c != 2 {
		t.Errorf("Expected two labels for domain name %q but got %d", dn, c)
	}

	dn = "www.test.com."
	c = getLabelsCount(dn)
	if c != 3 {
		t.Errorf("Expected three labels for fqdn %q but got %d", dn, c)
	}
}

func TestMakeWireDomainNameLower(t *testing.T) {
	dn := "example.com"
	wdn, err := MakeWireDomainNameLower(dn)
	if err != nil {
		t.Errorf("Expected no error for %q but got %s", dn, err)
	}

	if string(wdn) != "\x07example\x03com\x00" {
		t.Errorf("Got %q for %q", wdn, wdn)
	}

	dn = "tooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo.long.domain.label"
	wdn, err = MakeWireDomainNameLower(dn)
	if err != nil {
		if err != ErrLabelTooLong {
			t.Errorf("Expected error \"%s\" for %q but got \"%s\"", ErrLabelTooLong, dn, err)
		}
	} else {
		t.Errorf("Expected error for %q but got result %q", dn, wdn)
	}

	dn = "empty..domain.label"
	wdn, err = MakeWireDomainNameLower(dn)
	if err != nil {
		if err != ErrEmptyLabel {
			t.Errorf("Expected error \"%s\" for %q but got \"%s\"", ErrEmptyLabel, dn, err)
		}
	} else {
		t.Errorf("Expected error for %q but got result %q", dn, wdn)
	}

	dn = "toooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo." +
		"loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong." +
		"dooooooooooooooooooooooooooooooooooooooooooooooooooooooooooomai." +
		"naaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaame"
	wdn, err = MakeWireDomainNameLower(dn)
	if err != nil {
		if err != ErrNameTooLong {
			t.Errorf("Expected error \"%s\" for %q but got \"%s\"", ErrNameTooLong, dn, err)
		}
	} else {
		t.Errorf("Expected error for %q but got result %q", dn, wdn)
	}

	dn = "toooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo." +
		"loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong." +
		"dooooooooooooooooooooooooooooooooooooooooooooooooooooooooooomai." +
		"naaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaame." +
		"iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiis." +
		"toooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo." +
		"loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong"
	wdn, err = MakeWireDomainNameLower(dn)
	if err != nil {
		if err != ErrNameTooLong {
			t.Errorf("Expected error \"%s\" for %q but got \"%s\"", ErrNameTooLong, dn, err)
		}
	} else {
		t.Errorf("Expected error for %q but got result %q", dn, wdn)
	}
}

func TestWireDomainNameLowerString(t *testing.T) {
	dn := "example.com"
	wdn := WireDomainNameLower("\x07example\x03com\x00")
	sdn := wdn.String()
	if sdn != dn {
		t.Errorf("Expected %q for %q but got %q", dn, wdn, sdn)
	}

	dn = "example.com."
	wdn = WireDomainNameLower("\x07example\x03com\x05")
	sdn = wdn.String()
	if sdn != dn {
		t.Errorf("Expected %q for %q but got %q", dn, wdn, sdn)
	}

	dn = "example.com"
	wdn = WireDomainNameLower("\x07example\x03com")
	sdn = wdn.String()
	if sdn != dn {
		t.Errorf("Expected %q for %q but got %q", dn, wdn, sdn)
	}
}

func TestToLowerWireDomainName(t *testing.T) {
	wdn := WireDomainNameLower("\x07ExAmPlE\x03CoM\x00")
	ewdn := "\x07example\x03com\x00"
	wldn, err := ToLowerWireDomainName(wdn)
	if err != nil {
		t.Errorf("Expected no error for %q but got \"%s\"", string(wdn), err)
	} else if string(wldn) != ewdn {
		t.Errorf("Expected %q for %q but got %q", ewdn, string(wdn), string(wldn))
	}

	wdn = WireDomainNameLower(
		"\x3fTOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO" +
			"\x3fLOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOONG" +
			"\x3fDOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOMAI" +
			"\x3fNAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAME" +
			"\x3fIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIS" +
			"\x3fTOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO" +
			"\x3fLOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOONG" +
			"\x00")
	wldn, err = ToLowerWireDomainName(wdn)
	if err != nil {
		if err != ErrNameTooLong {
			t.Errorf("Expected error \"%s\" for %q but got \"%s\"", ErrNameTooLong, string(wdn), err)
		}
	} else {
		t.Errorf("Expected error for %q but got result %q", string(wdn), string(wldn))
	}

	wdn = WireDomainNameLower("\x05EMPTY\x00\x06DOMAIN\x05LABEL\x00")
	wldn, err = ToLowerWireDomainName(wdn)
	if err != nil {
		if err != ErrEmptyLabel {
			t.Errorf("Expected error \"%s\" for %q but got \"%s\"", ErrEmptyLabel, string(wdn), err)
		}
	} else {
		t.Errorf("Expected error for %q but got result %q", string(wdn), string(wldn))
	}

	wdn = WireDomainNameLower("\x0aCOMPRESSED\xff\xff")
	wldn, err = ToLowerWireDomainName(wdn)
	if err != nil {
		if err != ErrCompressedDN {
			t.Errorf("Expected error \"%s\" for %q but got \"%s\"", ErrCompressedDN, string(wdn), err)
		}
	} else {
		t.Errorf("Expected error for %q but got result %q", string(wdn), string(wldn))
	}

	wdn = WireDomainNameLower("\x05LABEL")
	wldn, err = ToLowerWireDomainName(wdn)
	if err != nil {
		if err != ErrLabelTooLong {
			t.Errorf("Expected error \"%s\" for %q but got \"%s\"", ErrLabelTooLong, string(wdn), err)
		}
	} else {
		t.Errorf("Expected error for %q but got result %q", string(wdn), string(wldn))
	}
}

func assertDomainName(labels []dltree.DomainLabel, elabels []string, dn string, t *testing.T) {
	for i := range elabels {
		elabels[i] += "\n"
	}

	s := make([]string, len(labels))
	for i, label := range labels {
		s[i] = label.String() + "\n"
	}

	ctx := difflib.ContextDiff{
		A:        elabels,
		B:        s,
		FromFile: "Expected",
		ToFile:   "Got"}

	diff, err := difflib.GetContextDiffString(ctx)
	if err != nil {
		panic(fmt.Errorf("can't compare labels for domain name \"%s\": %s", dn, err))
	}

	if len(diff) > 0 {
		t.Errorf("Labels for domain name \"%s\" don't match:\n%s", dn, diff)
	}
}
