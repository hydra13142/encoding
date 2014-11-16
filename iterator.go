package encoding

import "io"

// 用于从Reader接口获取数据，采用panic替代返回错误值
type Iterator struct {
	r    io.Reader
	m    []byte
	a, b int
	o    error
}

// 创建并返回一个Iterator
func NewIterator(r io.Reader) *Iterator {
	return &Iterator{r, make([]byte, 4096), 0, 0, nil}
}

// 获取一个字节
func (p *Iterator) ReadByte() byte {
	var c byte
	for {
		if p.a < p.b {
			c = p.m[p.a]
			p.a++
			break
		}
		if p.o != nil {
			panic(p.o)
		}
		p.b, p.o = p.r.Read(p.m)
		p.a = 0
	}
	return c
}

// 退回一个字节
func (p *Iterator) UnreadByte() {
	p.a--
}

// 读取一定量字节
func (p *Iterator) ReadBytes(n int) []byte {
	u := make([]byte, n)
	for v := u; ; {
		if p.a+n <= p.b {
			copy(v, p.m[p.a:p.a+n])
			p.a += n
			break
		}
		x := copy(v, p.m[p.a:p.b])
		n, v = n-x, v[x:]
		if p.o != nil {
			panic(p.o)
		}
		p.b, p.o = p.r.Read(p.m)
		p.a = 0
	}
	return u
}
