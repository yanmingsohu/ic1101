# IC1101


## 测试/开发

`air`

## 运行测试用例

`go test ic1101/src/test`

## 发布

`make`

> make 会编译 c 加密库, 在测试前必须编译.

> 参数: `#cgo LDFLAGS: -lstdc++` 可以解决编译时异常
  `undefined reference to 'operator new[](unsigned long long)'`
  https://github.com/golang/go/issues/18460


# TODO

* 设备属性表单支持宽度和顺序定义
* 字典可导入/导出
* 日志分类
* 设备版本落后于原型版本时, 在列表中提示


# 参考

* [命名](ttps://www.universeguide.com/galaxy/ic1101)
* [Web样式](https://purecss.io/layouts/)
* [air](https://github.com/cosmtrek/air)
* [MongoDB](https://docs.mongodb.com/manual/reference/method/db.collection.insertOne/)
* [Select2](https://select2.org/data-sources/ajax)
* [modbus从站模拟器](https://www.modbusdriver.com/diagslave.html)
* [modbus开发资料](http://www.dalescott.net/modbus-development/)
* [Logger](https://godoc.org/go.uber.org/zap)
* [chart](https://github.com/apache/incubator-echarts)
* [JavaScript](https://github.com/dop251/goja)
* [Markdown解析器](https://github.com/markdown-it/markdown-it)
* [语法高亮](https://prismjs.com/)[git](https://github.com/PrismJS/prism)


# 协议参考

* [Modbus](github.com/yanmingsohu/modbus)
* [MBus](https://github.com/karl-gustav/ams-han)
* [MQTT](https://github.com/eclipse/paho.mqtt.golang)
* [MQTT](https://github.com/VolantMQ/volantmq)


# DTU

* [北京科慧铭远自控](http://www.msi-automation.com/jishuzhichi.html)


# Modbus 浮点数

地址    +0          +1           +2           +3
内容    SEEE EEEE   EMMM MMMM    MMMM MMMM    MMMM MMMM
 
S   符号位，1是负，0是正
E   偏移127的幂，二进制阶码=(EEEEEEEE)-127
M   24位的尾数保存在23位中，只存储23位，最高位固定为1