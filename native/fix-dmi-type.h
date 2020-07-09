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