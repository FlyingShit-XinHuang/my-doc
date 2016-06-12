## 获取类型

用reflect.TypeOf(variable)来获取variable的类型。

variable.(type)用在switch语句中，需要提前对variable范围有了解以便在case中匹配。

v := i.(Type)为类型断言，会判断interface是否为Type类型，如是则将对应值赋予v。否则会触发panic。
v, ok := i.(Type)不会触发panic，通过bool值ok判断是否为Type类型。
