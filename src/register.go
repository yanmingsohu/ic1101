//
// 在这里注册的包, 会调用 init, 注册表只有这个作用
//
package register

import (
	_ "ic1101/src/bus/m-bus"
	_ "ic1101/src/bus/modbus"
	_ "ic1101/src/bus/random"

	_ "ic1101/src/dtu"
	_ "ic1101/src/dtu/kh-mt-m"

	_ "ic1101/src/jsslib"
)
