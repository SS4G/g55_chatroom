package model

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type MSG_TYPE int

const (
	TYPE_MASSAGE  MSG_TYPE = 0
	TYPE_REGISTER MSG_TYPE = 1
	TYPE_LOGIN    MSG_TYPE = 2
	TYPE_LOGOUT   MSG_TYPE = 3
)

const DELIM_BYTE = byte('$')

type UserInfo struct {
	UserName    string `json:"user_name"`
	UserId      uint64 `json:"user_id"`
	PassWordMD5 string `json:"passwd_md5"`
}

type Message struct {
	MsgData         string   `json:"msg_data"`
	Type            MSG_TYPE `json:"type"`
	FromUser        UserInfo `json:"from_user"`
	ToUser          UserInfo `json:"to_user"`
	ServerTimeStamp int64    `json:"server_timestamp"`
	SrcAddr         string   `json:"src_addr"`
	DstAddr         string   `json:"dst_addr"`
}

func (m *Message) ToString() string {
	b, err := json.Marshal(m)
	if err == nil {
		return string(b)
	}
	return `{"error": "MarshalError"}`
}

func (m *Message) ToBytes() []byte {
	b, err := json.Marshal(m)
	byte_buf := bytes.Buffer{}
	if err == nil {
		byte_buf.Write(b)
		byte_buf.WriteByte(DELIM_BYTE)
	} else {
		byte_buf.WriteString(`{"error": "MarshalError"}`)
		byte_buf.WriteByte(DELIM_BYTE)
	}
	return byte_buf.Bytes()
}

func NewUser(userName, passwd string, uid uint64) *UserInfo {
	passwdMD5 := fmt.Sprintf("%x", md5.Sum([]byte(passwd)))

	return &UserInfo{
		UserName:    userName,
		PassWordMD5: passwdMD5,
		UserId:      uid,
	}
}

func NewChatMessage(fromUser, toUser *UserInfo, msgData string) *Message {
	return &Message{
		FromUser:        *fromUser, // 只带有uid即可
		ToUser:          *toUser,   // 只带有uid即可
		Type:            TYPE_MASSAGE,
		MsgData:         msgData,
		ServerTimeStamp: time.Now().Unix(),
	}
}

func NewRegisterMessage(userId uint64, passwd string) *Message {
	passwdMD5 := fmt.Sprintf("%x", md5.Sum([]byte(passwd)))

	return &Message{
		FromUser:        UserInfo{UserId: userId, PassWordMD5: passwdMD5},
		ToUser:          UserInfo{},
		Type:            TYPE_REGISTER,
		MsgData:         "",
		ServerTimeStamp: time.Now().Unix(),
	}
}

func NewLoginMessage(user *UserInfo) *Message {
	return &Message{
		FromUser:        *user,
		ToUser:          UserInfo{},
		Type:            TYPE_LOGIN,
		MsgData:         "",
		ServerTimeStamp: time.Now().Unix(),
	}
}

func NewLogoutMessage(userId uint64) *Message {
	return &Message{
		FromUser:        UserInfo{UserId: userId},
		ToUser:          UserInfo{},
		Type:            TYPE_LOGOUT,
		MsgData:         "",
		ServerTimeStamp: time.Now().Unix(),
	}
}

func RandomGenUser() *UserInfo {
	user_names := []string{"songziheng.666", "zhaoyanbin.zyb", "wuyu.wy01", "liuyang.x"}
	user_ids := []uint64{9283111123, 7178381423, 193892391, 131998419}
	passwds := []string{"asdhkakdu131*", "asdnk18o(", "138819had", "adkahdu1"}
	randIdx := rand.Intn(4)
	passwdMD5 := fmt.Sprintf("%x", md5.Sum([]byte(passwds[randIdx])))
	return &UserInfo{
		UserName:    user_names[randIdx],
		PassWordMD5: passwdMD5,
		UserId:      user_ids[randIdx],
	}
}

func RandomGenMessage() *Message {
	msg_datas_tmp := []string{
		"Hello %s!",
		"Fuck Off %s!",
		"How Are you %s",
	}

	randUsers := []*UserInfo{RandomGenUser(), RandomGenUser(), RandomGenUser(), RandomGenUser()}

	types := []MSG_TYPE{TYPE_MASSAGE, TYPE_REGISTER, TYPE_LOGIN, TYPE_LOGOUT}
	fromUser := randUsers[rand.Intn(len(randUsers))]
	toUser := randUsers[rand.Intn(len(randUsers))]
	msg := fmt.Sprintf(msg_datas_tmp[rand.Intn(len(msg_datas_tmp))], toUser.UserName)

	return &Message{
		MsgData:         msg,
		Type:            types[rand.Intn(len(types))],
		FromUser:        *fromUser,
		ToUser:          *toUser,
		ServerTimeStamp: time.Now().Unix(),
	}
}

func ModelTestRun() {
	log.Infof("user:%+v", *RandomGenUser())
	log.Infof("message:%+v", *RandomGenMessage())
	msgP := RandomGenMessage()
	log.Infof("message:%+v", *msgP)
	log.Infof("message_json:%+s", string(msgP.ToBytes()))
}
