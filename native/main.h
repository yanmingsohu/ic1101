#pragma once
// #define GOEXPORT __declspec(dllexport)
#define GOEXPORT

typedef unsigned char UBYTE;
typedef UBYTE* PUBYTE;
typedef unsigned int BLEN;
typedef char CHAR;


#ifdef __cplusplus
extern "C" {
#endif

//
// 使用本机硬件信息与 i 产生校验, 返回校验和长度, 返回缓冲区在 out 中
// out 必须分配 crypto_length() 的长度, 否则缓冲区溢出.
//
GOEXPORT BLEN crypto_encode(CHAR *i, BLEN ilen, UBYTE *out);

//
// 返回输出缓冲区的最小长度
//
GOEXPORT BLEN crypto_length();

//
// 初始化模块
//
GOEXPORT void crypto_init();

#ifdef __cplusplus
}
#endif