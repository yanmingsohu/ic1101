package core

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const (
	LOG_MAX_SIZE = 5e6
	LOG_DIR = "logs/"
	LOG_BUFF_COUNT = 64
	LOG_BUFF_SIZE = 1024
)


func UninstallLogger() {
	log.SetOutput(os.Stderr)
}


func SetupLogger() {
	name := log_file_name()
	file := open_log_file(name)
	if file == nil {
		return 
	}

	out := log_output_file{
		O: make(chan *log_buff, LOG_BUFF_COUNT),
		getBuf: make(chan *log_buff, LOG_BUFF_COUNT),
	}

	for i:=0; i<LOG_BUFF_COUNT; i++ {
		out.getBuf <- &log_buff{[LOG_BUFF_SIZE]byte{}, 0}
	}

	log.SetOutput(&out)
	c := 0

	go (func() {
		for b := range out.O {
			d := b.b[:b.l]
			os.Stdout.Write(d)
			file.Write(d)
			file.Sync()
			out.getBuf <- b

			c++
			if c > 100 {
				if info, err := file.Stat(); err == nil {
					if info.Size() > LOG_MAX_SIZE {
						if f := open_log_file(log_file_name()); f != nil {
							file.Close()
							file = f
						}
					}
				} else {
					log.Println("Get log info", err)
				}
				c = 0
			}
		}
	})()
}


type log_buff struct {
	b 	[LOG_BUFF_SIZE]byte
	l 	int
}


type log_output_file struct {
	O 			chan *log_buff
	getBuf	chan *log_buff
}


func (log *log_output_file) Write(p []byte) (n int, err error) {
	// c := make([]byte, l)
	c := <- log.getBuf
	c.l = len(p)
	copy(c.b[:], p)
	log.O <- c
	return c.l, err
}


func log_file_name() string {
	os.MkdirAll(LOG_DIR, os.ModePerm)
	name := fmt.Sprintf("ic.%s.log", time.Now().Format(time.RFC1123))
	name = strings.Replace(name, ":", "_", -1)
	return LOG_DIR + name
}


func open_log_file(name string) *os.File {
	flag := os.O_WRONLY|os.O_CREATE|os.O_APPEND
	file, err := os.OpenFile(name, flag, 0666)
	if err != nil {
		log.Println("Setup log fail", err)
		return nil
	}
	return file
}


type LogFile struct {
  Name   string      `json:"name"`
	Size   int64       `json:"size"`
	Time   time.Time   `json:"time"` 
}


func ReadLogFileList() ([]LogFile, error) {
  files, err := ioutil.ReadDir(LOG_DIR)
  if err != nil {
    return nil, err
  }
  ret := make([]LogFile, len(files))
  for i, f := range files {
    ret[i] = LogFile{ f.Name(), f.Size(), f.ModTime() }
  }
  return ret, nil
}


type LogReader interface {
  io.Reader
  io.Closer
}


func ReadLogFile(file string) (LogReader, error) {
  p := path.Join(LOG_DIR, file)
  if !strings.HasPrefix(p, LOG_DIR) {
    return nil, errors.New("无效地址 "+ file)
  }
  return os.Open(p)
}