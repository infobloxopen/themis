package bitradix

import (
	"net"
	"reflect"
	"testing"
)

type bittest struct {
	key uint32
	bit int
}

var tests = map[uint32]uint32{
	0x80000000: 2012,
	0x40000000: 2010,
	0x90000000: 2013,
}

const bits32 = 5

func newTree32() *Radix32 {
	r := New32()
	for k, v := range tests {
		r.Insert(k, bits32, v)
	}
	return r
}

func TestInsert(t *testing.T) {
	tests := map[bittest]uint32{
		bittest{0x81000000, 9}: 2012,
		bittest{0x80000000, 2}: 2013,
	}
	r := New32()
	for bits, value := range tests {
		t.Logf("Inserting %032b/%d\n", bits.key, bits.bit)
		if x := r.Insert(bits.key, bits.bit, value); x.Value != value {
			t.Logf("Expected %d, got %d for %d (node type %v)\n", value, x.Value, bits.key, x.Leaf())
			t.Fail()
		}
		t.Logf("Tree\n")
		r.Do(func(r1 *Radix32, i int) { t.Logf("(%2d): %032b/%d -> %d\n", i, r1.key, r1.bits, r1.Value) })
	}
}

func TestInsert2(t *testing.T) {
	tests := map[bittest]uint32{
		bittest{0x81000000, 9}: 2012,
		bittest{0xA0000000, 4}: 1000,
		bittest{0x80000000, 2}: 2013,
	}
	r := New32()
	for bits, value := range tests {
		t.Logf("Inserting %032b/%d\n", bits.key, bits.bit)
		if x := r.Insert(bits.key, bits.bit, value); x.Value != value {
			t.Logf("Expected %d, got %d for %d (node type %v)\n", value, x.Value, bits.key, x.Leaf())
			t.Fail()
		}
		t.Logf("Tree\n")
		r.Do(func(r1 *Radix32, i int) { t.Logf("(%2d): %032b/%d -> %d\n", i, r1.key, r1.bits, r1.Value) })
	}
}

func TestInsertIdempotent(t *testing.T) {
	r := New32()
	r.Insert(0x80000000, bits32, 2012)
	t.Logf("Tree\n")
	r.Do(func(r1 *Radix32, i int) { t.Logf("(%2d): %032b/%d -> %d\n", i, r1.key, r1.bits, r1.Value) })
	r.Insert(0x80000000, bits32, 2013)
	t.Logf("Tree\n")
	r.Do(func(r1 *Radix32, i int) { t.Logf("(%2d): %032b/%d -> %d\n", i, r1.key, r1.bits, r1.Value) })
	if x := r.Find(0x80000000, bits32); x.Value != 2013 {
		t.Logf("Expected %d, got %d for %d\n", 2013, x.Value, 0x08)
		t.Fail()
	}
}

func TestFindExact(t *testing.T) {
	tests := map[uint32]uint32{
		0x80000000: 2012,
		0x40000000: 2010,
		0x90000000: 2013,
	}
	r := New32()
	for k, v := range tests {
		t.Logf("Tree after insert of %032b (%x %d)\n", k, k, k)
		r.Insert(k, bits32, v)
		r.Do(func(r1 *Radix32, i int) { t.Logf("%p (%2d): %032b/%d -> %d\n", r1, i, r1.key, r1.bits, r1.Value) })
	}
	for k, v := range tests {
		x := r.Find(k, bits32)
		if x == nil {
			t.Logf("Got nil for %032b\n", k)
			t.Fail()
			continue
		}
		if x.Value != v {
			t.Logf("Expected %d, got %d for %032b (node type %v)\n", v, x.Value, k, x.Leaf())
			t.Fail()
		}
	}
}

func TestRemove(t *testing.T) {
	r := newTree32()
	t.Logf("Tree complete\n")
	r.Do(func(r1 *Radix32, i int) {
		t.Logf("[%010p %010p] (%2d): %032b/%d -> %d\n", r1.branch[0], r1.branch[1], i, r1.key, r1.bits, r1.Value)
	})
	k, v := uint32(0x40000000), uint32(2010)
	t.Logf("Tree after removal of %032b/%d %d (%x %d)\n", k, bits32, v, k, k)
	r.Remove(k, bits32)
	r.Do(func(r1 *Radix32, i int) {
		t.Logf("[%010p %010p] (%2d): %032b/%d -> %d\n", r1.branch[0], r1.branch[1], i, r1.key, r1.bits, r1.Value)
	})
	k, v = uint32(0x80000000), uint32(2012)
	t.Logf("Tree after removal of %032b/%d %d (%x %d)\n", k, bits32, v, k, k)
	r.Remove(k, bits32)
	r.Do(func(r1 *Radix32, i int) {
		t.Logf("[%010p %010p] (%2d): %032b/%d -> %d\n", r1.branch[0], r1.branch[1], i, r1.key, r1.bits, r1.Value)
	})
	k, v = uint32(0x90000000), uint32(2013)
	t.Logf("Tree after removal of %032b/%d %d (%x %d)\n", k, bits32, v, k, k)
	r.Remove(k, bits32)
	r.Do(func(r1 *Radix32, i int) {
		t.Logf("[%010p %010p] (%2d): %032b/%d -> %d\n", r1.branch[0], r1.branch[1], i, r1.key, r1.bits, r1.Value)
	})
}

// Insert one value and remove it again
func TestRemove2(t *testing.T) {
	r := New32()
	t.Logf("Tree empty\n")
	r.Do(func(r1 *Radix32, i int) {
		t.Logf("[%010p %010p] (%2d): %032b/%d -> %d\n", r1.branch[0], r1.branch[1], i, r1.key, r1.bits, r1.Value)
	})
	k, v := uint32(0x90000000), uint32(2013)
	r.Insert(k, bits32, v)

	t.Logf("Tree complete\n")
	r.Do(func(r1 *Radix32, i int) {
		t.Logf("[%010p %010p] (%2d): %032b/%d -> %d\n", r1.branch[0], r1.branch[1], i, r1.key, r1.bits, r1.Value)
	})
	t.Logf("Tree after removal of %032b/%d %d (%x %d)\n", k, bits32, v, k, k)
	r.Remove(k, bits32)
	r.Do(func(r1 *Radix32, i int) {
		t.Logf("[%010p %010p] (%2d): %032b/%d -> %d\n", r1.branch[0], r1.branch[1], i, r1.key, r1.bits, r1.Value)
	})
}

// Test with "real-life" ip addresses
func ipToUint(t *testing.T, n *net.IPNet) (i uint32, mask int) {
	ip := n.IP.To4()
	fv := reflect.ValueOf(&i).Elem()
	fv.SetUint(uint64(uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[+3])))
	mask, _ = n.Mask.Size()
	return
}

func uintToIP(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

func addRoute(t *testing.T, r *Radix32, s string, asn uint32) {
	_, ipnet, _ := net.ParseCIDR(s)
	net, mask := ipToUint(t, ipnet)
	t.Logf("Route %s (%032b), AS %d\n", s, net, asn)
	r.Insert(net, mask, asn)
}

func findRoute(t *testing.T, r *Radix32, s string) interface{} {
	_, ipnet, _ := net.ParseCIDR(s)
	net, mask := ipToUint(t, ipnet)
	t.Logf("Search %18s %032b/%d\n", s, net, mask)
	node := r.Find(net, mask)
	if node == nil {
		return uint32(0)
	}
	return node.Value
}

func TestFindIP(t *testing.T) {
	r := New32()
	// not a map to have influence on the order
	addRoute(t, r, "10.0.0.2/8", 10)
	addRoute(t, r, "10.20.0.0/14", 20)
	addRoute(t, r, "10.21.0.0/16", 21)
	addRoute(t, r, "192.168.0.0/16", 192)
	addRoute(t, r, "192.168.2.0/24", 1922)

	addRoute(t, r, "8.0.0.0/9", 3356)
	addRoute(t, r, "8.8.8.0/24", 15169)

	r.Do(func(r1 *Radix32, i int) {
		t.Logf("(%2d): %032b/%d -> %d\n", i, r1.key, r1.bits, r1.Value)
	})
	testips := map[string]uint32{
		"10.20.1.2/32":   20,
		"10.22.1.2/32":   20,
		"10.19.0.1/32":   10,
		"10.21.0.1/32":   21,
		"192.168.2.3/32": 1922,
		"230.0.0.1/32":   0,

		"8.8.8.8/32": 15169,
		"8.8.7.1/32": 3356,
	}

	for ip, asn := range testips {
		if x := findRoute(t, r, ip); asn != x {
			if x == nil && asn == 0 {
				continue
			}
			t.Logf("Expected %d, got %d for %ss\n", asn, x, ip)
			t.Fail()
		}
	}
}

func TestFindIPShort(t *testing.T) {
	r := New32()
	// not a map to have influence on the inserting order
	// The /14 will overwrite the /10 ...
	addRoute(t, r, "10.0.0.2/8", 10)
	addRoute(t, r, "10.0.0.0/14", 11)
	addRoute(t, r, "10.20.0.0/14", 20)
	addRoute(t, r, "210.168.0.0/17", 4694)
	addRoute(t, r, "210.168.96.0/19", 2554)
	addRoute(t, r, "210.168.192.0/18", 2516)
	addRoute(t, r, "210.169.0.0/17", 2516)
	addRoute(t, r, "210.168.128.0/18", 4716)
	addRoute(t, r, "210.169.128.0/17", 4725)
	addRoute(t, r, "210.169.212.0/24", 4725)
	addRoute(t, r, "210.16.14.0/24", 4759)
	addRoute(t, r, "210.16.0.0/24", 4759)
	addRoute(t, r, "210.16.1.0/24", 4759)
	addRoute(t, r, "210.16.40.0/24", 4759)
	addRoute(t, r, "210.166.0.0/19", 7672)
	addRoute(t, r, "210.166.5.0/24", 7668)
	addRoute(t, r, "210.167.0.0/19", 7668)
	addRoute(t, r, "210.166.0.0/20", 7672)
	addRoute(t, r, "210.166.96.0/19", 4693)
	addRoute(t, r, "210.167.112.0/20", 4685)
	addRoute(t, r, "210.167.128.0/18", 4716)
	addRoute(t, r, "210.167.192.0/18", 4716)
	addRoute(t, r, "210.167.32.0/19", 7663)
	addRoute(t, r, "210.166.209.0/24", 7663)
	addRoute(t, r, "210.166.211.0/24", 7663)

	r.Do(func(r1 *Radix32, i int) { t.Logf("(%2d): %032b/%d -> %d\n", i, r1.key, r1.bits, r1.Value) })

	testips := map[string]uint32{
		"10.20.1.2/32":     20,
		"10.19.0.1/32":     0, // because 10.0.0.2/8 isn't there this return 0
		"10.0.0.2/32":      11,
		"10.1.0.1/32":      11,
		"210.169.0.0/17":   2516,
		"210.168.96.0/19":  2554,
		"210.16.14.0/24":   4759,
		"210.16.0.0/24":    4759,
		"210.16.1.0/24":    4759,
		"210.16.40.0/24":   4759,
		"210.166.0.0/19":   7672,
		"210.166.0.0/20":   7672,
		"210.166.96.0/19":  4693,
		"210.167.112.0/20": 4685,
		"210.167.128.0/18": 4716,
		"210.167.192.0/18": 4716,
		"210.167.32.0/19":  7663,
	}

	for ip, asn := range testips {
		if x := findRoute(t, r, ip); asn != x {
			t.Logf("Expected %d, got %d for %s\n", asn, x, ip)
			t.Fail()
		}
	}
}

func TestFindMySelf(t *testing.T) {
	r := New32()
	routes := map[string]uint32{
		"210.168.0.0/17":   4694,
		"210.168.96.0/19":  2554,
		"210.168.192.0/18": 2516,
		"210.169.0.0/17":   2516,
		"210.168.128.0/18": 4716,
		"210.169.128.0/17": 4725,
		"210.169.212.0/24": 4725,
		"210.16.14.0/24":   4759,
		"210.16.0.0/24":    4759,
		"210.16.1.0/24":    4759,
		"210.16.40.0/24":   4759,
		"210.166.0.0/19":   7672,
		"210.166.5.0/24":   7668,
		"210.167.0.0/19":   7668,
		"210.166.0.0/20":   7672,
		"210.166.96.0/19":  4693,
		"210.167.112.0/20": 4685,
		"210.167.128.0/18": 4716,
		"210.167.192.0/18": 4716,
		"210.167.32.0/19":  7663,
		"210.166.209.0/24": 7663,
		"210.166.211.0/24": 7663,
		"87.71.192.0/18":   1001,
		"87.71.128.0/18":   1001,
	}
	for ip, asn := range routes {
		addRoute(t, r, ip, asn)
	}
	fail := false
	for ip, asn := range routes {
		if x := findRoute(t, r, ip); asn != x {
			t.Logf("Expected %d, got %d for %s\n", asn, x, ip)
			fail = true
			t.Fail()
		}
	}
	if fail {
		r.Do(func(r1 *Radix32, i int) {
			t.Logf("(%2d): %032b/%d %s -> %d\n", i, r1.key, r1.bits, uintToIP(r1.key), r1.Value)
		})
	}
}

func TestFindOverwrite(t *testing.T) {
	r := New32()
	routes := map[string]uint32{
		"1.0.20.0/23": 2518,
		"1.0.22.0/23": 2519,
		"1.0.24.0/23": 2520,
		"1.0.28.0/22": 2517,
		"1.0.64.0/18": 18144,
	}
	for ip, asn := range routes {
		addRoute(t, r, ip, asn)
	}
	r.Do(func(r1 *Radix32, i int) {
		t.Logf("(%2d): %032b/%d %s -> %d\n", i, r1.key, r1.bits, uintToIP(r1.key), r1.Value)
	})

	for ip, asn := range routes {
		x := findRoute(t, r, ip)
		if x == nil {
			t.Logf("Expected %d, got nil\n", asn)
			t.Fail()
		}
		if x != asn {
			t.Logf("Expected %d, got %d\n", asn, x)
			t.Fail()
		}
	}
}

func TestBitK32(t *testing.T) {
	tests := map[bittest]byte{
		bittest{0x40, 0}: 0,
		bittest{0x40, 6}: 1,
	}
	for test, expected := range tests {
		if x := bitK32(test.key, test.bit); x != expected {
			t.Logf("Expected %d for %032b (bit #%d), got %d\n", expected, test.key, test.bit, x)
			t.Fail()
		}
	}
}

func TestQueue(t *testing.T) {
	q := make(queue32, 0)
	r := New32()
	r.Value = 10

	q.Push(&node32{r, -1})
	if r1 := q.Pop(); r1.Value != 10 {
		t.Logf("Expected %d, got %d\n", 10, r.Value)
		t.Fail()
	}
	if r1 := q.Pop(); r1 != nil {
		t.Logf("Expected nil, got %d\n", r.Value)
		t.Fail()
	}
}

func TestQueue2(t *testing.T) {
	q := make(queue32, 0)
	tests := []uint32{20, 30, 40}
	for _, val := range tests {
		q.Push(&node32{&Radix32{Value: val}, -1})
	}
	for _, val := range tests {
		x := q.Pop()
		if x == nil {
			t.Logf("Expected non-nil, got nil\n")
			t.Fail()
			continue
		}
		if x.Radix32.Value != val {
			t.Logf("Expected %d, got %d\n", val, x.Radix32.Value)
			t.Fail()
		}
	}
	if x := q.Pop(); x != nil {
		t.Logf("Expected nil, got %d\n", x.Radix32.Value)
		t.Fail()
	}
	// Push and pop again, see if that works too
	for _, val := range tests {
		q.Push(&node32{&Radix32{Value: val}, -1})
	}
	for _, val := range tests {
		x := q.Pop()
		if x == nil {
			t.Logf("Expected non-nil, got nil after emptying\n")
			t.Fail()
			continue
		}
		if x.Radix32.Value != val {
			t.Logf("Expected %d, got %d\n", val, x.Radix32.Value)
			t.Fail()
		}
	}
}

func TestPanic32(t *testing.T) {
	r := New32()
	var k uint32
	for k = 0; k <= 255; k++ {
		r.Insert(k, 32, k)
	}
}

func TestPanic64(t *testing.T) {
	r := New64()
	var k uint64
	for k = 0; k <= 255; k++ {
		r.Insert(k, 64, k)
	}
}
