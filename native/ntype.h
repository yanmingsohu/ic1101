#ifndef NOTIFY_TYPE_H_INCLUDED
#define NOTIFY_TYPE_H_INCLUDED

#if defined(_WIN32) || defined(_WIN64)
#include <windows.h>
#endif

#if defined(__linux) || defined(__gnu_linux__) || defined(linux)
#endif

#include <string>

extern "C"{
#include "dmi/types.h"
#include "dmi/dmidecode.h"
}


#endif // NOTIFY_TYPE_H_INCLUDED