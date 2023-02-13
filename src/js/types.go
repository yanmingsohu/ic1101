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
package js

import "github.com/dop251/goja"


type JSValue interface {
  // 把 go 对象包装为 js 对象
  Value(i interface{}) goja.Value
  // 创建一个空 js 对象
  NewObject() *goja.Object
}


//
// 创建程序库的工厂
//
type LibFact interface {
  // 创建程序库的实例
  New(JSValue) interface{}
}