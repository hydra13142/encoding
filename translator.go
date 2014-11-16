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

// 判断一个值是否为零值
func Zero(x reflect.Value) bool {
	switch x.Kind() {
	case reflect.Bool:
		return x.Bool() == false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return x.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return x.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return x.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return x.Complex() == (0 + 0i)
	case reflect.UnsafePointer:
		return x.Pointer() == 0
	case reflect.String:
		return x.String() == ""
	case reflect.Interface, reflect.Ptr, reflect.Chan, reflect.Func:
		return x.IsNil()
	case reflect.Slice, reflect.Map:
		return x.IsNil() || x.Len() == 0
	case reflect.Array:
		for i, l := 0, x.Len(); i < l; i++ {
			if !Zero(x.Index(i)) {
				return false
			}
		}
	case reflect.Struct:
		for i, l := 0, x.NumField(); i < l; i++ {
			if !Zero(x.Field(i)) {
				return false
			}
		}
	}
	return true
}

// 实现中间数据与具体类型编解码
type Translator string

// 获取某个字段的信息（是否导出，名称，是否零值不编码）
func (t Translator) Refer(f *reflect.StructField) (bool, string, bool) {
	if f.Name[0] < 'A' || f.Name[0] > 'Z' {
		return false, "", false
	}
	tag := f.Tag.Get(string(t))
	if tag == "-" {
		return false, "", false
	}
	if tag == "" {
		return true, f.Name, false
	}
	x := strings.SplitN(tag, ",", 2)
	n := strings.TrimSpace(x[0])
	o := false
	if n == "" {
		n = f.Name
	}
	if len(x) == 2 {
		o = (strings.TrimSpace(x[1]) == "omitempty")
	}
	return true, n, o
}

// 编码一个值为中间数据
func (t Translator) Encode(x reflect.Value) (interface{}, error) {
	y := x.Type()
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
		return t.Encode(x.Elem())
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
			v, e := t.Encode(x.Index(i))
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
		if y.Key().Kind() == reflect.String {
			s := []Attr{}
			for _, k := range x.MapKeys() {
				V, e := t.Encode(x.MapIndex(k))
				if e != nil {
					return nil, e
				}
				s = append(s, Attr{k.String(), V})
			}
			return s, nil
		}
		s := []Item{}
		for _, k := range x.MapKeys() {
			K, e := t.Encode(k)
			if e != nil {
				return nil, e
			}
			V, e := t.Encode(x.MapIndex(k))
			if e != nil {
				return nil, e
			}
			s = append(s, Item{K, V})
		}
		return s, nil
	case reflect.Struct:
		s := []Attr{}
		for i, l := 0, x.NumField(); i < l; i++ {
			f := y.Field(i)
			v := x.Field(i)
			ex, nm, op := t.Refer(&f)
			if ex {
				if !op || !Zero(v) {
					V, e := t.Encode(v)
					if e != nil {
						return nil, e
					}
					s = append(s, Attr{nm, V})
				}
			}
		}
		return s, nil
	}
	return nil, UnsupportType
}

// 解码中间数据并填充值
func (t Translator) Decode(x reflect.Value, d interface{}) error {
	y := x.Type()
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
			return t.Decode(x.Elem(), d)
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
				e := t.Decode(v, u[i])
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
				e := t.Decode(x.Index(i), u[i])
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
				e := t.Decode(k, u[i].K)
				if e != nil {
					return e
				}
				v := reflect.New(y.Elem()).Elem()
				e = t.Decode(v, u[i].V)
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
					e := t.Decode(v, V)
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
			for i, l := 0, x.NumField(); i < l; i++ {
				f := y.Field(i)
				v := x.Field(i)
				ex, nm, op := t.Refer(&f)
				if ex {
					if w, ok := u[nm]; ok {
						e := t.Decode(v, w)
						if e != nil {
							return e
						}
					} else if op {
						v.Set(reflect.Zero(f.Type))
					}
				}
			}
			return nil
		}
	}
	return UnmatchedType
}
