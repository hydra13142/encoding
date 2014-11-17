package encoding

import "reflect"

// 表示一个标签
type Label struct {
	N int // 字段索引
	V []string
}

// 字段名称
func (this *Label) Name() string {
	return this.V[0]
}

// 是否具有某属性
func (this *Label) Has(s string) bool {
	for i := 1; i < len(this.V); i++ {
		if this.V[i] == s {
			return true
		}
	}
	return false
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
