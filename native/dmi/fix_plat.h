#ifndef FIX_PLAT_H
#define FIX_PLAT_H

// Windows platform
#if defined(_WIN32) || defined(_WIN64) || defined(WIN32)
#if NODE_MAJOR_VERSION < 4
  typedef int ssize_t;
#endif
#endif

// Linux platform
#if defined(__LINUX) || defined(linux) || defined(__linux)
  #include <string.h>
#endif

#endif