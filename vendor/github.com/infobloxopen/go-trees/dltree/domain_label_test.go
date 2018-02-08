package dltree

import (
	"bytes"
	"testing"
)

func TestGetFirstLabelSize(t *testing.T) {
	name := "example"
	esize := 7
	enext := 7
	size, next := GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "example.com"
	esize = 7
	enext = 7
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "example\\.dot.com"
	esize = 11
	enext = 12
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "example\\\\slash"
	esize = 13
	enext = 14
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "example\\aletter"
	esize = 14
	enext = 15
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "example\\065letter"
	esize = 14
	enext = 17
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleletter\\065"
	esize = 14
	enext = 17
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleletter\\065."
	esize = 14
	enext = 17
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleletter\\065\\."
	esize = 15
	enext = 19
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleletter\\065\\\\"
	esize = 15
	enext = 19
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "example\\0invalid"
	esize = 16
	enext = 16
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "example\\06invalid"
	esize = 17
	enext = 17
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleinvalid\\0."
	esize = 16
	enext = 16
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleinvalid\\0\\."
	esize = 17
	enext = 18
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "example\\999invalid"
	esize = 18
	enext = 18
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleinvalid\\"
	esize = 15
	enext = 15
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleinvalid\\9"
	esize = 16
	enext = 16
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleinvalid\\99"
	esize = 17
	enext = 17
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}

	name = "exampleinvalid\\999"
	esize = 18
	enext = 18
	size, next = GetFirstLabelSize(name)
	if size != esize || next != enext {
		t.Errorf("expected %d bytes from %d bytes in %q domain label but got %d (%d)",
			esize, enext, name, size, next)
	}
}

func TestMakeDomainLabel(t *testing.T) {
	name := "example"
	data := []byte("example")
	dl, _ := MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "eXaMpLe"
	data = []byte("example")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example.com"
	data = []byte("example")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\.dot.com"
	data = []byte("example.dot")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\\\slash"
	data = []byte("example\\slash")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\aletter"
	data = []byte("examplealetter")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\Aletter"
	data = []byte("examplealetter")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\065letter"
	data = []byte("examplealetter")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\065Letter"
	data = []byte("examplealetter")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleletter\\065"
	data = []byte("examplelettera")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleletter\\065."
	data = []byte("examplelettera")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleletter\\065\\."
	data = []byte("examplelettera.")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleletter\\065\\\\"
	data = []byte("examplelettera\\")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\0invalid"
	data = []byte("example\\0invalid")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\0Invalid"
	data = []byte("example\\0invalid")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\06invalid"
	data = []byte("example\\06invalid")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleinvalid\\0."
	data = []byte("exampleinvalid\\0")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleinvalid\\0\\."
	data = []byte("exampleinvalid\\0.")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "example\\999invalid"
	data = []byte("example\\999invalid")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleinvalid\\"
	data = []byte("exampleinvalid\\")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleinvalid\\9"
	data = []byte("exampleinvalid\\9")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleinvalid\\99"
	data = []byte("exampleinvalid\\99")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}

	name = "exampleinvalid\\999"
	data = []byte("exampleinvalid\\999")
	dl, _ = MakeDomainLabel(name)
	if bytes.Compare(dl, data) != 0 {
		t.Errorf("got %d (% x) bytes for %q", len(dl), dl, name)
	}
}

func TestString(t *testing.T) {
	in := "example\\009\\.\\128\\013\\\\"
	dl, _ := MakeDomainLabel(in)
	out := dl.String()
	if out != in {
		t.Errorf("expected %q for domanin label (% x) but got %q", in, dl, out)
	}
}

func TestCompare(t *testing.T) {
	aName := "short"
	bName := "muchlonger"
	a, _ := MakeDomainLabel(aName)
	b, _ := MakeDomainLabel(bName)

	d := compare(a, b)
	if d >= 0 {
		t.Errorf("expected less than zero result but got %d", d)
	}

	d = compare(b, a)
	if d <= 0 {
		t.Errorf("expected greater than zero result but got %d", d)
	}

	aName = "aaa"
	bName = "bbb"
	a, _ = MakeDomainLabel(aName)
	b, _ = MakeDomainLabel(bName)

	d = compare(a, b)
	if d >= 0 {
		t.Errorf("expected less than zero result but got %d", d)
	}

	d = compare(b, a)
	if d <= 0 {
		t.Errorf("expected less than zero result but got %d", d)
	}

	aName = "equal"
	bName = "equal"
	a, _ = MakeDomainLabel(aName)
	b, _ = MakeDomainLabel(bName)

	d = compare(a, b)
	if d != 0 {
		t.Errorf("expected zero but got %d", d)
	}

	d = compare(b, a)
	if d != 0 {
		t.Errorf("expected zero but got %d", d)
	}
}
