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
package core

import (
	"fmt"
)

const GVersion = "1.0.5"
const GAppName = "IOT-IC1101 物联网中台/组态/智能数据中心"

var LOGO = `
.___ ________ ___________      .___ _________  ____  ____ _______   ____ 
|   |\_____  \\__    ___/      |   |\_   ___ \/_   |/_   |\   _  \ /_   |
|   | /   |   \ |    |  ______ |   |/    \  \/ |   | |   |/  /_\  \ |   |
|   |/    |    \|    | /_____/ |   |\     \____|   | |   |\  \_/   \|   |
|___|\_______  /|____|         |___| \______  /|___| |___| \_____  /|___|
             \/                             \/                   \/      
QQ: 412475540 / Email: yanming-sohu@sohu.com
--------------------------------------------
`

func init() {
	fmt.Print(LOGO)
	fmt.Print()
}
