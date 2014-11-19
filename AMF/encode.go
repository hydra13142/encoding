package AMF

import (
	"github.com/hydra13142/encoding"
	"io"
	"time"
	"unsafe"
)

// 编码器
type Encoder struct {
	io.Writer
	Str map[string]int
}

func (this *Encoder) float(x float64) {
	s := make([]byte, 8)
	p := uintptr(unsafe.Pointer(&x))
	for i := 0; i < 8; i++ {
		s[7-i] = *(*byte)(unsafe.Pointer(p + uintptr(i)))
	}
	this.Write(s)
}

func (this *Encoder) short(i uint) {
	this.Write([]byte{byte(i >> 8), byte(i)})
}

func (this *Encoder) long(i uint) {
	this.Write([]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
}

func (this *Encoder) bytes(s string) {
	i, ok := this.Str[s]
	if ok {
		this.uint29(uint(i<<1) | 0)
		return
	} else {
		this.uint29(uint(len(s)<<1) | 1)
		this.Str[s] = len(this.Str)
		this.Write([]byte(s))
	}
}

func (this *Encoder) uint29(i uint) {
	switch {
	case i>>21 != 0:
		this.Write([]byte{byte(i>>22) | 128, byte(i>>15) | 128, byte(i>>8) | 128, byte(i)})
	case i>>14 != 0:
		this.Write([]byte{byte(i>>14) | 128, byte(i>>7) | 128, byte(i) & 127})
	case i>>7 != 0:
		this.Write([]byte{byte(i>>7) | 128, byte(i) & 127})
	default:
		this.Write([]byte{byte(i)})
	}
}

func (this *Encoder) encodeAMF0(x interface{}) error {
	if x == nil {
		this.Write([]byte{0x05})
		return nil
	}
	switch x.(type) {
	case bool:
		if x.(bool) {
			this.Write([]byte{0x01, 0x01})
		} else {
			this.Write([]byte{0x01, 0x00})
		}
	case float64:
		this.Write([]byte{0x00})
		this.float(x.(float64))
	case string:
		s := x.(string)
		l := uint(len(s))
		if l < 65535 {
			this.Write([]byte{0x02})
			this.short(l)
		} else {
			this.Write([]byte{0x0c})
			this.long(l)
		}
		this.Write([]byte(s))
	case XML:
		s := x.(XML)
		h, l := uint16(len(s)>>16), uint16(len(s)&65535)
		d := []byte{0x0f, byte(h >> 8), byte(h & 127), byte(l >> 8), byte(l & 127)}
		this.Write(d)
		this.Write([]byte(s))
	case time.Time:
		this.Write([]byte{0x0b})
		this.float(float64(x.(time.Time).UnixNano()) / 1e6)
		this.Write([]byte{0, 0})
	case []interface{}:
		d := x.([]interface{})
		l := len(d)
		this.Write([]byte{0x0a})
		this.long(uint(l))
		for i := 0; i < l; i++ {
			err := this.encodeAMF0(d[i])
			if err != nil {
				return err
			}
		}
	case []encoding.Item:
		d := x.([]encoding.Item)
		l := len(d)
		if _, ok := d[0].K.(string); !ok {
			return encoding.UnsupportType
		}
		this.Write([]byte{0x08})
		this.long(uint(l))
		for i := 0; i < l; i++ {
			s := d[i].K.(string)
			this.short(uint(len(s)))
			this.Write([]byte(s))
			err := this.encodeAMF0(d[i].V)
			if err != nil {
				return err
			}
		}
	case []encoding.Attr:
		name := ""
		d := x.([]encoding.Attr)
		l := len(d)
		if d[0].K == "$" {
			var ok bool
			name, ok = d[0].V.(string)
			if ok {
				l, d = l-1, d[1:]
			}
		}
		if name == "" {
			this.Write([]byte{0x03})
		} else {
			this.Write([]byte{0x10})
			this.Write([]byte(name))
		}
		for i := 0; i < l; i++ {
			this.short(uint(len(d[i].K)))
			this.Write([]byte(d[i].K))
			err := this.encodeAMF0(d[i].V)
			if err != nil {
				return err
			}
		}
		this.Write([]byte{0x00, 0x00, 0x09})
	default:
		return encoding.UnsupportType
	}
	return nil
}

func (this *Encoder) encodeAMF3(x interface{}) error {
	if x == nil {
		this.Write([]byte{0x01})
		return nil
	}
	switch x.(type) {
	case bool:
		if x.(bool) {
			this.Write([]byte{0x03})
		} else {
			this.Write([]byte{0x02})
		}
	case int64:
		this.Write([]byte{0x04})
		this.uint29(uint(x.(int64)))
	case float64:
		this.Write([]byte{0x05})
		this.float(x.(float64))
	case string:
		this.Write([]byte{0x06})
		s := x.(string)
		this.bytes(s)
	case []byte:
		this.Write([]byte{0x0c})
		s := x.([]byte)
		this.bytes(string(s))
	case XML:
		this.Write([]byte{0x07})
		s := x.(XML)
		this.bytes(string(s))
	case E4X:
		this.Write([]byte{0x0b})
		s := x.(E4X)
		this.bytes(string(s))
	case time.Time:
		this.Write([]byte{0x08, 0x01})
		this.float(float64(x.(time.Time).UnixNano()) / 1e6)
		this.Write([]byte{0, 0})
	case []interface{}:
		this.Write([]byte{0x09})
		d := x.([]interface{})
		l := len(d)
		this.uint29(uint(l<<1) | 1)
		this.Write([]byte{0x01})
		for i := 0; i < l; i++ {
			err := this.encodeAMF3(d[i])
			if err != nil {
				return err
			}
		}
	case []encoding.Item:
		this.Write([]byte{0x09, 0x01})
		d := x.([]encoding.Item)
		l := len(d)
		if _, ok := d[0].K.(string); !ok {
			return encoding.UnsupportType
		}
		for i := 0; i < l; i++ {
			this.bytes(d[i].K.(string))
			err := this.encodeAMF3(d[i].V)
			if err != nil {
				return err
			}
		}
		this.Write([]byte{0x01})
	case []encoding.Attr:
		this.Write([]byte{0x0a})
		name := ""
		d := x.([]encoding.Attr)
		if d[0].K == "$" {
			var ok bool
			name, ok = d[0].V.(string)
			if ok {
				d = d[1:]
			}
		}
		var a, b []encoding.Attr
		for _, t := range d {
			if t.K[0] == '@' {
				b = append(b, t)
			} else {
				a = append(a, t)
			}
		}
		if len(b) == 0 {
			this.uint29(uint(len(a)<<4) | 3)
		} else {
			this.uint29(uint(len(a)<<4) | 11)
		}
		this.bytes(name)
		for i := 0; i < len(a); i++ {
			this.bytes(a[i].K)
		}
		for i := 0; i < len(a); i++ {
			err := this.encodeAMF3(a[i].V)
			if err != nil {
				return err
			}
		}
		for i := 0; i < len(b); i++ {
			this.bytes(b[i].K)
			err := this.encodeAMF3(b[i].V)
			if err != nil {
				return err
			}
		}
	default:
		return encoding.UnsupportType
	}
	return nil
}
