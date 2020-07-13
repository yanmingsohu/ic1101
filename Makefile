
OBJPATH1 = native
OBJPATH2 = native/dmi
OUTDIR = ./build
INCLUDE = ./native

# 添加 -static-libgcc -static-libstdc++ 可以解决 libgcc 等动态库的依赖, 但是在哪里加参数?

.PHONY: clean all
all: c_static_lib go_executable

c_static_lib:
	cd build &&  gcc -c ../native/dmi/*.c 
	cd build &&  g++ -c ../native/*.cpp 
	cd build &&  ar -rv libnative.a *.o

go_executable:
	go build -o ./build/ic1101.exe -ldflags "-w -s" .

clean:
	cd build && rm *.o
