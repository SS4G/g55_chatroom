package client

import (
	"bufio"
	"fmt"
	"g55.com/chat/config"
	"g55.com/chat/model"
	"g55.com/chat/util"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

type ChatRoomClient struct {
	tcpClient   *util.TCPNetCient
	currentUser *model.UserInfo
	tcpRecvChan chan *model.Message
	cmdChan     chan string
}

func NewChatRoomClient() *ChatRoomClient {
	p := &util.TCPNetCient{}
	cErr := p.Open(config.SERVER_IP_PORT)
	if cErr != nil {
		log.Errorf("client fail")
	}
	return &ChatRoomClient{
		tcpClient:   p,
		currentUser: nil,
		tcpRecvChan: make(chan *model.Message, 100),
		cmdChan:     make(chan string, 100),
	}
}

//login:uid=1 passwd=342
//login:uid=2 passwd=342

//send:dst_uid=2 msg=hellp
// 交互端 TODO: 图形化
func (cl *ChatRoomClient) CmdInput() {
	go func() {
		for {
			//_, err := fmt.Scanln(&scanStr)
			scanStr, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil && scanStr != "quit" {
				break
			}
			log.Debugf("got from stdin %s", scanStr)
			cl.cmdChan <- scanStr
		}
	}()
}

func (cl *ChatRoomClient) cmdProcess(cmd string) {
	sps := strings.Split(cmd, ":")
	cmdType, args := sps[0], sps[1]
	log.Debugf("cmdType=%s", cmdType)
	var dstUserID, fromUserId uint64
	var msg string
	var passwd string
	//var uname string
	switch strings.ToLower(cmdType) {
	case "send":
		log.Debug("send")
		fmt.Sscanf(args, "dst_uid=%d msg=%s", &dstUserID, &msg)
		log.Infof("[Client] dst_uid=%d msg=%s", dstUserID, msg)
		dstUserInfo := model.NewUser("null", "null", dstUserID)
		sendMsg := model.NewChatMessage(cl.currentUser, dstUserInfo, msg)
		log.Debugf("[Client] sendMsg=%s", sendMsg.ToString())
		sendErr := cl.tcpClient.ClientSendMsg(sendMsg)
		if sendErr != nil {
			log.Errorf("[Client] Send Err")
		}
	case "login":
		log.Debug("login")
		fmt.Sscanf(args, "uid=%d passwd=%s", &fromUserId, &passwd)
		log.Infof("[Client] uid=%d passwd=%s\n", fromUserId, passwd)
		cl.currentUser = model.NewUser(fmt.Sprintf("user_%u", fromUserId), passwd, fromUserId)
		loginMsg := model.NewLoginMessage(cl.currentUser)
		log.Debugf("[Client] sendMsg=%s", loginMsg.ToString())
		sendErr := cl.tcpClient.ClientSendMsg(loginMsg)
		if sendErr != nil {
			log.Errorf("[Client] Login send Err")
		}
	case "logout":
		log.Errorf("unsupport op logout")
	case "reg":
		log.Errorf("unsupport op reg")
	case "quit":
		log.Errorf("unsupport op quit")
	}
}

func (cl *ChatRoomClient) ClientRun() {
	recvErr := cl.tcpClient.ClientChannelRecvMsg(cl.tcpRecvChan)
	if recvErr != nil {
		log.Errorf("[Client] chanRecv failed")
		panic("[Client] channel error")
	}

	go cl.CmdInput()

	for {
		log.Infof("pleaseInput:\n")
		select {
		case tcpMsg := <-cl.tcpRecvChan:
			fmt.Printf("[Client] recv msg %s\n", tcpMsg.ToString())
		case cmdMsg := <-cl.cmdChan:
			log.Debugf("got cmdMsg %s", cmdMsg)
			cl.cmdProcess(cmdMsg)
		}
	}
}
