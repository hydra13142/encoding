package AMF

import (
	"errors"
	"github.com/hydra13142/encoding"
	"io"
	"reflect"
	"time"
)

// 解码的目标填充参数必须是指针
var TypeError = errors.New("need point type")

var rawtype = map[reflect.Type]struct{}{
	reflect.TypeOf(XML("")):         struct{}{},
	reflect.TypeOf(E4X("")):         struct{}{},
	reflect.TypeOf(time.Unix(0, 0)): struct{}{},
}

var translator = encoding.Translator{"amf", make(map[reflect.Type][]encoding.Label), rawtype}

// 创建解码器
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{Iterator: encoding.NewIterator(r)}
}

// 创建编码器
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w, make(map[string]int)}
}

// 解码并填充
func (this *Decoder) Decode(x interface{}) error {
	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Invalid {
		_, e := this.decodeAMF0()
		if e != nil {
			return e
		}
		return nil
	} else if v.Kind() != reflect.Ptr {
		return TypeError
	}
	u, e := this.decodeAMF0()
	if e != nil {
		return e
	}
	return translator.Decode(v.Elem(), u)
}

// 编码并写入
func (this *Encoder) Encode(x interface{}, s string) error {
	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	u, e := translator.Encode(v)
	if e != nil {
		return e
	}
	switch s {
	case "0", "amf0", "AMF0":
		return this.encodeAMF0(u)
	case "3", "amf3", "AMF3":
		this.Write([]byte{0x11})
		return this.encodeAMF3(u)
	}
	return errors.New("codec must be AMF0 or AMF3")
}
