/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package motan

import (
	motan "github.com/Loopring/motan-go"
)

func InitClient(options MotanClientOptions) *motan.Client {
	mccontext := motan.GetClientContext(options.ConfFile)
	//extFactory := motan.GetDefaultExtFactory()
	//extFactory.RegistryExtSerialization(serialize.Gob, 8, func() motancore.Serialization {
	//	return &serialize.GobSerialization{}
	//})
	mccontext.Start(nil)
	mclient := mccontext.GetClient(options.ClientId)
	return mclient
}
