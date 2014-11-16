package encoding

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

// 用于有条理的将值表示出来，方便调试
func Detail(v interface{}) string {
	var (
		show func(int, reflect.Value, string)
		name func(reflect.Type) string
	)
	w := bytes.NewBuffer(nil)
	name = func(t reflect.Type) string {
		if x := t.Name(); x != "" {
			pre := strings.Replace(t.PkgPath(), "/", ".", -1)
			if pre != "" {
				return pre + "." + x
			} else {
				return x
			}
		}
		return t.String()
	}
	show = func(a int, x reflect.Value, sf string) {
		if x.Kind() == reflect.Interface {
			show(a, x.Elem(), sf)
			return
		}
		if a >= 0 {
			for i := 0; i < a; i++ {
				fmt.Fprint(w, "    ")
			}
		} else {
			a = -a
		}
		defer fmt.Fprint(w, sf)

		if !x.IsValid() {
			fmt.Fprint(w, "<nil>")
			return
		}
		y := x.Type()
		switch y.Kind() {
		case reflect.Ptr:
			if x.Elem().IsValid() {
				fmt.Fprint(w, "&")
				show(-a, x.Elem(), "")
			} else {
				fmt.Fprint(w, "nil")
			}
		case reflect.Struct:
			fmt.Fprintf(w, "%s{", name(y))
			if x.NumField() != 0 {
				fmt.Fprint(w, "\r\n")
				for i := 0; i < x.NumField(); i++ {
					for t := 0; t <= a; t++ {
						fmt.Fprint(w, "    ")
					}
					fmt.Fprintf(w, "%s:", y.Field(i).Name)
					show(-a-1, x.Field(i), ",\r\n")
				}
				for i := 0; i < a; i++ {
					fmt.Fprint(w, "    ")
				}
			}
			fmt.Fprint(w, "}")
		case reflect.Slice, reflect.Array:
			fmt.Fprintf(w, "%s{", name(y))
			if x.Len() != 0 {
				fmt.Fprint(w, "\r\n")
				for i := 0; i < x.Len(); i++ {
					show(a+1, x.Index(i), ",\r\n")
				}
				for i := 0; i < a; i++ {
					fmt.Fprint(w, "    ")
				}
			}
			fmt.Fprint(w, "}")
		case reflect.Map:
			fmt.Fprintf(w, "%s{", name(y))
			if x.Len() != 0 {
				fmt.Fprint(w, "\r\n")
				for _, k := range x.MapKeys() {
					show(a+1, k, ":")
					show(-a-1, x.MapIndex(k), ",\r\n")
				}
				for i := 0; i < a; i++ {
					fmt.Fprint(w, "    ")
				}
			}
			fmt.Fprint(w, "}")
		case reflect.Bool:
			fmt.Fprintf(w, "%t", x.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fmt.Fprintf(w, "%d", x.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fmt.Fprintf(w, "%d", x.Uint())
		case reflect.Float32, reflect.Float64:
			fmt.Fprintf(w, "%g", x.Float())
		case reflect.Complex64, reflect.Complex128:
			fmt.Fprintf(w, "%g", x.Complex())
		case reflect.Uintptr:
			fmt.Fprintf(w, "0x%X", x.Uint())
		case reflect.String:
			fmt.Fprintf(w, "%q", x.String())
		default:
			fmt.Fprintf(w, "<%s>", name(y))
		}
	}
	show(0, reflect.ValueOf(v), "\r\n")
	return w.String()
}
