.PHONY: clean all

all: c_static_lib go_executable

c_static_lib:
	gcc -c greetings/*.c
	ar rs build/greetings.a build/*.o

go_executable:
  go build -o ./build/ic1101.exe -ldflags "-w -s" .

clean: