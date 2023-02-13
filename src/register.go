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
//
// 在这里注册的包, 会调用 init, 注册表只有这个作用
//
package register

import (
	_ "ic1101/src/bus/m-bus"
	_ "ic1101/src/bus/modbus"
	_ "ic1101/src/bus/mqtt"
	_ "ic1101/src/bus/random"

	_ "ic1101/src/dtu"
	_ "ic1101/src/dtu/kh-mt-m"

	_ "ic1101/src/jsslib"
)
