# IC1101


## 测试/开发

`air`

## 运行测试用例

`go test ic1101/src/test`


# TODO

* 设备属性支持宽度和顺序定义
* 字典可导入/导出
* 版本号的显示
* 日志分类
* 设备版本落后于原型版本时, 在列表中提示
* 在总线中 '立即下发' 指令


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


# 协议参考

* [Modbus](https://github.com/goburrow/modbus)
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