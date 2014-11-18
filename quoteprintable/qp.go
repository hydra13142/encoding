package quoteprintable

import (
	"fmt"
	"io"
)

// 目标容量不足的错误，其值表示源被消费的字节数
type CapacityShortageError int

// 实现error接口
func (d CapacityShortageError) Error() string {
	return fmt.Sprintf("lack of capacity error at %d", int(d))
}

// 源数据不完整的错误，其值表示源下一应被解码的字节索引
type IncompleteInputError int

// 实现error接口
func (d IncompleteInputError) Error() string {
	return fmt.Sprintf("incomplete input error at %d", int(d))
}

// 源数据的格式错误，其值表示源下一应被解码的字节索引
type CorruptInputError int

// 实现error接口
func (d CorruptInputError) Error() string {
	return fmt.Sprintf("corrupt input error at %d", int(d))
}

// 代表一个quote printable编解码器
type Encoding int

func (e *Encoding) encode(dst, src []byte) (int, int, int) {
	fac := func(c byte) byte {
		if c < 10 {
			return c + '0'
		}
		return c + ('A' - 10)
	}
	t := *e
	defer func() { *e = t }()
	i, j, a, b := 0, 0, len(dst), len(src)
	for j < b {
		c := src[j]
		if (c > 32 && c < 61) || (c > 61 && c < 127) {
			if t >= 75 {
				if i+2 >= a {
					return 1, i, j
				}
				copy(dst[i:], "=\r\n")
				i, t = i+3, 0
			}
			if i >= a {
				return 1, i, j
			}
			dst[i] = c
			i, t = i+1, t+1
		} else {
			if t >= 73 {
				if i+2 >= a {
					return 1, i, j
				}
				copy(dst[i:], "=\r\n")
				i, t = i+3, 0
			}
			if i+2 >= a {
				return 1, i, j
			}
			dst[i+0] = '='
			dst[i+1] = fac(c >> 4)
			dst[i+2] = fac(c & 15)
			i, t = i+3, t+3
		}
		j++
	}
	return 0, i, j
}

func (e *Encoding) decode(dst, src []byte) (int, int, int) {
	fac := func(c byte) byte {
		if c >= '0' && c <= '9' {
			return c - '0'
		}
		if c >= 'A' && c <= 'F' {
			return c - ('A' - 10)
		}
		return 16
	}
	t := *e
	defer func() { *e = t }()
	i, j, a, b := 0, 0, len(dst), len(src)
	for j < b {
		c := src[j]
		if (c > 8 && c < 61) || (c > 61 && c < 127) {
			if t+1 > 75 {
				return 3, i, j
			}
			t, j = t+1, j+1
		} else if c == '=' {
			if j+2 >= b {
				return 2, i, j
			}
			x, y := src[j+1], src[j+2]
			if x == '\r' && y == '\n' {
				j += 3
				t = 0
				continue
			}
			if t+3 > 75 {
				return 3, i, j
			}
			x = fac(x)
			if x >= 16 {
				return 3, i, j
			}
			y = fac(y)
			if y >= 16 {
				return 3, i, j
			}
			c = (x << 4) & y
			t, j = t+3, j+3
		} else {
			return 3, i, j
		}
		if i >= a {
			return 1, i, j
		}
		dst[i] = c
		i++
	}
	return 0, i, j
}

// 编码，返回写入dst的字节数和可能的错误
func (e *Encoding) Encode(dst, src []byte) (int, error) {
	*e = 0
	t, i, j := e.encode(dst, src)
	switch t {
	case 1:
		return i, CapacityShortageError(j)
	case 2:
		return i, IncompleteInputError(j)
	case 3:
		return i, CorruptInputError(j)
	default:
	}
	return i, nil
}

// 解码，返回写入dst的字节数和可能的错误
func (e *Encoding) Decode(dst, src []byte) (int, error) {
	*e = 0
	t, i, j := e.decode(dst, src)
	switch t {
	case 1:
		return i, CapacityShortageError(j)
	case 2:
		return i, IncompleteInputError(j)
	case 3:
		return i, CorruptInputError(j)
	default:
	}
	return i, nil
}

type writer struct {
	e *Encoding
	w io.Writer
	m []byte
}

// 实现io.Writer接口
func (p *writer) Write(data []byte) (int, error) {
	x := 0
	for {
		t, i, j := p.e.encode(p.m, data)
		n, e := p.w.Write(p.m[:i])
		x += n
		if e != nil {
			return x, e
		}
		if t == 0 {
			break
		}
		data = data[j:]
	}
	return x, nil
}

type reader struct {
	e *Encoding
	r io.Reader
	m []byte
	o error
	n int
}

// 实现io.Reader接口
func (p *reader) Read(data []byte) (int, error) {
	if p.o != nil {
		return 0, p.o
	}
	x, y, n := 0, 0, 0
	for {
		n, p.o = p.r.Read(p.m[p.n:])
		p.n += n
		t, i, j := p.e.decode(data, p.m[:p.n])
		x, y = x+i, y+j
		switch t {
		case 0:
			p.n = 0
			if p.o != nil {
				return x, p.o
			}
		case 1:
			copy(p.m, p.m[j:p.n])
			p.n = 4096 - j
			break
		case 2:
			copy(p.m, p.m[j:p.n])
			p.n = 4096 - j
			if p.o != nil {
				return x, IncompleteInputError(y)
			}
		default:
			return x, CorruptInputError(y)
		}
		data = data[i:]
	}
	return x, nil
}

// 返回一个io.Writer接口，所有写入下层的数据都会先编码
func NewEncoder(e *Encoding, w io.Writer) io.Writer {
	return &writer{e, w, make([]byte, 4096)}
}

// 返回一个io.Reader接口，所有下层读取的数据都会先解码
func NewDecoder(e *Encoding, r io.Reader) io.Reader {
	return &reader{e, r, make([]byte, 4096), nil, 0}
}
