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

package main

import (
	"fmt"
	"github.com/Loopring/extractor/node"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/params"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
)

func main() {
	app := NewApp()
	app.Action = startNode
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2013-2017 The Loopring Authors"

	globalFlags := GlobalFlags()
	app.Flags = append(app.Flags, globalFlags...)
	app.Commands = []cli.Command{}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Version = params.Version
	app.Usage = "the Loopring/extractor command line interface"
	app.Author = ""
	app.Email = ""
	return app
}

func GlobalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "config,c",
			Usage: "config file",
		},
	}
}

func startNode(ctx *cli.Context) error {
	globalConfig := SetGlobalConfig(ctx)

	logger := log.Initialize(globalConfig.Log)
	defer func() {
		if nil != logger {
			logger.Sync()
		}
	}()

	var n *node.Node
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, os.Kill)
	go func() {
		for {
			select {
			case sig := <-signalChan:
				log.Infof("captured %s, exiting...\n", sig.String())
				if nil != n {
					n.Stop()
				}
				os.Exit(1)
			}
		}
	}()

	n = node.NewNode(logger, globalConfig)

	n.Start()

	n.Wait()
	return nil
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
