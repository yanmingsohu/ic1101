@echo
echo 编译时删除符号表和调试信息

pushd brick
node build
popd

go build -o ./build/ic1101.exe -ldflags "-w -s" .

pushd build
zip /down1/ic.zip ic1101.exe
popd

echo ok