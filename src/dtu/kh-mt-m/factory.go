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
  d := dtu_impl{
    url     : url, 
    event   : handle, 
  }
  if err := d.init(); err != nil {
    return nil, err
  }
  return &d, nil
}