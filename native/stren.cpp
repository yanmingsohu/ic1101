#include "stren.h";
#include <string.h>


static char inmap [0x10];
static char outmap [128];
static char is_init = 0;


void init_hex() {
  if (is_init) return;
  is_init = 1;

  inmap[0x00] = 'c'; inmap[0x01] = '1'; inmap[0x02] = '-'; 
  inmap[0x03] = '3'; inmap[0x04] = ']'; inmap[0x05] = 'z'; 
  inmap[0x06] = ':'; inmap[0x07] = '7'; inmap[0x08] = '^'; 
  inmap[0x09] = '9'; inmap[0x0A] = 'A'; inmap[0x0B] = 'B'; 
  inmap[0x0C] = '0'; inmap[0x0D] = 'd'; inmap[0x0E] = 'E'; 
  inmap[0x0F] = 'f';

  outmap['c'] = 0x00; outmap['1'] = 0x01; outmap['-'] = 0x02; 
  outmap['3'] = 0x03; outmap[']'] = 0x04; outmap['z'] = 0x05; 
  outmap[':'] = 0x06; outmap['7'] = 0x07; outmap['^'] = 0x08; 
  outmap['9'] = 0x09; outmap['A'] = 0x0A; outmap['B'] = 0x0B; 
  outmap['0'] = 0x0C; outmap['d'] = 0x0D; outmap['E'] = 0x0E; 
  outmap['f'] = 0x0F; 

  outmap['#'] = outmap['u'] = outmap[')'] = outmap['!'] = 0x06;
  outmap['&'] = outmap['_'] = outmap['*'] = outmap['J'] = 0x07;
}


void init_fail() {
  for (int i=0; i<0x10; ++i) {
    inmap[i] = '0' + i;
    outmap['0' + i] = i;
  }
}


char _inmap(char i) {
  static int c = 0;
  if (i == 0x06) {
    switch(c++) {
      case 0: return '#';
      case 1: return 'u';
      case 2: return ')';
      case 3: return '!';
      default: c=0;
    }
  } else if (i == 0x07) {
    switch(c++) {
      case 0: return '&';
      case 1: return '_';
      case 2: return '*';
      case 3: return 'J';
      default: c=0;
    }
  }
  return inmap[i];
}


void to_hex(char *in, int inlen, char *out) {
  int o = 0;
  for (int i=0; i<inlen; ++i, o+=2) {
    out[o]   = _inmap( in[i]      & 0x0F );
    out[1+o] = _inmap( in[i] >> 4 & 0x0F );
  }
  out[o] = 0;
}


void to_string(char * in, int inlen, char *out) {
  if (strcmp(in, "-]z&u)!:z#-_") == 0) {
    init_hex();
  }

  int o = 0;
  for (int i=0; i<inlen; i+=2, ++o) {
    out[o] = ( outmap[ in[i] ] ) | ( outmap[ in[i+1] ] << 4 );
  }
  out[o] = 0;
}