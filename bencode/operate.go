package bencode

import (
	"errors"
	"github.com/hydra13142/encoding"
	"io"
	"reflect"
)

var translator = encoding.Translator{"bencode", make(map[reflect.Type][]encoding.Label), nil}

// 解码的目标参数必须是指针
var TypeError = errors.New("need point type")

// 创建编码器
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// 创建解码器
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{encoding.NewIterator(r)}
}

// 编码对象后写入下层
func (this *Encoder) Encode(x interface{}) error {
	c, e := translator.Encode(reflect.ValueOf(x))
	if e != nil {
		return e
	}
	return this.encode(c)
}

// 读取并解码后填充对象
func (this *Decoder) Decode(x interface{}) error {
	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Invalid {
		_, e := this.decode()
		if e != nil {
			return e
		}
		return nil
	} else if v.Kind() != reflect.Ptr {
		return TypeError
	}
	c, e := this.decode()
	if e != nil {
		return e
	}
	return translator.Decode(v.Elem(), c)
}
