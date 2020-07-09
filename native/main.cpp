#include "ntype.h"
#include "dmi/fix_plat.h"
#include "sha512.h"
#include "stren.h"
#include "fix-dmi-type.h"
#include "main.h"


PUBYTE hd_buf;
int hd_len;


GOEXPORT BLEN crypto_encode(CHAR *i, BLEN ilen, UBYTE *out) {
  SHA512 ctx;
  ctx.init();
  ctx.update((UBYTE*) i, ilen);
  ctx.update(hd_buf, hd_len);
  ctx.final(out);
  return SHA512::DIGEST_SIZE;
}


GOEXPORT BLEN crypto_length() {
  return SHA512::DIGEST_SIZE;
}
  

void buffer_init(unsigned char *buf, int len) {
  // print_buf(buf, len);
  hd_len = len;
  hd_buf = new unsigned char[len];
  memcpy((char *)hd_buf, (char *)buf, len);
  hd_len = fix_dmi(hd_buf, hd_len);
}


void free_buffer() {
  memset(hd_buf, 0, hd_len);
  delete[] hd_buf;
  hd_len = 0;
}


#if defined(_WIN32) || defined(_WIN64)
BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
  switch (fdwReason) {
    case DLL_PROCESS_ATTACH:
      get_dev_info(buffer_init);
      break;

    case DLL_PROCESS_DETACH:
      free_buffer();
      break;
  }
  return TRUE;
}
#endif


#if defined(__LINUX) || defined(linux) || defined(__linux)
__attribute ((constructor)) void lib_init(void) {
  get_dev_info(buffer_init);
}

__attribute__((destructor)) void after_main() {
  free_buffer();
}
#endif