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

package gateway

import (
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ipfs/go-ipfs-api"
	"strconv"
	"strings"
)

type IPFSPubService interface {
	PublishOrder(order types.Order) error
}

type IpfsOptions struct {
	Server          string
	Port            int
	ListenTopics    []string
	BroadcastTopics []string
}

func (opts IpfsOptions) Url() string {
	url := opts.Server
	if !strings.HasSuffix(url, ":") {
		url = url + ":"
	}
	return url + strconv.Itoa(opts.Port)
}

type IPFSPubServiceImpl struct {
	options *IpfsOptions
	sh      *shell.Shell
	url     string
}

func NewIPFSPubService(options *IpfsOptions) *IPFSPubServiceImpl {
	l := &IPFSPubServiceImpl{}
	l.url = options.Url()
	l.options = options
	l.sh = shell.NewShell(l.url)
	return l
}

func (p *IPFSPubServiceImpl) PublishOrder(order types.Order) error {
	orderJson, err := order.MarshalJSON()
	if err != nil {
		log.Debugf("ipfs pub,marshal order error:%s", err.Error())
		return err
	}
	pubErr := p.sh.PubSubPublish(p.options.BroadcastTopics[0], string(orderJson))
	if pubErr != nil {
		log.Debugf("ipfs pub,pub sub publish error:%s", pubErr.Error())
	} else {
		log.Debugf("ipfs publish order:%s", order.Hash.Hex())
	}
	return pubErr
}
