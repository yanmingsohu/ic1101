@echo
echo 编译时删除符号表和调试信息

go build -o ./build/ic1101.exe -ldflags "-w -s" .
echo ok