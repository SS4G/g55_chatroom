package server

import (
	"errors"
	"g55.com/chat/config"
	"g55.com/chat/model"
	"g55.com/chat/util"
	log "github.com/sirupsen/logrus"
	"time"
)

type ChatRoomServer struct {
	tcpServer        *util.TCPNetServer
	userLoginAddrMap map[uint64]string
	userPasswdMap    map[uint64]string
	recvChannel      chan *model.Message
}

func NewChatRoomServer() *ChatRoomServer {
	return &ChatRoomServer{
		tcpServer:        &util.TCPNetServer{},
		userLoginAddrMap: make(map[uint64]string),
		userPasswdMap:    make(map[uint64]string),
		recvChannel:      make(chan *model.Message, 100),
	}
}

func (srv *ChatRoomServer) Serve() {
	err := srv.tcpServer.Open(config.SERVER_IP_PORT)
	//recvChan := make(chan *model.Message, 100)

	if err != nil {
		log.Errorf("a")
	}
	go srv.AcceptConnectionProcess()
	go srv.RouterProcess()
	for {
		time.Sleep(time.Second * 5)
		//og.Infof("[Server] heart beat %lld", time.Now().Unix())
	}
}

func (srv *ChatRoomServer) RouterProcess() {
	for {
		msgP := <-srv.recvChannel //recvChannel
		log.Infof("[Server] IRecv msg=%s", msgP.ToString())
		switch msgP.Type {
		case model.TYPE_REGISTER:
			srv.registerUser(msgP)
		case model.TYPE_LOGIN:
			srv.loginUser(msgP)

		case model.TYPE_LOGOUT:
			srv.logoutUser(msgP)

		case model.TYPE_MASSAGE:
			log.Debugf("[Server] msg type")
			err := srv.routeMsg(msgP)
			if err != nil {
				log.Errorf("[Server] err=%v", err)
			}

		default:
			log.Errorf("[Server] invalid")
		}
	}
}

func (srv *ChatRoomServer) AcceptConnectionProcess() {
	for {
		remoteAddr, connErr := srv.tcpServer.Accept()
		if connErr != nil {
			log.Errorf("f")
		}
		go srv.tcpServer.ServerChannelRecvMsg(remoteAddr, srv.recvChannel) //tcpServerP.sendMsgTest(remoteAddr, "connections build! response from server!")
	}
}

// TODO: 添加持久化逻辑
func (srv *ChatRoomServer) registerUser(msg *model.Message) {
	srv.userPasswdMap[msg.FromUser.UserId] = msg.FromUser.PassWordMD5
}

// TODO: 登录map 暂未支持 并发冲突
func (srv *ChatRoomServer) loginUser(msg *model.Message) {
	srv.userLoginAddrMap[msg.FromUser.UserId] = msg.SrcAddr
}

//
func (srv *ChatRoomServer) logoutUser(msg *model.Message) {
	delete(srv.userLoginAddrMap, msg.FromUser.UserId)
}

// TODO: 暂时只支持 同时在线用户通信
func (srv *ChatRoomServer) routeMsg(msg *model.Message) error {
	//fromUid := msg.FromUser.UserId
	toUid := msg.ToUser.UserId
	if toUserAddr, ok := srv.userLoginAddrMap[toUid]; !ok {
		log.Errorf("[Server] to user not found or not login")
		return errors.New("to user not found or not login")
	} else {
		log.Debugf("[Server] msg send to %s", toUserAddr)
		sendErr := srv.tcpServer.ServerSendMsg(toUserAddr, msg)
		return sendErr
	}

}
