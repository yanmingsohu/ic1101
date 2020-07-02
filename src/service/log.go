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