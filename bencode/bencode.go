package bencode

import (
	"fmt"
	"io"
	"github.com/hydra13142/encoding"
)

// bencode编码器
type Encoder struct {
	io.Writer
}

// bencode解码器
type Decoder struct {
	*encoding.Iterator
}

func (p *Encoder) encode(x interface{}) error {
	var e error
	switch x.(type) {
	case int64:
		if _, e = fmt.Fprintf(p.Writer, "i%de", x); e != nil {
			return e
		}
	case string:
		if _, e = fmt.Fprintf(p.Writer, "%d:%s", len(x.(string)), x); e != nil {
			return e
		}
	case []interface{}:
		if _, e = p.Writer.Write([]byte{'l'}); e != nil {
			return e
		}
		for _, v := range x.([]interface{}) {
			if e = p.encode(v); e != nil {
				return e
			}
		}
		if _, e = p.Writer.Write([]byte{'e'}); e != nil {
			return e
		}
	case []encoding.Item:
		if _, e = p.Writer.Write([]byte{'d'}); e != nil {
			return e
		}
		for _, u := range x.([]encoding.Item) {
			k, ok := u.K.(string)
			if !ok {
				return encoding.UnsupportType
			}
			if _, e = fmt.Fprintf(p.Writer, "%d:%s", len(k), k); e != nil {
				return e
			}
			if e = p.encode(u.V); e != nil {
				return e
			}
		}
		if _, e = p.Writer.Write([]byte{'e'}); e != nil {
			return e
		}
	case []encoding.Attr:
		if _, e = p.Writer.Write([]byte{'d'}); e != nil {
			return e
		}
		for _, u := range x.([]encoding.Attr) {
			if _, e = fmt.Fprintf(p.Writer, "%d:%s", len(u.K), u.K); e != nil {
				return e
			}
			if e = p.encode(u.V); e != nil {
				return e
			}
		}
		if _, e = p.Writer.Write([]byte{'e'}); e != nil {
			return e
		}
	default:
		return encoding.UnsupportType
	}
	return nil
}

func (p *Decoder) number() int {
	n, t := 0, true
	c := p.Iterator.ReadByte()
	if c == '-' {
		t, c = false, p.Iterator.ReadByte()
	}
	for {
		if c < '0' || c > '9' {
			p.Iterator.UnreadByte()
			break
		}
		n = n*10 + int(c-'0')
		c = p.Iterator.ReadByte()
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
				obj, err = nil, encoding.SyntaxError
			}
		}
	}()
	c = p.Iterator.ReadByte()
	t = 1
	switch c {
	case 'i':
		n = p.number()
		c = p.Iterator.ReadByte()
		if c != 'e' {
			return nil, encoding.SyntaxError
		}
		return int64(n), nil
	case 'l':
		l := []interface{}{}
		for {
			c = p.Iterator.ReadByte()
			if c == 'e' {
				return l, nil
			}
			p.Iterator.UnreadByte()
			i, e := p.decode()
			if e != nil {
				return nil, e
			}
			l = append(l, i)
		}
	case 'd':
		d := map[string]interface{}{}
		for {
			c = p.Iterator.ReadByte()
			if c == 'e' {
				return d, nil
			}
			if c < '0' || c > '9' {
				return nil, encoding.SyntaxError
			}
			p.Iterator.UnreadByte()
			n = p.number()
			c = p.Iterator.ReadByte()
			if c != ':' {
				return nil, encoding.SyntaxError
			}
			s = string(p.Iterator.ReadBytes(n))
			i, e := p.decode()
			if e != nil {
				return nil, e
			}
			d[s] = i
		}
	default:
		if c < '0' || c > '9' {
			return nil, encoding.SyntaxError
		}
		p.Iterator.UnreadByte()
		n = p.number()
		c = p.Iterator.ReadByte()
		if c != ':' {
			return nil, encoding.SyntaxError
		}
		s = string(p.Iterator.ReadBytes(n))
		return s, nil
	}
}
