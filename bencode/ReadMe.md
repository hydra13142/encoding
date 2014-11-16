bencode
====

bencode包提供bencode编码格式的编解码器。

本包可以编解码如下类型：

1. int、string、interface{}
2. 如果slice的成员类型可编解码，则该slice也可编解码
3. 如果struct的所有可导出字段都可编解码，则该struct可编解码
4. 如果map的键类型为string，而值类型可编解码，则该map可编解码
5. 如果值的类型为interface{}，该接口下层应可以编解码，否则会出错
6. 可以安全的处理类型的循环引用，但值的循环引用会导致死循环

可以使用标签来修改编码后的字段名，如：

1. `bencode:"xxxx"`表示使用xxxx作为字典的键
2. `bencode:"-"`表示忽略该字段
3. `bencode:""`或没有标签时，会使用字段的名字作为字典的键
4. 匿名字段和普通字段同等对待（不会压平）
5. 可设置omitempty属性：`bencode:",omitempty"`和`bencode:"xxxx,omitempty"`
6. 如设置omitempty，编码时该字段为零值不会编码，解码时如无该字段会赋以零值