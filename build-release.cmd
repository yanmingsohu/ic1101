@echo
echo Remove symbol table and debug information during compilation

pushd brick
node build
popd

go build -o ./build/ic1101.exe -ldflags "-w -s" .

pushd build
rem  zip /down1/ic.zip ic1101.exe
popd

echo ok