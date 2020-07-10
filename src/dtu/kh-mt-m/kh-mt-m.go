//
// <北京科慧铭远自控技术有限公司> 生产的 dtu
// http://www.msi-automation.com/
//
package kh_mt_m

import (
	"ic1101/src/dtu"
	"net/url"
)


func init() {
  dtu.RegFact(&factory{})
}


type factory struct {
}


func (f *factory) Name() string {
  return "kh-mt-m"
}


func (f *factory) Desc() string {
  return "科慧铭远 MBUS-MODBUS-DTU"
}


func (f *factory) New(url *url.URL, handle dtu.Event) (dtu.Impl, error) {
  d := dtu_impl{url, handle, make(map[int]dtu.Context)}
  if err := d.init(); err != nil {
    return nil, err
  }
  return &d, nil
}


type dtu_impl struct {
  url     *url.URL
  event   dtu.Event
  ctxMap  map[int]dtu.Context
}


func (d *dtu_impl) init() error {
  return nil
}


func (d *dtu_impl) GetContext(id int) (dtu.Context, error) {
  return nil, nil
}