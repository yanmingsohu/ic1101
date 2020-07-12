# 设备脚本开发指南

每个设备都有自己的脚本, 可以用于数据转换.  
脚本使用 Javascript ES5 编写.

* [JS 中文文档](https://developer.mozilla.org/zh-CN/docs/Web/JavaScript)
* [JS English Documents](https://developer.mozilla.org/en-US/docs/Web/JavaScript)


## 数据槽

数据从总线中读出, 然后发送到脚本, 脚本对数据进行处理, 返回的数据保存到实时数据库中;  
如果脚本在初始化时失败, 则总线无法运行; 脚本在运行时失败, 将保存从总线中读取的原始数据.


### 脚本原型代码

这是一个默认的脚本代码, 该代码返回总线传来的数据, 此外什么都不做.

```javascript
{
  //
  // 在保存数据之前, 对数据进行转换, 如果什么都不处理则返回原始值
  //
  // dev  : class Dev 参看 Dev 类的说明.
  // time : 数据时间的毫秒值, 可以通过 Date 转换为 js 原生类型.
  // data : 设备传来的数据, js 原生数据类型.
  //
  on_data : function(dev, time, data) {
    // 默认直接返回原始值
    return data;
  },
}
```


## Class Dev

on_data 第一个参数 dev 的定义

### String GetName()

返回数据名称, 既设备原型中, 数据槽的名称.

该属性可以用来针对不同的数据进行分类处理

```javascript
{
  on_data : function(dev, time, data) {
    // 如果是 '瞬时流量' 数据, 就把数据缩小 100 倍
    if (dev.GetName() == '瞬时流量') {
        return data/100;
    } else {
        // 否则返回原始值
        return data;
    }
  },
}
```

### String GetSlot() 

返回总线上数据槽的地址 ID, 在不同的总线上, 该 ID 格式不同.

### String GetDev() 

返回设备 ID

### String GetType() 

返回数据的存储类型, 即使处理函数返回的类型与此不同, 最终也会被存储为该字段指定的类型.

* 1: 整数类型
* 2: 浮点类型
* 3: 虚拟数据
* 4: 开关类型
* 5: 字符串类型