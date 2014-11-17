package encoding

import (
	"errors"
	"reflect"
	"strings"
)

var (
	// 编码时可能出现的错误
	UnsupportType = errors.New("unsupport type")
	// 解码时可能出现的错误
	UnmatchedType = errors.New("unmatched type")
	// 编码格式错误
	SyntaxError = errors.New("syntax error")
)

// 表示映射的一个键值对
type Item struct {
	K interface{}
	V interface{}
}

// 表示键为字符串的映射的一个键值对，或者结构体的某个字段标识符及其值
type Attr struct {
	K string
	V interface{}
}

// 实现中间数据与具体类型编解码
type Translator struct {
	Name string
	Tag  map[reflect.Type][]Label
	Raw  map[reflect.Type]struct{}
}

// 获取某结构体类型的标签信息
func (this *Translator) GetLabel(x reflect.Type) []Label {
	p := make([]Label, 0, 0)
	for i, l := 0, x.NumField(); i < l; i++ {
		f := x.Field(i)
		if f.Name[0] < 'A' || f.Name[0] > 'Z' {
			continue
		}
		tag := f.Tag.Get(string(this.Name))
		if tag == "-" {
			continue
		}
		if tag == "" {
			p = append(p, Label{i, []string{f.Name}})
		}
		y := make([]string, 0, 0)
		for _, x := range strings.Split(tag, ",") {
			y = append(y, strings.TrimSpace(x))
		}
		if y[0] == "" {
			y[0] = f.Name
		}
		p = append(p, Label{i, y})
	}
	this.Tag[x] = p
	return p
}

// 编码一个值为中间数据
func (this *Translator) Encode(x reflect.Value) (interface{}, error) {
	y := x.Type()
	if _, ok := this.Raw[y]; ok {
		return x.Interface(), nil
	}
	switch x.Kind() {
	case reflect.Bool:
		return x.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return x.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return x.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return x.Float(), nil
	case reflect.Complex64, reflect.Complex128:
		return x.Complex(), nil
	case reflect.String:
		return x.String(), nil
	case reflect.Ptr, reflect.Interface:
		if x.IsNil() {
			return nil, nil
		}
		return this.Encode(x.Elem())
	case reflect.Slice:
		if x.IsNil() {
			return nil, nil
		}
		if y.Elem().Kind() == reflect.Uint8 {
			return x.Bytes(), nil
		}
		fallthrough
	case reflect.Array:
		s := []interface{}{}
		for i, l := 0, x.Len(); i < l; i++ {
			v, e := this.Encode(x.Index(i))
			if e != nil {
				return nil, e
			}
			s = append(s, v)
		}
		return s, nil
	case reflect.Map:
		if x.IsNil() {
			return nil, nil
		}
		s := []Item{}
		for _, k := range x.MapKeys() {
			K, e := this.Encode(k)
			if e != nil {
				return nil, e
			}
			V, e := this.Encode(x.MapIndex(k))
			if e != nil {
				return nil, e
			}
			s = append(s, Item{K, V})
		}
		return s, nil
	case reflect.Struct:
		label, ok := this.Tag[y]
		if !ok {
			label = this.GetLabel(y)
		}
		s := []Attr{}
		for i := 0; i < len(label); i++ {
			v := x.Field(label[i].N)
			if !label[i].Has("omitempty") || !Zero(v) {
				V, e := this.Encode(v)
				if e != nil {
					return nil, e
				}
				s = append(s, Attr{label[i].Name(), V})
			}
		}
		return s, nil
	}
	return nil, UnsupportType
}

// 解码中间数据并填充值
func (this *Translator) Decode(x reflect.Value, d interface{}) error {
	y := x.Type()
	if y == reflect.TypeOf(d) {
		x.Set(reflect.ValueOf(d))
		return nil
	}
	switch x.Kind() {
	case reflect.Bool:
		if u, ok := d.(bool); ok {
			x.SetBool(u)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if u, ok := d.(int64); ok {
			x.SetInt(u)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, ok := d.(uint64); ok {
			x.SetUint(u)
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if u, ok := d.(float64); ok {
			x.SetFloat(u)
			return nil
		}
	case reflect.Complex64, reflect.Complex128:
		if u, ok := d.(complex128); ok {
			x.SetComplex(u)
			return nil
		}
	case reflect.String:
		if u, ok := d.(string); ok {
			x.SetString(u)
			return nil
		}
		if u, ok := d.([]byte); ok {
			x.SetString(string(u))
			return nil
		}
	case reflect.Ptr:
		if d == nil {
			x.Set(reflect.Zero(y))
			return nil
		} else {
			if x.IsNil() {
				x.Set(reflect.New(y))
			}
			return this.Decode(x.Elem(), d)
		}
	case reflect.Interface:
		if d == nil {
			x.Set(reflect.Zero(y))
		} else {
			x.Set(reflect.ValueOf(d))
		}
		return nil
	case reflect.Slice:
		if d == nil {
			x.Set(reflect.Zero(y))
			return nil
		}
		if y.Elem().Kind() == reflect.Uint8 {
			if u, ok := d.(string); ok {
				x.SetBytes([]byte(u))
				return nil
			}
			if u, ok := d.([]byte); ok {
				x.SetBytes(u)
				return nil
			}
		}
		if u, ok := d.([]interface{}); ok {
			n := x
			for i, l := 0, len(u); i < l; i++ {
				v := reflect.New(y.Elem()).Elem()
				e := this.Decode(v, u[i])
				if e != nil {
					return e
				}
				n = reflect.Append(n, v)
			}
			x.Set(n)
			return nil
		}
	case reflect.Array:
		if u, ok := d.([]interface{}); ok {
			l := x.Len()
			if l > len(u) {
				l = len(u)
			}
			for i := 0; i < l; i++ {
				e := this.Decode(x.Index(i), u[i])
				if e != nil {
					return e
				}
			}
			return nil
		}
	case reflect.Map:
		if d == nil {
			x.Set(reflect.Zero(y))
			return nil
		}
		if u, ok := d.([]Item); ok {
			for i, l := 0, len(u); i < l; i++ {
				k := reflect.New(y.Key()).Elem()
				e := this.Decode(k, u[i].K)
				if e != nil {
					return e
				}
				v := reflect.New(y.Elem()).Elem()
				e = this.Decode(v, u[i].V)
				if e != nil {
					return e
				}
				x.SetMapIndex(k, v)
			}
			return nil
		}
		if u, ok := d.(map[string]interface{}); ok {
			if y.Key().Kind() == reflect.String {
				for K, V := range u {
					v := reflect.New(y.Elem()).Elem()
					e := this.Decode(v, V)
					if e != nil {
						return e
					}
					x.SetMapIndex(reflect.ValueOf(K), v)
				}
				return nil
			}
		}
	case reflect.Struct:
		if u, ok := d.(map[string]interface{}); ok {
			label, ok := this.Tag[y]
			if !ok {
				label = this.GetLabel(y)
			}
			for i := 0; i < len(label); i++ {
				v := x.Field(label[i].N)
				if w, ok := u[label[i].Name()]; ok {
					e := this.Decode(v, w)
					if e != nil {
						return e
					}
				} else if label[i].Has("omitempty") {
					v.Set(reflect.Zero(v.Type()))
				}
			}
			return nil
		}
	}
	return UnmatchedType
}
