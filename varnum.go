/*
Copyright 2014 Tamás Gulácsi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gocilib

import (
	"strings"
)

// http://docs.oracle.com/cd/B10500_01/appdev.920/a96584/oci03typ.htm#421773
/*
Oracle stores values of the NUMBER datatype in a variable-length format. The
first byte is the exponent and is followed by 1 to 20 mantissa bytes. The
high-order bit of the exponent byte is the sign bit; it is set for positive
numbers and it is cleared for negative numbers. The lower 7 bits represent the
exponent, which is a base-100 digit with an offset of 65.

To calculate the decimal exponent, add 65 to the base-100 exponent and add
another 128 if the number is positive. If the number is negative, you do the
same, but subsequently the bits are inverted. For example, -5 has a base-100
exponent = 62 (0x3e). The decimal exponent is thus (~0x3e) -128 - 65 = 0xc1
-128 -65 = 193 -128 -65 = 0.

Each mantissa byte is a base-100 digit, in the range 1..100. For positive
numbers, the digit has 1 added to it. So, the mantissa digit for the value 5 is
6. For negative numbers, instead of adding 1, the digit is subtracted from 101.
So, the mantissa digit for the number -5 is 96 (101 - 5). Negative numbers have
a byte containing 102 appended to the data bytes. However, negative numbers
that have 20 mantissa bytes do not have the trailing 102 byte. Because the
mantissa digits are stored in base 100, each byte can represent 2 decimal
digits. The mantissa is normalized; leading zeroes are not stored.

Up to 20 data bytes can represent the mantissa. However, only 19 are guaranteed
to be accurate. The 19 data bytes, each representing a base-100 digit, yield a
maximum precision of 38 digits for an Oracle NUMBER.
*/

type OCINumber [22]byte

func (n *OCINumber) SetBytes(data []byte) {
	copy(n[:], data)
}

func (n OCINumber) String() string {
	if n[0] == 0xff { // NULL
		return ""
	}
	var txt [42]byte
	// (number) = (sign) 0.(mantissa100 * 100**(exponent100)
	length := int(n[0])
	first := n[1]
	positive := first>>7 > 0
	exp := first & ((1 << 8) - 1)
	i := 0
	if positive {
		exp = exp + 128 + 65
		for j := 0; j < length; j++ {
			digit := n[j+2] - 1
			txt[i] = '0' + digit/10
			txt[i+1] = '0' + digit%10
			i += 2
		}
	} else {
		exp = exp - 128 - 65
		txt[0] = '-'
		i = 1
		for j := 0; j < length; j++ {
			digit := 101 - n[j+2]
			txt[i] = '0' + digit/10
			txt[i+1] = '0' + digit%10
			i += 2
		}
	}
	return string(txt[:i])
}

func (n *OCINumber) SetString(txt string) {
	if txt == "" {
		n[0] = 0xff
		return
	}
	positive := true
	if txt[0] == '-' { // negative
		positive = false
		txt = txt[1:]
	}
	dot := strings.IndexByte(txt, '.')
	if dot >= 0 {
		txt = txt[:dot] + txt[dot+1:]
	} else {
		dot = len(txt)
	}
	if len(txt)%2 == 1 {
		txt = txt + "0"
	}
	if positive {
		n[1] = byte(128 + (dot+1)/2 + 64)
		j := 2
		for i := 0; i < len(txt); i += 2 {
			n[j] = (txt[i]-'0')*10 + txt[i+1] - '0' - 1
			j++
		}
	} else {
		n[1] = byte(255 - (dot+1)/2 - 64)
		j := 2
		for i := 0; i < len(txt); i += 2 {
			n[j] = 101 - ((txt[i]-'0')*10 + txt[i+1] - '0')
			j++
		}
	}
}
