package idgen

import "errors"

// FormatInt returns the string representation of i in the given base,
// for 2 <= base <= 36. The result uses the lower-case letters 'a' to 'z'
// for digit values >= 10.
func formatInt(i int64, base int) string {
	_, s := formatBits(nil, uint64(i), base, false, false)
	return s
}

const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

const host32bit = ^uint(0)>>32 == 0

const digits = "abcdefghijklmnopqrstuvwxyz"

// formatBits computes the string representation of u in the given base.
// If neg is set, u is treated as negative int64 value. If append_ is
// set, the string is appended to dst and the resulting byte slice is
// returned as the first result value; otherwise the string is returned
// as the second result value.
//
func formatBits(dst []byte, u uint64, base int, neg, append_ bool) (d []byte, s string) {
	if base < 2 || base > len(digits) {
		panic("strconv: illegal AppendInt/FormatInt base")
	}
	// 2 <= base && base <= len(digits)

	var a [64 + 1]byte // +1 for sign of 64bit value in base 2
	i := len(a)

	if neg {
		u = -u
	}

	// general case
	b := uint64(base)
	for u >= b {
		i--
		// Avoid using r = a%b in addition to q = a/b
		// since 64bit division and modulo operations
		// are calculated by runtime functions on 32bit machines.
		q := u / b
		a[i] = digits[uint(u-q*b)]
		u = q
	}
	// u < base
	i--
	a[i] = digits[uint(u)]

	// add sign, if any
	if neg {
		i--
		a[i] = '-'
	}

	if append_ {
		d = append(dst, a[i:]...)
		return
	}
	s = string(a[i:])
	return
}

func parseInt26(s string) (i int64, err error) {
	const fnParseInt = "ParseInt"

	// Empty string bad.
	if len(s) == 0 {
		return 0, errors.New("empty string")
	}

	// Convert unsigned and check range.
	un, err := ParseUint(s, 26)
	return int64(un), err
}

func ParseUint(s string, base int) (uint64, error) {
	const fnParseUint = "ParseUint"

	if len(s) == 0 {
		return 0, errors.New("invalid syntax")
	}

	bitSize := int(IntSize)

	// Cutoff is the smallest number such that cutoff*base > maxUint64.
	// Use compile-time constants for common cases.
	var cutoff uint64
	cutoff = maxUint64/uint64(base) + 1
	maxVal := uint64(1)<<uint(bitSize) - 1

	var n uint64
	for _, c := range []byte(s) {
		var d byte
		if 'a' <= c && c <= 'z' {
			d = c - 'a'
		} else {
			return 0, errors.New("invalid syntax")
		}

		if d >= byte(base) {
			return 0, errors.New("invalid syntax")
		}

		if n >= cutoff {
			// n*base overflows
			return maxVal, errors.New("invalid syntax")
		}
		n *= uint64(base)

		n1 := n + uint64(d)
		if n1 < n || n1 > maxVal {
			// n+v overflows
			return maxVal, errors.New("invalid syntax")
		}
		n = n1
	}

	return n, nil
}

const intSize = 32 << (^uint(0) >> 63)

// IntSize is the size in bits of an int or uint value.
const IntSize = intSize

const maxUint64 = (1<<64 - 1)
