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
  

static void buffer_init(unsigned char *buf, int len) {
  // print_buf(buf, len);
  hd_len = len;
  hd_buf = new unsigned char[len];
  memcpy((char *)hd_buf, (char *)buf, len);
  hd_len = fix_dmi(hd_buf, hd_len);
}


GOEXPORT void crypto_init() {
  get_dev_info(buffer_init);
}


static void free_buffer() {
  memset(hd_buf, 0, hd_len);
  delete[] hd_buf;
  hd_len = 0;
}


#if defined(_WIN32) || defined(_WIN64)
BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
  switch (fdwReason) {
    case DLL_PROCESS_ATTACH:
      crypto_init();
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