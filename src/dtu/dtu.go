package dtu

import (
	"bytes"
	"errors"
	"log"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"
)


// 不要直接引用
var _dtu_factorys = make(map[string]Fact)


//
// 解析 url 并返回 dtu 实现对象
// dtu url 格式:
//   dtu://[host][:port][/dtu-type][?parameaters]
//
func GetDtu(url *url.URL, handle Event) (Impl, error) {
  if handle == nil {
    return nil, errors.New("事件处理不能为空")
  }
  fact, err := CheckUrl(url)
  if err != nil {
    return nil, err
  }
  return fact.New(url, handle)
}


//
// 检查 url 的基础参数, 只有在绝对可以生成有效 DTU 的时候才返回 DTU 工厂
//
func CheckUrl(url *url.URL) (Fact, error) {
  if url.Scheme != "dtu" {
    return nil, errors.New("不是 dtu URL")
  }
  
  factName := url.EscapedPath()
  if factName == "" {
    return nil, errors.New("必须选择 DTU 类型")
  }
  if factName[0] == '/' {
    factName = factName[1:]
  } else {
    return nil, errors.New("必须选择 DTU 类型")
  }
  f, has := _dtu_factorys[factName]
  if !has {
    return nil, errors.New("DTU 不存在 "+ factName)
  }
  return f, nil
}


//
// 注册一个 dtu 工厂, 任何错误都会导致 panic, 如命名冲突
//
func RegFact(f Fact) {
  if _, has := _dtu_factorys[f.Name()]; has {
    panic(errors.New("工厂名称冲突" + f.Name()))
  }
  _dtu_factorys[f.Name()] = f
  log.Println("DTU reg:", f.Name(), f.Desc())
}


//
// 创建 DTU 实例的工厂
//
type Fact interface {
  //
  // 返回 DTU 实例, 或创建失败的原因
  //
  New(url *url.URL, handle Event) (Impl, error)
  //
  // 该 dtu 的类型名, 使用该名称注册到 dtu 注册表
  //
  Name() string
  //
  // 对 dtu 的描述 如: 型号, 版本, 厂家
  //
  Desc() string
}


//
// 与 DTU 之间交换数据, 一个 DTU 的实现可以与一组物理 DTU 通信,
// 此时 DTU 处于 tcp/udp 服务器模式, 使用 Context 来确定对端.
// 该对象是线程安全的
//
type Impl interface {
  //
  // 返回上下文, 如果上下文关闭返回错误
  //
  GetContext(id int) (Context, error)
  //
  // 停止 DTU 所有任务
  //
  Stop()
  //
  // 关闭上下文, 并释放内存
  //
  CloseContext(id int) error
}


//
// 上下文用于确定, 一个对端(客户端)实例的连接, 该对象可以安全的保存或重用,
// 该对象生命期总是短于 Impl. 该对象线程不安全
//
type Context interface {
  //
  // 返回该上下文的唯一标识
  //
  Id() (int)
  //
  // 返回上下文参数, 不同的 DTU 实现会有不同的参数定义.
  //
  Get(name string) (interface{}, error)
  //
  // 允许在上下文绑定变量, 之后用 Get() 可以取得该变量.
  //
  Bind(name string, data interface{}) error
  //
  // 关闭这个上下文, 实现通常会关闭底层网络连接, 重复调用该方法是安全的.
  //
  Close() error
  //
  // 对于该函数的使用方, DTU 中的数据转换是透明的, 
  // 用户只关心与设备之间的原始数据格式.
  // 如果 dtu 不支持该方法则总是返回错误,
  // Conn.Close() 方法与 Context.Close() 有相同的效果
  //
  GetConn() (net.Conn, error)
}


//
// 接受 dtu 消息, 需要保证线程安全
//
type Event interface {
  //
  // 创建新的上下文, 通常是因为一个新的客户端发起了请求.
  // 如果创建失败则 ctx 为空, e 将保存失败原因
  //
  NewContext(ctx Context, e error)
  //
  // 在启动成功后, 任何关闭了 DTU 的行为都会调用该方法.
  // 如果是因为错误导致 DTU 终止, 则带有 error 参数, 否则为空.
  //
  OnClose(error)
  //
  // DTU 启动消息
  //
  OnStart(msg string)
}


//
// Impl 接口辅助类, 提供基本功能, 该对象本身是一把锁
//
type ImplHelp struct {
  sy      *sync.Mutex
  CtxMap  map[int]Context
  Run     bool
}


func NewImplHelp() ImplHelp {
  return ImplHelp{
    sy      : new(sync.Mutex),
    CtxMap  : make(map[int]Context),
    Run     : true,
  }
}


func (h *ImplHelp) SaveContext(c Context) error {
  h.sy.Lock()
  defer h.sy.Unlock()
  
  if !h.Run {
    return errors.New("DTU 已经关闭")
  }

  old, has := h.CtxMap[c.Id()]
  if has {
    old.Close()
  }
  h.CtxMap[c.Id()] = c
  return nil
}


func (h *ImplHelp) GetContext(id int) (Context, error) {
  h.sy.Lock()
  defer h.sy.Unlock()

  if !h.Run {
    return nil, errors.New("DTU 已经关闭")
  }
  ctx, has := h.CtxMap[id]
  if !has {
    return nil, errors.New("上下文不存在"+ strconv.Itoa(id))
  }
  return ctx, nil
}


func (h *ImplHelp) CloseContext(id int) error {
  h.sy.Lock()
  defer h.sy.Unlock()

  if !h.Run {
    return errors.New("DTU 已经关闭")
  }
  delete(h.CtxMap, id)
  return nil
}


//
// Context 复制类, 提供基本功能
//
type CtxHelp struct {
  bind   map[string]interface{}
  ID     int
  closed bool
  impl   Impl
}


func (c *CtxHelp) Id() int {
  return c.ID
}


func (c *CtxHelp) InitHelp(id int, di Impl) {
  c.ID = id
  c.bind = make(map[string]interface{})
  c.closed = false
  c.impl = di
}


func (c *CtxHelp) Get(name string) (interface{}, error) {
  if c.closed {
    return nil, errors.New("上下文已经关闭")
  }
  ret, has := c.bind[name]
  if !has {
    return nil, errors.New("变量不存在 "+ name)
  }
  return ret, nil
}


func (c *CtxHelp) Bind(name string, data interface{}) error {
  if c.closed {
    return errors.New("上下文已经关闭")
  }
  c.bind[name] = data
  return nil
}


//
// 通常在 Close() 中调用该方法, 关闭 help 的所有资源
// 如果关闭成功返回 true, 已经关闭返回 false
//
func (c *CtxHelp) CloseHelp() bool {
  if c.closed {
    return false
  }
  c.closed = true
  c.bind = nil
  c.impl.CloseContext(c.ID)
  return true
}


//
// 这是 net.Conn 的包装器, 除了 Read/Write 方法之外都直接调用 Conn 的方法.
//
type ConnWrap struct {
  Conn net.Conn
  ctx  Context
}

func (c *ConnWrap) InitWrap(ctx Context, conn net.Conn) {
  c.Conn = conn
  c.ctx = ctx
}

func (c *ConnWrap) Close() error {
  return c.ctx.Close()
}

func (c *ConnWrap) LocalAddr() net.Addr {
  return c.Conn.LocalAddr()
}

func (c *ConnWrap) RemoteAddr() net.Addr {
  return c.Conn.RemoteAddr()
}

func (c *ConnWrap) SetDeadline(t time.Time) error {
  return c.Conn.SetDeadline(t)
}

func (c *ConnWrap) SetReadDeadline(t time.Time) error {
  return c.Conn.SetReadDeadline(t)
}

func (c *ConnWrap) SetWriteDeadline(t time.Time) error {
  return c.Conn.SetWriteDeadline(t)
}


type RemoveDirty struct {
  state int
  dirty []byte
  dl    int
}


func NewRemoveDirty(dirty []byte) *RemoveDirty {
  return &RemoveDirty{0, dirty, len(dirty)}
}


//
// 从 has_dirty 中删除 drity 数据并返回最终数据长度
//
func (r *RemoveDirty) Modify(has_dirty []byte) int {
  end := len(has_dirty)
  begin := 0
  for {
    if i := bytes.Index(has_dirty[begin:], r.dirty); i>=0 {
      i += begin
      j := i + r.dl 
      if j < end {
        for x := j; x < end; x++ {
          has_dirty[x-r.dl] = has_dirty[x]
          has_dirty[x] = 0
        }
      } else {
        end -= r.dl
        break;
      }
      end -= r.dl
      begin = i + r.dl
    } else {
      break;
    }
  }
  return end
}