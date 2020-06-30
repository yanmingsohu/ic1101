package bus

import "strconv"


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


func (r *IntData) String() string {
  return strconv.Itoa(r.D)
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


func (r *Int64Data) String() string {
  return strconv.FormatInt(r.D, 10)
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


func (r *FloatData) String() string {
  return strconv.FormatFloat(float64(r.D), 'f', 10, 32)
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


func (r *StringData) String() string {
  return r.D
}
