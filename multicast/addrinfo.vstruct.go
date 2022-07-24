package multicast

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

type _ = strings.Builder
type _ = unsafe.Pointer

var _ = math.Float32frombits
var _ = math.Float64frombits
var _ = strconv.FormatInt
var _ = strconv.FormatUint
var _ = strconv.FormatFloat
var _ = fmt.Sprint

type AddrInfo []byte

func (s AddrInfo) Port() uint16 {
	_ = s[1]
	var __v uint16 = uint16(s[0]) |
		uint16(s[1])<<8
	return uint16(__v)
}

func (s AddrInfo) Ip() string {
	_ = s[9]
	var __off0 uint64 = 10
	var __off1 uint64 = uint64(s[2]) |
		uint64(s[3])<<8 |
		uint64(s[4])<<16 |
		uint64(s[5])<<24 |
		uint64(s[6])<<32 |
		uint64(s[7])<<40 |
		uint64(s[8])<<48 |
		uint64(s[9])<<56
	var __v = s[__off0:__off1]

	return *(*string)(unsafe.Pointer(&__v))
}

func (s AddrInfo) Vstruct_Validate() bool {
	if len(s) < 10 {
		return false
	}

	_ = s[9]

	var __off0 uint64 = 10
	var __off1 uint64 = uint64(s[2]) |
		uint64(s[3])<<8 |
		uint64(s[4])<<16 |
		uint64(s[5])<<24 |
		uint64(s[6])<<32 |
		uint64(s[7])<<40 |
		uint64(s[8])<<48 |
		uint64(s[9])<<56
	var __off2 uint64 = uint64(len(s))
	return __off0 <= __off1 && __off1 <= __off2
}

func (s AddrInfo) String() string {
	if !s.Vstruct_Validate() {
		return "AddrInfo (invalid)"
	}
	var __b strings.Builder
	__b.WriteString("AddrInfo {")
	__b.WriteString("Port: ")
	__b.WriteString(strconv.FormatUint(uint64(s.Port()), 10))
	__b.WriteString(", ")
	__b.WriteString("Ip: ")
	__b.WriteString(strconv.Quote(s.Ip()))
	__b.WriteString("}")
	return __b.String()
}

func Serialize_AddrInfo(dst AddrInfo, Port uint16, Ip string) AddrInfo {
	_ = dst[9]
	dst[0] = byte(Port)
	dst[1] = byte(Port >> 8)

	var __index = uint64(10)
	__tmp_1 := uint64(len(Ip)) + __index
	dst[2] = byte(__tmp_1)
	dst[3] = byte(__tmp_1 >> 8)
	dst[4] = byte(__tmp_1 >> 16)
	dst[5] = byte(__tmp_1 >> 24)
	dst[6] = byte(__tmp_1 >> 32)
	dst[7] = byte(__tmp_1 >> 40)
	dst[8] = byte(__tmp_1 >> 48)
	dst[9] = byte(__tmp_1 >> 56)
	copy(dst[__index:__tmp_1], Ip)
	return dst
}

func New_AddrInfo(Port uint16, Ip string) AddrInfo {
	var __vstruct__size = 10 + len(Ip)
	var __vstruct__buf = make(AddrInfo, __vstruct__size)
	__vstruct__buf = Serialize_AddrInfo(__vstruct__buf, Port, Ip)
	return __vstruct__buf
}
