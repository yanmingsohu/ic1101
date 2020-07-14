
EXT_PATH = native
DMI_PATH = native/dmi
OBJ_OUTDIR = ./build/
INCLUDE = ./native
CXXFLAGS = -static-libgcc -static-libstdc++ -Wl,-Bstatic\
 -lstdc++ -lpthread -Wl,-Bdynamic

T_DMI := $(shell ls $(DMI_PATH)/*.c)
T_DMI := $(addprefix $(OBJ_OUTDIR), $(patsubst %.c, %.o, $(notdir $(T_DMI))))

T_EXT := $(shell ls $(EXT_PATH)/*.cpp)
T_EXT := $(addprefix $(OBJ_OUTDIR), $(patsubst %.cpp, %.o, $(notdir $(T_EXT))))

# DEPS = $(T_DMI:.o=.d)

.PHONY: clean all c_static_lib go_executable www

# "目标: 普通编译 "
all: go_executable

# "目标: 网页资源编译进 exe 中 "
www: brick/resource_www.go go_executable

-include $(DEPS)

c_static_lib : build/libnative.a


build/libnative.a: $(T_DMI) $(T_EXT)
	ar -rv build/libnative.a build/*.o

# "make 会寻找不存在的目标, 并尝试使表达式匹配目标 "
$(OBJ_OUTDIR)%.o: $(EXT_PATH)/%.cpp
	g++ -c -o $@ $< 

$(OBJ_OUTDIR)%.o: $(DMI_PATH)/%.c
	gcc -c -o $@ $< 
	# gcc -MMD -MF $(patsubst %.o,%.d,$@) -MT $@ $<


go_executable: c_static_lib
	go build -o ./build/ic1101.exe -ldflags "-w -s -extldflags '-static'" .
	ldd build/ic1101


brick/resource_www.go:
	cd brick && node build

clean:
	rm build/*.o
	rm build/libnative.a
	rm brick/resource_www.go
	rm brick/ic1101.exe