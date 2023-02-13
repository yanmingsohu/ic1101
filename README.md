# IC1101

IC1101 是一个基于 web 的高性能组态软件, 可在线配置数据逻辑总线, 如 Modbus 逻辑总线, 总线配置完毕可立即接受数据,  
IC1101 可以将全部资源编译到一个单独的 exe 中, 部署时直接启动这个可执行文件即可.  
IC1101 允许开发者用 javascript 来编写数据处理逻辑, 接受设备发来的数据通过这个 js 脚本进行前置数据处理, 然后持久化.  
带有基本的数据显示和图表.  
总线框架和 DTU 框架, 方便扩展新的总线类型和 DTU 类型.  
带有一个基本的授权模块.

在线演示, 打开网站 https://xboson.net/ 选择菜单 '开放平台', '物联网平台', 用户名:root 密码:11118888

该项目由 [上海竹呗信息技术有限公司](https://xboson.net/) 提供技术支持.


## 依赖

一个 mongodb 3.6 或更高版本的数据库, nodejs 12 或更高版本的脚本.


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
  
  
## 配置文件

在主程序目录建立文件 `ic1101.yaml`:

```
# 服务器监听端口
httpPort : 7700

# mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
mongoURL : mongodb://localhost:27017/

# 数据库名称
mongoDBname : ic1101

# 密钥加密公钥
salt : fiownvcxz,.iwo
```

在主程序目录建立文件 `root-user.yaml`:

```
username: root
password: xxxxxxxx
```


# TODO

* 设备属性表单支持宽度和顺序定义
* 日志分类
* 设备版本落后于原型版本时, 在列表中提示
* 可自定义数据持久化, 支持更多 DB 类型


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
  * [基于DCOM, 待验证*实现](https://github.com/konimarti/opc)
  * [服务端c++](https://github.com/technosoftware-gmbh/opc-daae-server-sdk)
  * [服务端, 古旧代码, 参考](https://github.com/gmist/frl)
* OPC HDA
* OPC UA
  * [基于TCP, 待验证*实现](https://github.com/gopcua/opcua)
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