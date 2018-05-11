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

package utils

import (
	"reflect"

	"github.com/Loopring/relay-cluster/node"
	"gopkg.in/urfave/cli.v1"
)

var (
	UnlockFlag = cli.StringFlag{
		Name:  "unlocks",
		Usage: "the list of accounts to unlock",
	}
	PasswordsFlag = cli.StringFlag{
		Name:  "passwords",
		Usage: "the file contains passwords used to unlock accounts ",
	}
)

func GlobalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "config,c",
			Usage: "config file",
		},
		UnlockFlag,
		PasswordsFlag,
	}
}

func MinerFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "ringMaxLength,rml",
			Usage: "the max length of ring",
		},
		cli.StringFlag{
			Name:  "miner",
			Usage: "the encrypted private key used to sign ring",
		},
		cli.StringFlag{
			Name:  "feeRecepient,r",
			Usage: "the fee recepient address when mined a ring",
		},
	}
}

func SetGlobalConfig(ctx *cli.Context) *node.GlobalConfig {
	file := ""
	if ctx.IsSet("config") {
		file = ctx.String("config")
	}
	globalConfig := node.LoadConfig(file)

	if _, err := node.Validator(reflect.ValueOf(globalConfig).Elem()); nil != err {
		panic(err)
	}

	return globalConfig
}
