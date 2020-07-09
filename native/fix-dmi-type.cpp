#include "fix-dmi-type.h"
#include <stdio.h>
#include <string.h>


//
// 将除了类型和长度外全部删除
//
static inline void clear_header_data(pdmibyte buf,int header_len) {
  memset(buf + 2, 0, header_len - 2);
}


//
// type4 是 cpu 状态, 头域中的装态位不稳定, 将除了类型和长度外全部删除
//
static inline void fix_type_4(pdmibyte buf, int hlen) {
  clear_header_data(buf, hlen);
}


//
// 在 hp 的机器上出现, 作用未知
// 
static inline void fix_type_219(pdmibyte buf, int hlen) {
  clear_header_data(buf, hlen);
}


//
// 在联想 QiTianM4600 上, dmi 长度和内容在 4000 字节之后重启变化, 长度也会变化
// 返回修正后的缓冲区长度 - 1000.
//
static inline int fix_buf_use_QiTianM4600(pdmibyte buf, int hlen) {
  if (memcmp(buf, "QiTianM4600", hlen) == 0) {
    printf("Check LENOVO QiTianM4600-N000.\n");
    return int(hlen / 1000 - 1) * 1000;
  }
  return hlen;
}


static pdmibyte dmi_strings(pdmibyte buf, pdmibyte bend) {
  pdmibyte begin;
  pdebug("\tStrings:\n");

  while (buf < bend) {
    begin = buf;
    while (*buf != 0 && buf < bend) {
      ++buf;
    }
    pdebug("\t\t%s\n", begin);
    ++buf;
    //
    // 连续两个 0x0 作为结束
    //
    if (*buf == 0) {
      ++buf;
      break;
    }
  };
 
  return buf;
}


static pdmibyte dmi_header(pdmibyte buf, pdmibyte bend) {
  dmibyte type = buf[0];
  dmibyte len = buf[1];

  switch (type) {
    case 4:
      fix_type_4(buf, len);
      break;

    case 0xdb:
      fix_type_219(buf, len);
      break;
  }

  pdebug("\tTYPE %d, header len %d\n", type, len);
  return dmi_strings(buf + len, bend);
}


//
// 返回修正后的缓冲区长度.
//
int fix_dmi(pdmibyte buf, int len) {
  pdmibyte end = buf + len;
  pdebug("\n\nDMI information [%x]\n", buf);

  do {
    buf = dmi_header(buf, end);
    pdebug("\n");
  } while(buf < end);

  len = fix_buf_use_QiTianM4600(buf, len);

  return len;
}