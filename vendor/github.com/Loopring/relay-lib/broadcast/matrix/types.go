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
	"encoding/json"
	"fmt"
)

const (
	ContentType      = "application/json"
	LoginPath        = "/_matrix/client/r0/login"
	JoinRoomPath     = "/_matrix/client/r0/rooms/%s/join"
	RoomMessagesPath = "/_matrix/client/r0/rooms/%s/messages"
	CreateFilterPath = "/_matrix/client/r0/user/%s/filter"
	SendMessagePath  = "/_matrix/client/r0/rooms/%s/send/%s/%s"
	WhoAmIPath       = "/_matrix/client/r0/account/whoami"
)

const (
	UNRECOGNISED_TOKEN_ERROR = "Unrecognised access token."
)

const (
	LoopringOrderType = "io.loopring.order"
)

type MatrixReq interface {
	Path() string
	Params() []string
}

type LoginReq struct {
	Type                     string `json:"type"`
	User                     string `json:"user"`
	Password                 string `json:"password"`
	Token                    string `json:"token"`
	DeviceId                 string `json:"device_id"`
	InitialDeviceDisplayName string `json:"initial_device_display_name"`
}

func (req *LoginReq) Path() string {
	return LoginPath
}

func (req *LoginReq) Params() []string {
	return []string{}
}

type LoginRes struct {
	UserId      string `json:"user_id"`
	AccessToken string `json:"access_token"`
	HomeServer  string `json:"home_server"`
	DeviceId    string `json:"device_id"`
}

type ErrorRes struct {
	ErrCode string `json:"errcode"`
	Error   string `json:"error"`
}

type JoinRoomReq struct {
	RoomId     string `json:"roomId"`
	Sender     string `json:"sender"`
	MxId       string `json:"mxid"`
	Token      string `json:"token"`
	Signatures string `json:"signatures"`
}

func (req *JoinRoomReq) Path() string {
	return fmt.Sprintf(JoinRoomPath, req.RoomId)
}

func (req *JoinRoomReq) Params() []string {
	return []string{}
}

type JoinRoomRes struct {
	RoomId string `json:"room_id"`
}

type RoomMessagesReq struct {
	RoomId string `json:"roomId"`
	From   string `json:"from"`
	To     string `json:"to"`
	Dir    string `json:"dir"`
	Limit  string `json:"limit"`
	Filter string `json:"filter"`
}

func (req *RoomMessagesReq) Path() string {
	return fmt.Sprintf(RoomMessagesPath, req.RoomId)
}

func (req *RoomMessagesReq) Params() []string {
	params := []string{}
	if "" != req.From {
		params = append(params, "from="+req.From)
	}
	if "" != req.To {
		params = append(params, "to="+req.To)
	}
	if "" != req.Dir {
		params = append(params, "dir="+req.Dir)
	}
	if "" != req.Limit {
		params = append(params, "limit="+req.Limit)
	}
	if "" != req.Filter {
		params = append(params, "filter="+req.Filter)
	}
	return params
}

type RoomMessagesRes struct {
	Start string      `json:"start"`
	End   string      `json:"end"`
	Chunk []RoomEvent `json:"chunk"`
}

type RoomEvent struct {
	Event
	EventId        string       `json:"event_id"`
	RoomId         string       `json:"room_id"`
	Sender         string       `json:"sender"`
	OriginServerTs int          `json:"origin_server_ts"`
	Unsigned       UnsignedData `json:"unsigned"`
}

type UnsignedData struct {
	Age             int    `json:"age"`
	RedactedBecause Event  `json:"redacted_because"`
	TransactionId   string `json:"transaction_id"`
}

type Event struct {
	Content MsgContent `json:"content"`
	Type    string     `json:"type"`
}

func (evt Event) Parse(res json.Unmarshaler) (interface{}, error) {
	if _, err := json.Marshal(evt.Content); nil != err {
		return nil, err
	} else {
		return nil, nil
	}
}

type FilterReq struct {
	UserId      string     `json:"-"`
	EventFields []string   `json:"event_fields,omitempty"`
	EventFormat string     `json:"event_format,omitempty"`
	Presence    Filter     `json:"presence,omitempty"`
	AccountData Filter     `json:"account_data,omitempty"`
	Room        RoomFilter `json:"room,omitempty"`
}

type Filter struct {
	Limit      int      `json:"limit"`
	NotSenders []string `json:"not_senders,omitempty"`
	NotTypes   []string `json:"not_types,omitempty"`
	Senders    []string `json:"senders,omitempty"`
	Types      []string `json:"types,omitempty"`
}

type RoomFilter struct {
	NotRooms     []string        `json:"not_rooms,omitempty"`
	Rooms        []string        `json:"rooms,omitempty"`
	Ephemeral    RoomEventFilter `json:"ephemeral,omitempty"`
	IncludeLeave bool            `json:"include_leave,omitempty"`
	State        RoomEventFilter `json:"state,omitempty"`
	Timeline     RoomEventFilter `json:"timeline,omitempty"`
	AccountData  RoomEventFilter `json:"account_data,omitempty"`
}

type RoomEventFilter struct {
	Limit      int      `json:"limit"`
	NotSenders []string `json:"not_senders,omitempty"`
	NotTypes   []string `json:"not_types,omitempty"`
	Senders    []string `json:"senders,omitempty"`
	Types      []string `json:"types,omitempty"`
	NotRooms   []string `json:"not_rooms,omitempty"`
	Rooms      []string `json:"rooms,omitempty"`
	ContainUrl bool     `json:"contains_url,omitempty"`
}

type MsgContent struct {
	Body    string `json:"body"`
	MsgType string `json:"msgtype"`
}

func (req *FilterReq) Path() string {
	return fmt.Sprintf(CreateFilterPath, req.UserId)
}

func (req *FilterReq) Params() []string {
	return []string{}
}

type FilterRes struct {
	FilterId string `json:"filter_id"`
}

type SendMessageReq struct {
	RoomId    string `json:"-"`
	EventType string `json:"-"`
	TxnId     string `json:"-"`
	MsgType   string `json:"msgtype"`
	Body      string `json:"body"`
}

func (req *SendMessageReq) Path() string {
	return fmt.Sprintf(SendMessagePath, req.RoomId, req.EventType, req.TxnId)
}

func (req *SendMessageReq) Params() []string {
	return []string{}
}

type SendMessageRes struct {
	EventId string `json:"event_id"`
}

type WhoAmIReq struct {
}

func (req *WhoAmIReq) Path() string {
	return WhoAmIPath
}

func (req *WhoAmIReq) Params() []string {
	return []string{}
}

type WhoAmIRes struct {
	UserId string `json:"user_id"`
}
