package bus

import (
	"errors"
	"strconv"
)


type IntData struct {
  D int
}


func (r *IntData) Int() int {
  return r.D
}


func (r *IntData) Int64() int64 {
  return int64(r.D)
}


func (r *IntData) Float() float32 {
  return float32(r.D)
}


func (r *IntData) Float64() float64 {
  return float64(r.D)
}


func (r *IntData) String() string {
  return strconv.Itoa(r.D)
}


func (r *IntData) Bool() bool {
  return r.D != 0
}


func (r *IntData) Src() interface{} {
  return r.D
}


type Int64Data struct {
  D int64
}


func (r *Int64Data) Int() int {
  return int(r.D)
}


func (r *Int64Data) Int64() int64 {
  return int64(r.D)
}


func (r *Int64Data) Float() float32 {
  return float32(r.D)
}


func (r *Int64Data) Float64() float64 {
  return float64(r.D)
}


func (r *Int64Data) String() string {
  return strconv.FormatInt(r.D, 10)
}


func (r *Int64Data) Bool() bool {
  return r.D != 0
}


func (r *Int64Data) Src() interface{} {
  return r.D
}


type UInt64Data struct {
  D uint64
}


func (r *UInt64Data) Int() int {
  return int(r.D)
}


func (r *UInt64Data) Int64() int64 {
  return int64(r.D)
}


func (r *UInt64Data) Float() float32 {
  return float32(r.D)
}


func (r *UInt64Data) Float64() float64 {
  return float64(r.D)
}


func (r *UInt64Data) String() string {
  return strconv.FormatUint(r.D, 10)
}


func (r *UInt64Data) Bool() bool {
  return r.D != 0
}


func (r *UInt64Data) Src() interface{} {
  return r.D
}


type FloatData struct {
  D float32
}


func (r *FloatData) Int() int {
  return int(r.D)
}


func (r *FloatData) Int64() int64 {
  return int64(r.D)
}


func (r *FloatData) Float() float32 {
  return float32(r.D)
}


func (r *FloatData) Float64() float64 {
  return float64(r.D)
}


func (r *FloatData) String() string {
  return strconv.FormatFloat(float64(r.D), 'f', 10, 32)
}


func (r *FloatData) Bool() bool {
  return r.D != 0
}


func (r *FloatData) Src() interface{} {
  return r.D
}


type Float64Data struct {
  D float64
}


func (r *Float64Data) Int() int {
  return int(r.D)
}


func (r *Float64Data) Int64() int64 {
  return int64(r.D)
}


func (r *Float64Data) Float() float32 {
  return float32(r.D)
}


func (r *Float64Data) Float64() float64 {
  return r.D
}


func (r *Float64Data) String() string {
  return strconv.FormatFloat(float64(r.D), 'f', 10, 32)
}


func (r *Float64Data) Bool() bool {
  return r.D != 0
}


func (r *Float64Data) Src() interface{} {
  return r.D
}


type StringData struct {
  D string
}


func (r *StringData) Int() int {
  v , e := strconv.ParseInt(r.D, 10, 32)
  if e != nil {
    return 0
  }
  return int(v)
}


func (r *StringData) Int64() int64 {
  v , e := strconv.ParseInt(r.D, 10, 64)
  if e != nil {
    return 0
  }
  return v
}


func (r *StringData) Float() float32 {
  v, e := strconv.ParseFloat(r.D, 32)
  if e != nil {
    return 0
  }
  return float32(v)
}


func (r *StringData) Float64() float64 {
  v, e := strconv.ParseFloat(r.D, 64)
  if e != nil {
    return 0
  }
  return v
}


func (r *StringData) String() string {
  return r.D
}


func (r *StringData) Bool() bool {
  switch r.D {
  case "on", "ON":
    return true
  case "off", "OFF":
    return false
  }
  b, err := strconv.ParseBool(r.D)
  if err != nil {
    return false
  }
  return b
}


func (r *StringData) Src() interface{} {
  return r.D
}


type BoolData struct {
  D bool
}


func (r *BoolData) Int() int {
  if r.D {
    return 1
  }
  return 0
}


func (r *BoolData) Int64() int64 {
  if r.D {
    return 1
  }
  return 0
}


func (r *BoolData) Float() float32 {
  if r.D {
    return 1
  }
  return 0
}


func (r *BoolData) Float64() float64 {
  if r.D {
    return 1
  }
  return 0
}


func (r *BoolData) String() string {
  if r.D {
    return "true"
  }
  return "false"
}


func (r *BoolData) Bool() bool {
  return r.D 
}


func (r *BoolData) Src() interface{} {
  return r.D
}


//
// 根据对象类型返回包装器
// 如果无法包装该对象, 返回错误
//
func NewDataWrap(i interface{}) (DataWrap, error) {
  switch i.(type) {
  case string:
    return &StringData{i.(string)}, nil

  case int8:
    return &IntData{i.(int)}, nil
  case int16:
    return &IntData{i.(int)}, nil
  case int32:
    return &IntData{i.(int)}, nil
  case int64:
    return &Int64Data{i.(int64)}, nil

  
  case uint8:
    return &UInt64Data{i.(uint64)}, nil
  case uint16:
    return &UInt64Data{i.(uint64)}, nil
  case uint32:
    return &UInt64Data{i.(uint64)}, nil
  case uint64:
    return &UInt64Data{i.(uint64)}, nil

  case float32:
    return &FloatData{i.(float32)}, nil
  case float64:
    return &Float64Data{i.(float64)}, nil
  
  case bool:
    return &BoolData{i.(bool)}, nil
  }
  return nil, errors.New("无法转换类型")
}