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