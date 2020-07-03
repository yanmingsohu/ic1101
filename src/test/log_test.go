package test

import (
	"ic1101/src/core"
	"log"
	"testing"
)

const (
  str = "this is log test, this is log test, this is log test."
  count = 10000
)

func BenchmarkLog1(t *testing.B) {
  core.SetupLogger()
  for i := 0; i<t.N; i++ {
    log.Println(str, i)
  }
  t.Log("ok, test log with channel")
}


func BenchmarkLog2(t *testing.B) {
  core.UninstallLogger()
  for i := 0; i<t.N; i++ {
    log.Println(str, i)
  }
  t.Log("ok, test org log")
}