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
package test

import (
	"fmt"
	"ic1101/src/core"
	"testing"
)

const _yaml = `
appName: IOT-IC1101 物联网中台/组态/智能数据中心
company: 上海竹呗信息
dns: zhubei.com
email: zhubei.com
beginTime: 1527487634019
endTime: 1843106834000
z: |-
   ThUXpI5eoyOfStmXOsWULZ+0EpY5WtWonMWJYTAhTXSbECaK1eo
   q8VDbs3TNviVMpSgsxx0srIQ/TGbiY3aVlQ==
api: 
- app.module.cluster.functions()
- api.ide.code.modify.functions()
- app.module.chain.peer.platform()
- app.module.fabric.functions()
- app.module.webservice.functions()
- app.module.schedule.functions()
- app.module.mongodb.functions()
- app.module.apipm.functions()
- app.module.shell.functions()
signature: |-
   uuAzI2a9uJr4Q26E4u0DG134p4wL+wcBmbITwJwDKpTNAWHVMU
   fIbOdIy9JVC/XmTgLBF2X7AAsfdhgFX2K9lN7v5rke0kiZMqJJ
   SAN7/UI7kjNfWl/ZbTlFeprIch6f6JMvaHUwtYmsYFPylHbZ04
   MfsJOaLCXLPXfGiqJzvsD2/UIEDS6taeuYdGznQ/asPip1JMmT
   +LlwmuOSBScc92y4J/i6b4DO6q8+6FiCKjJaNwX67ZCcPSvR0K
   4fG+I5WAuthQ4wcIAlQ73dLbv9pBrXGDFtLDVw2m0APbmfC3M6
   mrwdzL8GQ7qkX+rZXDoqkvinLD2hL2GBUL32Rel9QQ==
`

func TestReadPub(t *testing.T) {
  li := core.Li{}
  li.Init(_yaml)
  li.ComputeZ()

  if err := li.Verification(); err != nil {
    fmt.Println(li.String())
    t.Fatal(err)
  }
}