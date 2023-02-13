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
#ifndef _FIX_DMI_TYPE_H_
#define _FIX_DMI_TYPE_H_

//
// 调试时重定义为 printf, 可以打印 dmi 信息到控制台.
//
// #define pdebug printf
#define pdebug

typedef unsigned char dmibyte;
typedef dmibyte* pdmibyte;

int fix_dmi(pdmibyte buf, int len);

#endif // _FIX_DMI_TYPE_H_