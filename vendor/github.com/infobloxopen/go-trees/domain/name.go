// Package domain provide functions to parse and handle domain names and labels.
package domain

import "errors"

var (
	// ErrCompressedName is the error returned by WireGet when domain label exceeds 63 bytes.
	ErrCompressedName = errors.New("can't handle compressed domain name")
	// ErrLabelTooLong is the error returned by WireGet when last domain label length doesn't
	// fit whole domain name length.
	ErrLabelTooLong = errors.New("label too long")
	// ErrEmptyLabel means that label of zero length met in the middle of domain name.
	ErrEmptyLabel = errors.New("empty label")
	// ErrNameTooLong is the error returned when overall domain name length exeeds 256 bytes.
	ErrNameTooLong = errors.New("domain name too long")
)

// Split slices its argument into labels. It assumes s is a name in human readable format.
func Split(s string) []Label {
	dn := make([]Label, getLabelsCount(s))
	if len(dn) > 0 {
		end := len(dn) - 1
		start := 0
		for i := range dn {
			label, p := MakeLabel(s[start:])
			start += p + 1
			dn[end-i] = label
		}
	}

	return dn
}

func getLabelsCount(s string) int {
	labels := 0
	start := 0
	for {
		size, p := GetFirstLabelSize(s[start:])
		start += p + 1
		if start >= len(s) {
			if size > 0 {
				labels++
			}

			break
		}

		labels++
	}

	return labels
}

// WireNameLower is a type to store domain name in "wire" format as described in RFC-1035 section "3.1. Name space definitions" with all lowercase ASCII letters.
type WireNameLower []byte

// MakeWireDomainNameLower creates lowercase "wire" representation of given domain name.
func MakeWireDomainNameLower(s string) (WireNameLower, error) {
	out := WireNameLower{}
	start := 0
	for {
		label, p := MakeLabel(s[start:])
		if len(label) > 63 {
			return nil, ErrLabelTooLong
		}

		start += p + 1
		if start >= len(s) {
			if len(label) > 0 {
				out = append(out, byte(len(label)))
				out = append(out, label...)
				if len(out) > 255 {
					return nil, ErrNameTooLong
				}
			}

			break
		}

		if len(label) <= 0 {
			return nil, ErrEmptyLabel
		}

		out = append(out, byte(len(label)))
		out = append(out, label...)
		if len(out) > 255 {
			return nil, ErrNameTooLong
		}
	}

	return append(out, 0), nil
}

// ToLowerWireDomainName converts "wire" domain name to lowercase.
func ToLowerWireDomainName(d []byte) (WireNameLower, error) {
	if len(d) > 256 {
		return nil, ErrNameTooLong
	}

	out := make(WireNameLower, len(d))
	ll := 0
	for i, c := range d {
		if ll > 0 {
			ll--

			if c >= 'A' && c <= 'Z' {
				c += 0x20
			}
		} else {
			ll = int(c)
			if ll <= 0 && i != len(d)-1 {
				return nil, ErrEmptyLabel
			}

			if ll > 63 {
				return nil, ErrCompressedName
			}
		}

		out[i] = c
	}

	if out[len(out)-1] != 0 {
		return nil, ErrLabelTooLong
	}

	return out, nil
}

// String returns domain name in human readable format.
func (d WireNameLower) String() string {
	out := ""
	start := 0
	for start < len(d) {
		ll := int(d[start])

		start++
		if start >= len(d) {
			if ll > 0 {
				out += "."
			}

			return out
		}

		if ll > 0 {
			label := Label(d[start : start+ll]).String()
			if len(out) > 0 {
				out += "." + label
			} else {
				out = label
			}

			start += ll
		}
	}

	return out
}

// WireSplitCallback function splits given name into labels and calls function for each label.
func WireSplitCallback(name WireNameLower, f func(label []byte) bool) error {
	if len(name) > 256 {
		return ErrNameTooLong
	}

	if len(name) > 0 {
		var lPos [256]int
		labels := 0
		idx := 0
		max := 0
		for {
			ll := int(name[idx])
			if ll <= 0 {
				if idx != len(name)-1 {
					return ErrEmptyLabel
				}

				break
			}

			if ll > 63 {
				return ErrCompressedName
			}

			if idx+ll+1 > len(name) {
				return ErrLabelTooLong
			}

			if ll > max {
				max = ll
			}

			lPos[labels] = idx
			labels++
			idx += ll + 1
		}

		for labels > 0 {
			labels--
			idx := lPos[labels]
			ll := int(name[idx])
			start := idx + 1
			end := start + ll
			if !f(name[start:end]) {
				break
			}
		}
	}

	return nil
}
