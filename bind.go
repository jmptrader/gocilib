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

// #cgo LDFLAGS: -locilib
// #include "ocilib.h"
//
// const int sof_OCI_DateP = sizeof(OCI_Date*);
// const int sof_OCI_IntervalP = sizeof(OCI_Interval*);
import "C"

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"gopkg.in/inconshreveable/log15.v2"
)

func (stmt *Statement) BindPos(pos int, arg driver.Value) error {
	return stmt.BindName(":"+strconv.Itoa(pos), arg)
}

func (stmt *Statement) BindName(name string, value driver.Value) error {
	h, nm, ok := stmt.handle, C.CString(name), C.int(C.FALSE)
	Log.Debug("BindName", "name", name,
		"type", log15.Lazy{func() string { return fmt.Sprintf("%T", value) }},
		"value", log15.Lazy{func() string { return fmt.Sprintf("%#v", value) }},
	)
Outer:
	switch x := value.(type) {
	case int16: // short
		ok = C.OCI_BindShort(h, nm, (*C.short)(unsafe.Pointer(&x)))
	case *int16: // short
		ok = C.OCI_BindShort(h, nm, (*C.short)(unsafe.Pointer(x)))
	case []int16:
		ok = C.OCI_BindArrayOfShorts(h, nm, (*C.short)(unsafe.Pointer(&x[0])), C.uint(len(x)))
	case uint16: // unsigned short
		ok = C.OCI_BindUnsignedShort(h, nm, (*C.ushort)(unsafe.Pointer(&x)))
	case *uint16: // unsigned short
		ok = C.OCI_BindUnsignedShort(h, nm, (*C.ushort)(unsafe.Pointer(x)))
	case []uint16:
		ok = C.OCI_BindArrayOfUnsignedShorts(h, nm, (*C.ushort)(unsafe.Pointer(&x[0])), C.uint(len(x)))
	case int: // int
		ok = C.OCI_BindInt(h, nm, (*C.int)(unsafe.Pointer(&x)))
	case []int:
		ok = C.OCI_BindArrayOfInts(h, nm, (*C.int)(unsafe.Pointer(&x[0])), C.uint(len(x)))
	case uint: // int
		ok = C.OCI_BindUnsignedInt(h, nm, (*C.uint)(unsafe.Pointer(&x)))
	case *uint: // int
		ok = C.OCI_BindUnsignedInt(h, nm, (*C.uint)(unsafe.Pointer(x)))
	case []uint:
		ok = C.OCI_BindArrayOfUnsignedInts(h, nm, (*C.uint)(unsafe.Pointer(&x[0])), C.uint(len(x)))
	case int64:
		ok = C.OCI_BindBigInt(h, nm, (*C.big_int)(unsafe.Pointer(&x)))
	case *int64:
		ok = C.OCI_BindBigInt(h, nm, (*C.big_int)(unsafe.Pointer(x)))
	case []int64:
		ok = C.OCI_BindArrayOfBigInts(h, nm, (*C.big_int)(unsafe.Pointer(&x[0])), C.uint(len(x)))
	case uint64:
		ok = C.OCI_BindUnsignedBigInt(h, nm, (*C.big_uint)(unsafe.Pointer(&x)))
	case *uint64:
		ok = C.OCI_BindUnsignedBigInt(h, nm, (*C.big_uint)(unsafe.Pointer(x)))
	case []uint64:
		ok = C.OCI_BindArrayOfUnsignedBigInts(h, nm, (*C.big_uint)(unsafe.Pointer(&x[0])), C.uint(len(x)))
	case string:
		m := len(x)
		if m == 0 {
			m = 32766
		}
		y := make([]byte, m+1)
		if m > 0 {
			copy(y, []byte(x))
		}
		y[m] = 0 // trailing 0
		ok = C.OCI_BindString(h, nm, (*C.dtext)(unsafe.Pointer(&y[0])), C.uint(len(x)))
	case StringVar:
		ok = C.OCI_BindString(h, nm, (*C.dtext)(unsafe.Pointer(&x.data[0])), C.uint(len(x.data)))
	case *StringVar:
		ok = C.OCI_BindString(h, nm, (*C.dtext)(unsafe.Pointer(&x.data[0])), C.uint(len(x.data)))
	case *string:
		y := make([]byte, 32767)
		copy(y, []byte(*x))
		m := len(*x)
		y[m] = 0 // trailing 0
		*x = *(*string)(unsafe.Pointer(&reflect.StringHeader{Data: uintptr(unsafe.Pointer(&y[0])), Len: len(y)}))
		ok = C.OCI_BindString(h, nm, (*C.dtext)(unsafe.Pointer(&y[0])), C.uint(len(y)))
	case []string:
		m := 0
		for _, s := range x {
			if len(s) > m {
				m = len(s)
			}
		}
		if m == 0 {
			m = 32767
		}
		y := make([]byte, m*len(x))
		for i, s := range x {
			copy(y[i*m:(i+1)*m], []byte(s))
		}
		ok = C.OCI_BindArrayOfStrings(h, nm, (*C.dtext)(unsafe.Pointer(&y[0])), C.uint(m), C.uint(len(x)))
	case []byte:
		ok = C.OCI_BindRaw(h, nm, unsafe.Pointer(&x[0]), C.uint(cap(x)))
	/*case *[]byte:
	if len(*x) == 0 {
		*x = (*x)[:cap(*x)]
	}
	ok = C.OCI_BindString(h, nm, (*C.char)(unsafe.Pointer(&(*x)[0])), C.uint(cap(*x)))*/
	case [][]byte:
		m := 0
		for _, b := range x {
			if len(b) > m {
				m = len(b)
			}
		}
		if m == 0 {
			m = 32767
		}
		y := make([]byte, m*len(x))
		for i, b := range x {
			copy(y[i*m:(i+1)*m], b)
		}
		ok = C.OCI_BindArrayOfRaws(h, nm, unsafe.Pointer(&y[0]), C.uint(m), C.uint(len(x)))
	case float32:
		ok = C.OCI_BindFloat(h, nm, (*C.float)(&x))
	case *float32:
		ok = C.OCI_BindFloat(h, nm, (*C.float)(x))
	case []float32:
		ok = C.OCI_BindArrayOfFloats(h, nm, (*C.float)(&x[0]), C.uint(len(x)))
	case float64:
		ok = C.OCI_BindDouble(h, nm, (*C.double)(&x))
	case *float64:
		ok = C.OCI_BindDouble(h, nm, (*C.double)(x))
	case []float64:
		ok = C.OCI_BindArrayOfDoubles(h, nm, (*C.double)(&x[0]), C.uint(len(x)))
	case time.Time:
		od := C.OCI_DateCreate(C.OCI_StatementGetConnection(stmt.handle))
		y, m, d := x.Date()
		H, M, S := x.Clock()
		if C.OCI_DateSetDateTime(od, C.int(y), C.int(m), C.int(d), C.int(H), C.int(M), C.int(S)) != C.TRUE {
			break
		}
		ok = C.OCI_BindDate(h, nm, od)
	case []time.Time:
		od := C.OCI_DateArrayCreate(C.OCI_StatementGetConnection(stmt.handle), C.uint(len(x)))
		sof_OCI_DateP := C.int(C.sof_OCI_DateP)
		for i, t := range x {
			y, m, d := t.Date()
			H, M, S := t.Clock()
			if C.OCI_DateSetDateTime(
				(*C.OCI_Date)(unsafe.Pointer(
					uintptr(unsafe.Pointer(od))+uintptr(sof_OCI_DateP*C.int(i))),
				),
				C.int(y), C.int(m), C.int(d), C.int(H), C.int(M), C.int(S),
			) != C.TRUE {
				break Outer
			}
		}
		ok = C.OCI_BindArrayOfDates(h, nm, od, C.uint(len(x)))
	case time.Duration:
		oi := C.OCI_IntervalCreate(C.OCI_StatementGetConnection(stmt.handle), C.OCI_INTERVAL_DS)
		d, H, M, S, ms := durationAsDays(x)
		if C.OCI_IntervalSetDaySecond(oi, C.int(d), C.int(H), C.int(M), C.int(S), C.int(ms/100)) != C.TRUE {
			break
		}
		ok = C.OCI_BindInterval(h, nm, oi)
	case []time.Duration:
		oi := C.OCI_IntervalArrayCreate(C.OCI_StatementGetConnection(stmt.handle), C.OCI_INTERVAL_DS, C.uint(len(x)))
		sof_OCI_IntervalP := C.int(C.sof_OCI_IntervalP)
		for i, t := range x {
			d, H, M, S, ms := durationAsDays(t)
			if C.OCI_IntervalSetDaySecond(
				(*C.OCI_Interval)(unsafe.Pointer(
					uintptr(unsafe.Pointer(oi))+uintptr(sof_OCI_IntervalP*C.int(i)))),
				C.int(d), C.int(H), C.int(M), C.int(S), C.int(ms/100),
			) != C.TRUE {
				break Outer
			}
		}
		ok = C.OCI_BindArrayOfIntervals(h, nm, oi, C.OCI_INTERVAL_DS, C.uint(len(x)))
	case LOB:
		ok = C.OCI_BindLob(h, nm, x.handle)
	case []LOB:
		if len(x) > 0 {
			lo := make([]*C.OCI_Lob, len(x))
			for i := range x {
				lo[i] = x[i].handle
			}
			ok = C.OCI_BindArrayOfLobs(h, nm, (**C.OCI_Lob)(unsafe.Pointer(&lo[0])), x[0].Type(), C.uint(len(x)))
		}
	case File:
		ok = C.OCI_BindFile(h, nm, x.handle)
	case []File:
		if len(x) > 0 {
			fi := make([]*C.OCI_File, len(x))
			for i := range x {
				fi[i] = x[i].handle
			}
			ok = C.OCI_BindArrayOfFiles(h, nm, (**C.OCI_File)(unsafe.Pointer(&fi[0])), x[0].Type(), C.uint(len(x)))
		}
	case Object:
		ok = C.OCI_BindObject(h, nm, x.handle)
	case []Object:
		if len(x) > 0 {
			ob := make([]*C.OCI_Object, len(x))
			for i := range x {
				ob[i] = x[i].handle
			}
			ok = C.OCI_BindArrayOfObjects(h, nm, (**C.OCI_Object)(unsafe.Pointer(&ob[0])), x[0].Type(), C.uint(len(x)))
		}
	case Coll:
		ok = C.OCI_BindColl(h, nm, x.handle)
	case []Coll:
		if len(x) > 0 {
			co := make([]*C.OCI_Coll, len(x))
			for i := range x {
				co[i] = x[i].handle
			}
			ok = C.OCI_BindArrayOfColls(h, nm, (**C.OCI_Coll)(unsafe.Pointer(&co[0])), x[0].Type(), C.uint(len(x)))
		}
	case Ref:
		ok = C.OCI_BindRef(h, nm, x.handle)
	case []Ref:
		if len(x) > 0 {
			re := make([]*C.OCI_Ref, len(x))
			for i := range x {
				re[i] = x[i].handle
			}
			ok = C.OCI_BindArrayOfRefs(h, nm, (**C.OCI_Ref)(unsafe.Pointer(&re[0])), x[0].Type(), C.uint(len(x)))
		}
	case Statement:
		ok = C.OCI_BindStatement(h, nm, x.handle)
	case Long:
		ok = C.OCI_BindLong(h, nm, x.handle, x.Len())
	default:
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			return stmt.BindName(name, v.Elem().Interface())
		}
		return fmt.Errorf("BindName(%s): unknown type %T", name, value)
	}
	if ok != C.TRUE {
		return fmt.Errorf("BindName(%s): %v", name, getLastErr())
	}
	return nil
}

func durationAsDays(d time.Duration) (days, hours, minutes, seconds, milliseconds int) {
	ns := d.Nanoseconds()
	days = int(ns / int64(time.Hour) / 24)
	ns -= int64(days) * int64(time.Hour) * 240
	hours = int(ns / int64(time.Hour))
	ns -= int64(hours) * int64(time.Hour)
	minutes = int(ns / int64(time.Minute))
	ns -= int64(minutes) * int64(time.Minute)
	seconds = int(ns / int64(time.Second))
	ns -= int64(seconds) * int64(time.Second)
	milliseconds = int(ns / int64(time.Millisecond))
	return
}
