package AMF

import (
	"github.com/hydra13142/encoding"
	"math"
	"time"
	"unsafe"
)

type Decoder struct {
	*encoding.Iterator
	Obj []interface{}
	Str []string
	Tra [][]string
}

func (this *Decoder) float() (x float64) {
	s := this.ReadBytes(8)
	p := uintptr(unsafe.Pointer(&x))
	for i := 0; i < 8; i++ {
		*(*byte)(unsafe.Pointer(p + uintptr(i))) = s[7-i]
	}
	return x
}

func (this *Decoder) short() uint {
	a := this.ReadByte()
	b := this.ReadByte()
	return (uint(a) << 8) + uint(b)
}

func (this *Decoder) long() uint {
	a := this.short()
	b := this.short()
	return (a << 16) + b
}

func (this *Decoder) uint29() (s uint) {
	for i := 0; i < 3; i++ {
		x := this.ReadByte()
		if x&128 == 0 {
			return (s << 7) + uint(x)
		}
		s = (s << 7) + (uint(x) & 127)
	}
	return s<<8 + uint(this.ReadByte())
}

func (this *Decoder) int29() int64 {
	return int64((int32(this.uint29()) << 3) >> 3)
}

func (this *Decoder) bytes() string {
	l := int(this.short())
	return string(this.ReadBytes(l))
}

func (this *Decoder) utf8() string {
	s := int(this.uint29())
	p, s := s&1, s>>1
	if p == 0 {
		return this.Str[s]
	}
	if s == 0 {
		return ""
	}
	str := string(this.ReadBytes(s))
	this.Str = append(this.Str, str)
	return str
}

func (this *Decoder) decodeAMF0() (it interface{}, er error) {
	defer func() {
		if e := recover(); e != nil {
			it, er = nil, e.(error)
		}
	}()
	switch this.ReadByte() {
	case 0x00: // float64
		return this.float(), nil
	case 0x01: // boolean
		return this.ReadByte() != 0, nil
	case 0x05: // null
		return nil, nil
	case 0x06: // undefined
		return encoding.Undefined{}, nil
	case 0x02: // string
		return this.bytes(), nil
	case 0x0c: // long string
		l := int(this.long())
		return string(this.ReadBytes(l)), nil
	case 0x0f: // XML document
		l := int(this.long())
		return XML(this.ReadBytes(l)), nil
	case 0x0b: // date
		i, f := math.Modf(this.float() / 1000)
		offset := time.Duration(this.short() * 3600)
		return time.Unix(int64(i), int64(f*1e9)).Add(offset), nil
	case 0x07: // reference
		l := int(this.short())
		if l != len(this.Obj) {
			return nil, encoding.SyntaxError
		}
		return this.Obj[l], nil
	case 0x0a: // strict array
		arr := []interface{}{}
		l := int(this.long())
		for i := 0; i < l; i++ {
			vlu, err := this.decodeAMF0()
			if err != nil {
				return nil, err
			}
			arr = append(arr, vlu)
		}
		this.Obj = append(this.Obj, arr)
		return arr, nil
	case 0x08: // ECMA array
		obj := make(map[string]interface{})
		l := int(this.long())
		for i := 0; i < l; i++ {
			key := this.bytes()
			vlu, err := this.decodeAMF0()
			if err != nil {
				return nil, err
			}
			obj[key] = vlu
		}
		this.Obj = append(this.Obj, obj)
		return obj, nil
	case 0x03: // object
		obj := make(map[string]interface{})
		obj["$"] = ""
		for {
			key := this.bytes()
			if key == "" {
				if this.ReadByte() != 0x09 {
					return nil, encoding.SyntaxError
				}
				break
			}
			vlu, err := this.decodeAMF0()
			if err != nil {
				return nil, err
			}
			obj[key] = vlu
		}
		this.Obj = append(this.Obj, obj)
		return obj, nil
	case 0x10: // typed object
		obj := make(map[string]interface{})
		obj["$"] = this.bytes()
		for {
			key := this.bytes()
			if key == "" {
				if this.ReadByte() != 0x09 {
					return nil, encoding.SyntaxError
				}
				break
			}
			vlu, err := this.decodeAMF0()
			if err != nil {
				return nil, err
			}
			obj[key] = vlu
		}
		this.Obj = append(this.Obj, obj)
		return obj, nil
	case 0x11: // amf3
		return this.decodeAMF3()
	}
	return nil, encoding.UnsupportType
}

func (this *Decoder) decodeAMF3() (it interface{}, er error) {
	defer func() {
		if e := recover(); e != nil {
			it, er = nil, e.(error)
		}
	}()
	switch this.ReadByte() {
	case 0x00: // undefined
		return encoding.Undefined{}, nil
	case 0x01: // null
		return nil, nil
	case 0x02: // false
		return false, nil
	case 0x03: // true
		return true, nil
	case 0x04: // int
		return this.int29(), nil
	case 0x05: // float
		return this.float(), nil
	case 0x06: // string
		return this.utf8(), nil
	case 0x0c: // byte-array
		s := int(this.uint29())
		if p, s := s&1, s>>1; p == 0 {
			return this.Obj[s], nil
		}
		str := this.ReadBytes(s)
		this.Obj = append(this.Obj, str)
		return str, nil
	case 0x07: // xml-doc
		s := int(this.uint29())
		p, s := s&1, s>>1
		if p == 0 {
			return this.Obj[s], nil
		}
		str := XML(this.ReadBytes(s))
		this.Obj = append(this.Obj, str)
		return str, nil
	case 0x0b: // xml
		s := int(this.uint29())
		p, s := s&1, s>>1
		if p == 0 {
			return this.Obj[s], nil
		}
		str := E4X(this.ReadBytes(s))
		this.Obj = append(this.Obj, str)
		return str, nil
	case 0x08: // date
		s := this.uint29()
		p, s := s&1, s>>1
		if p == 0 {
			return this.Obj[s], nil
		}
		i, f := math.Modf(this.float() / 1000)
		offset := time.Duration(this.short() * 3600)
		return time.Unix(int64(i), int64(f*1e9)).Add(offset), nil
	case 0x09: // array
		s := this.uint29()
		p, s := s&1, s>>1
		if p == 0 {
			return this.Obj[s], nil
		}
		if s == 0 {
			arr := make(map[string]interface{})
			for {
				key := this.utf8()
				if key == "" {
					break
				}
				vlu, err := this.decodeAMF3()
				if err != nil {
					return nil, err
				}
				arr[key] = vlu
			}
			this.Obj = append(this.Obj, arr)
			return arr, nil
		} else if this.utf8() == "" {
			arr := make([]interface{}, 0)
			for i := 0; i < int(s); i++ {
				vlu, err := this.decodeAMF3()
				if err != nil {
					return nil, err
				}
				arr = append(arr, vlu)
			}
			this.Obj = append(this.Obj, arr)
			return arr, nil
		}
		return nil, encoding.UnsupportType
	case 0x0a: // object
		t := int(this.uint29())
		if t&1 == 0 {
			return this.Obj[t>>1], nil
		}
		if t&2 == 0 {
			tra := this.Tra[t>>2]
			obj := make(map[string]interface{})
			obj["$"] = tra[0]
			for i := 1; i < len(tra); i++ {
				vlu, err := this.decodeAMF3()
				if err != nil {
					return nil, err
				}
				obj[tra[i]] = vlu
			}
			this.Obj = append(this.Obj, obj)
			return obj, nil
		}
		if t&4 == 0 {
			tra := []string{}
			l := (t >> 4) + 1
			for i := 0; i < l; i++ {
				tra = append(tra, this.utf8())
			}
			obj := make(map[string]interface{})
			obj["$"] = tra[0]
			for i := 1; i < l; i++ {
				vlu, err := this.decodeAMF3()
				if err != nil {
					return nil, err
				}
				obj[tra[i]] = vlu
			}
			this.Tra = append(this.Tra, tra)
			if t&8 == 1 {
				for {
					key := this.utf8()
					if key == "" {
						break
					}
					vlu, err := this.decodeAMF3()
					if err != nil {
						return nil, err
					}
					obj["@"+key] = vlu
				}
			}
			this.Obj = append(this.Obj, obj)
			return obj, nil
		}
		return nil, encoding.UnsupportType
	}
	return nil, encoding.UnsupportType
}
