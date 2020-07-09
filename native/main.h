#pragma once

typedef unsigned char UBYTE;
typedef UBYTE* PUBYTE;
typedef unsigned int BLEN;

//
// 使用本机硬件信息与 i 产生校验, 返回校验和长度, 返回缓冲区在 out 中
// out 必须分配 crypto_length() 的长度, 否则缓冲区溢出.
//
BLEN crypto_encode(UBYTE *i, BLEN ilen, UBYTE *out);

//
// 返回输出缓冲区的最小长度
//
BLEN crypto_length();