/**
 *  Copyright 2023 Jing Yanming
 * 
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
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