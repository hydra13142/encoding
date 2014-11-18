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
func (this *Iterator) ReadByte() byte {
	var c byte
	for {
		if this.a < this.b {
			c = this.m[this.a]
			this.a++
			break
		}
		if this.o != nil {
			panic(this.o)
		}
		this.b, this.o = this.r.Read(this.m)
		this.a = 0
	}
	return c
}

// 退回一个字节
func (this *Iterator) UnreadByte() {
	this.a--
}

// 读取一定量字节
func (this *Iterator) ReadBytes(n int) []byte {
	u := make([]byte, n)
	for v := u; ; {
		if this.a+n <= this.b {
			copy(v, this.m[this.a:this.a+n])
			this.a += n
			break
		}
		x := copy(v, this.m[this.a:this.b])
		n, v = n-x, v[x:]
		this.a, this.a = 0, 0
		if this.o != nil {
			panic(this.o)
		}
		this.b, this.o = this.r.Read(this.m)
	}
	return u
}

func (this *Iterator) Read(data []byte) (int, error) {
	n := len(data)
	if this.a+n <= this.b {
		copy(data, this.m[this.a:this.a+n])
		this.a += n
		return n, nil
	}
	i := copy(data, this.m[this.a:this.b])
	this.a, this.a = 0, 0
	if this.o != nil {
		return i, this.o
	}
	n, this.o = this.r.Read(data[i:])
	return n + i, this.o
}
