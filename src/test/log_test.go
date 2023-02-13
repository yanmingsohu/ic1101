/**
 *  Copyright 2023 Jing Yanming
 * 
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
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