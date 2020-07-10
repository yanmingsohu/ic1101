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