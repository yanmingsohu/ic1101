package bus_mqtt

import (
	"crypto/tls"
	"ic1101/src/bus"
	"strconv"
	"time"

	mq "github.com/eclipse/paho.mqtt.golang"
)

const STOP_WAIT = 1e3 // ms


type message struct {
  m   mq.Message
  s   *slot_impl
  c   mq_conv_dw
}


type bus_impl struct {
  cli   mq.Client
  msg   chan *message
}


func new_opt(r bus.BusReal) *mq.ClientOptions {
  url := r.URL()
  q := url.Query()

  opt := mq.NewClientOptions()

  addr := url.Scheme +"://"+ url.Host
  opt.AddBroker(addr)
  opt.SetUsername(url.User.Username())

  pass, haspass := url.User.Password() 
  if haspass {
    opt.SetPassword(pass)
  }
  opt.SetClientID(q.Get("clientID"))

  cs, _ := strconv.ParseBool(q.Get("cleanSession"))
  opt.SetCleanSession(cs)

  sec, err := strconv.ParseInt(q.Get("timeout"), 10, 32)
  if err == nil && sec >= 0 {
    to := time.Second * time.Duration(sec)
    opt.SetConnectTimeout(to)
    opt.SetPingTimeout(to * 2)
    opt.SetWriteTimeout(to)
  }

  cert_pem := q.Get("certPEM");
  key_pem := q.Get("keyPEM");
  
  if cert_pem != "" || key_pem != "" {
    cert, err := tls.X509KeyPair([]byte(cert_pem), []byte(key_pem))
    if err != nil {
      r.Log("设置 TLS 失败", err)
    } else {
      tlsCfg := tls.Config{Certificates: []tls.Certificate{cert}}
      opt.SetTLSConfig(&tlsCfg)
    }
  }
  return opt
}


func (b *bus_impl) Start(r bus.BusReal) error {
  b.cli = mq.NewClient(new_opt(r))
  tk := b.cli.Connect()
  if tk.Wait() && tk.Error() != nil {
    return tk.Error()
  }
  r.Log("总线启动")
  b.msg = make(chan *message, 3)
  go b.recv_thread(r)

  for _, s := range r.Datas() {
    if err := b.subscribe(s.(*slot_impl), r); err != nil {
      defer b.release()
      return err
    }
  }
  return nil
}


func (b *bus_impl) release() {
  b.cli.Disconnect(STOP_WAIT)
  close(b.msg)
}


func (b *bus_impl) subscribe(s *slot_impl, r bus.BusReal) error {
  conv, err := get_conv_mq2dw(s.data_type)
  if err != nil {
    return err
  }

  tk := b.cli.Subscribe(s.topic, s.q, func(c mq.Client, msg mq.Message) {
    b.msg <- &message{msg, s, conv}
  })

  if tk.Wait() && tk.Error() != nil {
    return tk.Error()
  }
  return nil
}


func (b *bus_impl) recv_thread(r bus.BusReal) {
  for m := range b.msg {
    t := time.Now()
    d := m.m.Payload()[m.s.offset:]
    
    if len(d) > 0 {
      r.Event().OnData(m.s, &t, m.c(d))
    } else {
      r.Log("主题", m.s.topic, "没有数据载荷")
    }
  }
}


// 同步数据什么也不做
func (b *bus_impl) SyncData(r bus.BusReal, t *time.Time) error {
  return nil
}


func (b *bus_impl) Stop(r bus.BusReal) {
  b.release()
}


func (b *bus_impl) SendCtrl(_s bus.Slot, d bus.DataWrap, t *time.Time) error {
  s := _s.(*slot_impl)
  tk := b.cli.Publish(s.topic, s.q, false, conv_dw2mq(s.data_type, d))
  if tk.Wait() && tk.Error() != nil {
    return tk.Error()
  }
  return nil
}