package bencode

import (
	"fmt"
	"io"
	"github.com/hydra13142/encoding"
)

// bencode编码器
type Encoder struct {
	w io.Writer
}

// bencode解码器
type Decoder struct {
	r *encoding.Iterator
}

// 编码格式错误
var SyntaxError = fmt.Errorf("syntax error")

func (p *Encoder) encode(x interface{}) error {
	var e error
	switch x.(type) {
	case int64:
		if _, e = fmt.Fprintf(p.w, "i%de", x); e != nil {
			return e
		}
	case string:
		if _, e = fmt.Fprintf(p.w, "%d:%s", len(x.(string)), x); e != nil {
			return e
		}
	case []interface{}:
		if _, e = p.w.Write([]byte{'l'}); e != nil {
			return e
		}
		for _, v := range x.([]interface{}) {
			if e = p.encode(v); e != nil {
				return e
			}
		}
		if _, e = p.w.Write([]byte{'e'}); e != nil {
			return e
		}
	case []encoding.Attr:
		if _, e = p.w.Write([]byte{'d'}); e != nil {
			return e
		}
		for _, u := range x.([]encoding.Attr) {
			if _, e = fmt.Fprintf(p.w, "%d:%s", len(u.K), u.K); e != nil {
				return e
			}
			if e = p.encode(u.V); e != nil {
				return e
			}
		}
		if _, e = p.w.Write([]byte{'e'}); e != nil {
			return e
		}
	}
	return nil
}

func (p *Decoder) number() int {
	n, t := 0, true
	c := p.r.ReadByte()
	if c == '-' {
		t, c = false, p.r.ReadByte()
	}
	for {
		if c < '0' || c > '9' {
			p.r.UnreadByte()
			break
		}
		n = n*10 + int(c-'0')
		c = p.r.ReadByte()
	}
	if t {
		return n
	}
	return -n
}

func (p *Decoder) decode() (obj interface{}, err error) {
	var (
		c byte
		n int
		s string
		t int
	)
	defer func() {
		if e := recover(); e != nil {
			if t == 0 {
				obj, err = nil, e.(error)
			} else {
				obj, err = nil, SyntaxError
			}
		}
	}()
	c = p.r.ReadByte()
	t = 1
	switch c {
	case 'i':
		n = p.number()
		c = p.r.ReadByte()
		if c != 'e' {
			return nil, SyntaxError
		}
		return int64(n), nil
	case 'l':
		l := []interface{}{}
		for {
			c = p.r.ReadByte()
			if c == 'e' {
				return l, nil
			}
			p.r.UnreadByte()
			i, e := p.decode()
			if e != nil {
				return nil, e
			}
			l = append(l, i)
		}
	case 'd':
		d := map[string]interface{}{}
		for {
			c = p.r.ReadByte()
			if c == 'e' {
				return d, nil
			}
			if c < '0' || c > '9' {
				return nil, SyntaxError
			}
			p.r.UnreadByte()
			n = p.number()
			c = p.r.ReadByte()
			if c != ':' {
				return nil, SyntaxError
			}
			s = string(p.r.ReadBytes(n))
			i, e := p.decode()
			if e != nil {
				return nil, e
			}
			d[s] = i
		}
	default:
		if c < '0' || c > '9' {
			return nil, SyntaxError
		}
		p.r.UnreadByte()
		n = p.number()
		c = p.r.ReadByte()
		if c != ':' {
			return nil, SyntaxError
		}
		s = string(p.r.ReadBytes(n))
		return s, nil
	}
}
