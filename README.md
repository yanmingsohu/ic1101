# IC1101


## 测试/开发

`air`

## 运行测试用例

`go test ic1101/src/test`

## 发布

`make www`

> make 会编译 c 加密库, 在测试前必须编译.

> 参数: `#cgo LDFLAGS: -lstdc++` 可以解决编译时异常
  `undefined reference to 'operator new[](unsigned long long)'`
  https://github.com/golang/go/issues/18460
  
> 依赖 mingW 动态库:
  libgcc_s_seh-1.dll, libstdc++-6.dll, libwinpthread-1.dll


# TODO

* 设备属性表单支持宽度和顺序定义
* 日志分类
* 设备版本落后于原型版本时, 在列表中提示


# 参考

* [命名](ttps://www.universeguide.com/galaxy/ic1101)
* [Web样式](https://purecss.io/layouts/)
* [air](https://github.com/cosmtrek/air)
* [MongoDB](https://docs.mongodb.com/manual/reference/method/db.collection.insertOne/)
* [Select2](https://select2.org/data-sources/ajax)
* [Logger](https://godoc.org/go.uber.org/zap)
* [chart](https://github.com/apache/incubator-echarts)
* [JavaScript](https://github.com/dop251/goja)
* [Markdown解析器](https://github.com/markdown-it/markdown-it)
* [语法高亮](https://prismjs.com/)[git](https://github.com/PrismJS/prism)
* [当成 win 服务运行](http://nssm.cc/download)
* [cgo 静态链接](https://blog.madewithdrew.com/post/statically-linking-c-to-go/)
* [cgo 调用示例](https://github.com/draffensperger/go-interlang)


# Linux

编译时依赖

```sh
yum -y install compat-libstdc++-33.x86_64 libstdc++.x86_64 libstdc++-devel.x86_64\
  libstdc++-static.x86_64  glibc-common.x86_64 glibc.x86_64 glibc-devel.x86_64\
  glibc-static.x86_64 gcc-c++ gcc
```

配置 golang

```sh
wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz
tar -xzf go1.11.2.linux-amd64.tar.gz  -C /usr/local 
```

配置 nodejs

```sh
wget https://nodejs.org/dist/v12.18.2/node-v12.18.2-linux-x64.tar.xz
tar -xJvf node-v12.18.2-linux-x64.tar.xz  -C /usr/local/lib/nodejs 
```


# 自动化通讯协议

## 程序自动化
* BSAP
* CC-Link
* CIP
* CANopen
* ControlNet
* DeviceNet
* DF-1
* DirectNET
* EtherCAT
* Ethernet Global Data (EGD)
* Ethernet Powerlink
* EtherNet/IP
* FINS
* FOUNDATION fieldbus
* GE SRTP
* HART Protocol
* Honeywell SDS
* HostLink
* INTERBUS
* MECHATROLINK
* MelsecNet
* Optomux
* PieP
* PROFINET IO
* SERCOS interface
* SERCOS III
* Sinec H1
* SynqNet
* TTEthernet
* RAPIEnet

## 工业控制系统
* Modbus
  * [实现](github.com/yanmingsohu/modbus)
  * [modbus从站模拟器](https://www.modbusdriver.com/diagslave.html)
  * [modbus开发资料](http://www.dalescott.net/modbus-development/)
* OPC DA
* OPC HDA
* OPC UA
* MTConnect

## 智能建筑
* BACnet
* 1-Wire
* C-Bus
* DALI
* DSI
* KNX
* LonTalk
* oBIX
* VSCP
* X10
* xAP
* ZigBee

## 输配电通讯协定
* IEC 60870-5
* DNP3
* IEC 60870-6
* IEC 61850
* IEC 62351
* Profibus

## 智能电表
* M-Bus 
  * [实现](https://github.com/rscada/libmbus) 
  * [文档](https://m-bus.com/documentation-wired/01-introduction)
* ZigBee Smart Energy 2.0
* ANSI C12.18
* IEC 61107
* DLMS/IEC 62056

## 车用通讯
* CAN
  * [待验证*实现](https://github.com/brutella/can)
* FMS
* FlexRay
* IEBus
* J1587
* J1708
* J1939
* Keyword Protocol 2000
* LIN
* MOST
* NMEA 2000
* VAN

## 其他
* MQTT
  * [client 实现](https://github.com/eclipse/paho.mqtt.golang)
  * [server](https://github.com/VolantMQ/volantmq)
  * [文档](https://mcxiaoke.gitbooks.io/mqtt-cn/content/mqtt/01-Introduction.html)


## DTU
* [北京科慧铭远自控](http://www.msi-automation.com/jishuzhichi.html)