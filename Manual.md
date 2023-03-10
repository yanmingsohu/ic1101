# IOT-ic1101 手册

高度可定制的物联网服务中心平台
现场实施/服务开发/商业逻辑/数据分析/数据互联互通的全面解决方案


# 现场数据

使用元数据构建现场数据


## 总线管理

将虚拟设备与物理设备通过总线一对一连接, 接受物理设备的数据或控制物理设备;

根据不同的总线, 设置方法有差异;

总线在停止状态可以进行组态, 既: 将虚拟设备挂接到总线上, 与物理设备连通;

总线定时器用于定时同步数据, 控制定时器独立于总线定时器, 可以定时发送控制命令; 部分协议会忽略总线定时器, 而是被动接受数据, 如 MQTT;

总线启动后将禁止进行组态, 此时 '实时状态' 可以看到所有数据的总览, 设备数据需要进入设备管理中查看;

启动失败的总线通常由于组态错误, 或网络连接不可达, 根据不同的协议有不同;

总线数据首先发送到设备脚本, 经过脚本对数据的转换后存储到 DB 中.


## 设备管理

该设备指虚拟设备, 可以是一个设备的抽象或一系列设备的抽象;

设备保存历史数据, 通过 '数据' 可以查看设备中所有数据槽(数据通道)的历史数据, 可以按时间检索/钻取, 也可以进行同比统计分析;

设备的创建来自 '设备原型', '基本属性' 是固定的, '扩展属性' 由设备原型定义, 有那些数据槽/控制槽由 '设备原型' 定义;

创建号的设备可以挂接到总线上; 如果设备挂接的总线正在运行, 设备禁止删除;

如果设备的 '版本' 落后于 '设备原型' 的版本, 则需要对 '设备信息' 进行更新;


## 定时器

定时器描述一个时间周期, 在指定的时间, 每间隔时间, 启动一个任务, 并且允许重复执行;


# 元数据


## 字典

字典可以用于 '设备原型' 的属性字典, 由 键/值 对组成, 其内容会生成对应的下拉列表;

对于规则化数据, 应使用字典;

'auth' 是系统字典, 会影响角色权限配置中的内容: 键是 `权限字符串`, 值是 `分组,权限名称`;


## 设备脚本

设备脚本使用 Javascript ES5 编写, 用于对设备数据在保存到数据库之前进行清洗/计算/规格化/另存/智能控制;

脚本开发参考在线文档: '设备脚本开发指南'


## 设备原型

通常将一种品类的设备的固有属性, 做成一个设备原型, 之后利用这个原型可以创建出很多设备的实例(虚拟设备);

一个设备原型可以配置一系列属性, 这些属性描述了设备的特征, 属性可以来自字典, 或输入值;

设备原型可以配置一系列控制槽/数据槽, 描述了可进行通信的端口;

设备原型关联了脚本; 或不关联脚本, 将直接保存数据; 不能删除有实例化的设备原型;


## 用户管理

一个角色关联了多个权限;

一个用户关联一个角色, 并继承角色的所有权限;

如果用户没有指定的权限, 则不能调用指定的接口, 并且不能赋予自身或另一个用户该权限;

root 用户是特殊用户, 密码在配置文件中配置, 并且不显示在用户列表中, 修改密码需要重启服务器;
