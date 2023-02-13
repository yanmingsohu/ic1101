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
	"ic1101/src/dtu"
	"testing"
)


func TestDirty(t *testing.T) {
  dd := "abc"
  dirty := []byte(dd)
  d := dtu.NewRemoveDirty(dirty)

  s := func (b []byte, h int) {
    r := d.Modify(b)
    if r != len(b)-h {
      t.Fatal("bad offset")
    }
    t.Log(string(b))
  }

  s([]byte("|01234|abc"), 3)
  s([]byte("|8|abc"), 3)
  s([]byte("abc|*|"), 3)
  s([]byte("abc|"), 3)
  s([]byte("abc|01234|"), 3)
  s([]byte("|01234|abc|5678|"), 3)
  s([]byte("|01234|abc|5678|abc|(*)|"), 6)
  s([]byte("|01234|ab|5678|bc|(*)|"), 0)
}