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
package service

import (
	"ic1101/brick"
	"ic1101/src/core"
)


func installLogService(b *brick.Brick) {
  ctx := &ServiceGroupContext{"", ""}

  aserv(b, ctx, "log_file_list", log_file_list)
  aserv(b, ctx, "log_load_file", log_load_file)
}


func log_file_list(h *Ht) interface{} {
  list, err := core.ReadLogFileList()
  if err != nil {
    return HttpRet{1, "读取日志列表错误", err.Error()}
  }
  return list
}


func log_load_file(h *Ht) interface{} {
  f := checkstring("日志名", h.Get("f"), 1, 255)
  read, err := core.ReadLogFile(f)
  if err != nil {
    return err
  }
  defer read.Close()
  buf := make([]byte, 255)
  
  for {
    n, err := read.Read(buf)
    if n <= 0 || err != nil {
      break;
    }
    _, err = h.W.Write(buf[:n])
    if err != nil {
      break;
    }
  }
  return nil
}