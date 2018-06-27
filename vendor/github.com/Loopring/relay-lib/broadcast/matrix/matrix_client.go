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

package matrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Loopring/relay-lib/log"
	"io/ioutil"
	"net/http"
	"sync"
)

type MatrixClientOptions struct {
	HSUrl       string
	User        string
	Password    string
	AccessToken string
}

type MatrixClient struct {
	HSUrl    string
	User     string
	Password string
	LoginRes *LoginRes
	UserId   string
	mtx      sync.RWMutex
}

func NewMatrixClient(options MatrixClientOptions) (*MatrixClient, error) {
	matrixClient := &MatrixClient{}
	matrixClient.HSUrl = options.HSUrl
	matrixClient.User = options.User
	matrixClient.Password = options.Password
	if "" == options.AccessToken {
		if err := matrixClient.Login(); nil != err {
			return nil, err
		}
	} else {
		matrixClient.LoginRes = &LoginRes{
			AccessToken: options.AccessToken,
		}
		userId, _ := matrixClient.WhoAmI()
		matrixClient.UserId = userId
	}
	return matrixClient, nil
}

func (client *MatrixClient) RequestUrl(req MatrixReq) string {
	paramsStr := ""
	params := req.Params()
	path := req.Path()
	if len(params) > 0 {
		for _, p := range params {
			if "" != p {
				if "" == paramsStr {
					paramsStr = p
				} else {
					paramsStr = paramsStr + "&" + p
				}
			}
		}
	}
	if nil == client.LoginRes || "" == client.LoginRes.AccessToken {
		return client.HSUrl + path + "?" + paramsStr
	} else {
		return client.HSUrl + path + "?access_token=" + client.LoginRes.AccessToken + "&" + paramsStr
	}
}

func (client *MatrixClient) Post(req MatrixReq, res interface{}) (errRes *ErrorRes, err error) {
	if reqData, err := json.Marshal(req); nil != err {
		return nil, err
	} else {
		client.mtx.Lock()
		resp, err := http.Post(client.RequestUrl(req), ContentType, bytes.NewReader(reqData))
		if err != nil {
			return nil, err
		}
		defer func() {
			resp.Body.Close()
			client.mtx.Unlock()
		}()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			errRes = &ErrorRes{}
			if err := json.Unmarshal(body, errRes); nil != err {
				log.Errorf("errbody:%s, err:%s", string(body), err.Error())
				return nil, fmt.Errorf("errbody:%s, err:%s", string(body), err.Error())
			} else {
				log.Errorf("error:%s, errorcode:%s", errRes.Error, errRes.ErrCode)
				return errRes, fmt.Errorf("error:%s, errorcode:%s", errRes.Error, errRes.ErrCode)
			}
		} else {
			if err := json.Unmarshal(body, res); nil != err {
				return nil, err
			}
			return nil, nil
		}
	}
}

func (client *MatrixClient) Put(req MatrixReq, res interface{}) (errRes *ErrorRes, err error) {
	if reqData, err := json.Marshal(req); nil != err {
		return nil, err
	} else {
		client.mtx.Lock()
		req, err := http.NewRequest("PUT", client.RequestUrl(req), bytes.NewReader(reqData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", ContentType)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() {
			resp.Body.Close()
			client.mtx.Unlock()
		}()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			errRes = &ErrorRes{}
			if err := json.Unmarshal(body, errRes); nil != err {
				log.Errorf("errbody:%s, err:%s", string(body), err.Error())
				return nil, fmt.Errorf("errbody:%s, err:%s", string(body), err.Error())
			} else {
				log.Errorf("error:%s, errorcode:%s", errRes.Error, errRes.ErrCode)
				return errRes, fmt.Errorf("error:%s, errorcode:%s", errRes.Error, errRes.ErrCode)
			}
		} else {
			if err := json.Unmarshal(body, res); nil != err {
				return nil, err
			}
			return nil, nil
		}
	}
}

func (client *MatrixClient) Get(req MatrixReq, res interface{}) (errRes *ErrorRes, err error) {
	client.mtx.Lock()
	resp, err := http.Get(client.RequestUrl(req))
	if err != nil {
		return nil, err
	}
	defer func() {
		resp.Body.Close()
		client.mtx.Unlock()
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		errRes = &ErrorRes{}
		if err := json.Unmarshal(body, errRes); nil != err {
			log.Errorf("errbody:%s, err:%s", string(body), err.Error())
			return nil, fmt.Errorf("errbody:%s, err:%s", string(body), err.Error())
		} else {
			log.Errorf("error:%s, errorcode:%s", errRes.Error, errRes.ErrCode)
			return errRes, fmt.Errorf("error:%s, errorcode:%s", errRes.Error, errRes.ErrCode)
		}
	} else {
		if err := json.Unmarshal(body, res); nil != err {
			return nil, err
		}
		return nil, nil
	}
}

func (client *MatrixClient) Login() error {
	loginReq := &LoginReq{
		Type:                     "m.login.password",
		User:                     client.User,
		Password:                 client.Password,
		DeviceId:                 "ip",
		InitialDeviceDisplayName: "ip",
	}
	loginRes := &LoginRes{}
	if _, err := client.Post(loginReq, loginRes); nil != err {
		return err
	}
	client.LoginRes = loginRes
	return nil
}

func (client *MatrixClient) CreateRoom() {

}

func (client *MatrixClient) Logout() {

}
func (client *MatrixClient) WhoAmI() (string, error) {
	req := &WhoAmIReq{}
	res := &WhoAmIRes{}
	errRes, err := client.Get(req, res)
	if nil != errRes && UNRECOGNISED_TOKEN_ERROR == errRes.Error {
		if err1 := client.Login(); nil == err1 {
			errRes, err = client.Get(req, res)
		}
	}
	return res.UserId, err
}

func (client *MatrixClient) JoinRoom(roomid string) error {
	req := &JoinRoomReq{
		RoomId: roomid,
	}
	res := &JoinRoomRes{}
	_, err := client.Post(req, res)
	log.Debugf("roomid:%s", res.RoomId)
	return err
}

func (client *MatrixClient) RoomInvite() {

}

func (client *MatrixClient) RoomMessages(roomId, from, to, dir, limit, filter string) (*RoomMessagesRes, error) {
	req := &RoomMessagesReq{
		RoomId: roomId,
		From:   from,
		To:     to,
		Dir:    dir,
		Limit:  limit,
		Filter: filter,
	}
	res := &RoomMessagesRes{}
	errRes, err := client.Get(req, res)
	if nil != errRes && UNRECOGNISED_TOKEN_ERROR == errRes.Error {
		if err1 := client.Login(); nil == err1 {
			errRes, err = client.Get(req, res)
		}
	}

	log.Debugf("start:%s, end:%s, len:%d", res.Start, res.End, len(res.Chunk))

	return res, err
}

func (client *MatrixClient) SendMessages(roomId, eventType, txnid, msgtype, body string) (string, error) {
	req := &SendMessageReq{
		RoomId:    roomId,
		EventType: eventType,
		TxnId:     txnid,
		MsgType:   msgtype,
		Body:      body,
	}
	res := &SendMessageRes{}
	errRes, err := client.Put(req, res)
	if nil != errRes && UNRECOGNISED_TOKEN_ERROR == errRes.Error {
		if err1 := client.Login(); nil == err1 {
			errRes, err = client.Put(req, res)
		}
	}
	log.Debugf("eventId:%s", res.EventId)
	return res.EventId, err
}

func (client *MatrixClient) RoomInviteFilter() {

}
