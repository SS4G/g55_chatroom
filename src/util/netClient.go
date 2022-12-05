package util

import (
	"bufio"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"model"
	"net"
)

type NetClient interface {
	// 打开关闭
	// 打开对应 ${ip}:${port} 的端口
	Open(ip_port string) error
	// 关闭客户端
	Close() error
	// 测试连通性
	Ping() bool

	// 收发数据
	// 发送对应的
	clientSendData(data []byte) (int, error)
	// 发送消息
	ClientSendMsg(data *model.Message) error

	// 接受来自服务端的消息 底层需要使用bufio对消息按照协议解析
	clientRecvData() ([]byte, error)
	//
	ClientRecvMsg() (*model.Message, error)
	// 从channel中接受来自服务端 解析好的msg
	ClientChannelRecvMsg() (<-chan *model.Message, error)

	//ClientSendData(msg *model.Message) error
	// 获取对应的channel 实现IO多路复用的接受数据
	//GetRecvChannel(bufsize int64) (<-chan model.Message, error)
	// 阻塞式接受数据
	//ReciveData() (*model.Message, error)

	// 获取检查链接
	//GetServerAddr() []string
}

type TCPNetCient struct {
	server_ip_port string
	recvChan       chan []model.Message
	tcpConn        net.Conn
}

func (c *TCPNetCient) Open(ip_port string) error {
	conn, err := net.Dial("tcp", ip_port)
	if err != nil {
		return err
	}
	c.tcpConn = conn
	return nil
}

func (c *TCPNetCient) Close() error {
	return c.tcpConn.Close()
}

func (c *TCPNetCient) Ping() error {
	return errors.New("unsupport")
}

func (c *TCPNetCient) clientSendData(data []byte) (int, error) {
	writeN, err := c.tcpConn.Write(data)
	if err != nil {
		log.Errorf("[Client] write bytes Error err=%v", err)
		panic("client write bytes Error")
	}
	if writeN != len(data) {
		log.Errorf("[Client] write bytes NumError")
		return writeN, errors.New("client write Error")
	}
	log.Infof("[Client] Send data success %d", writeN)
	return writeN, nil
}

func (c *TCPNetCient) ClientSendMsg(msg *model.Message) error {
	log.Infof("client sending %s", msg.ToString())
	_, err := c.clientSendData(msg.ToBytes())
	return err
}

func (c *TCPNetCient) clientRecvData() ([]byte, error) {
	bufReader := bufio.NewReader(c.tcpConn)
	rdBytes, err := bufReader.ReadBytes(model.DELIM_BYTE)
	if len(rdBytes) > 0 && rdBytes[len(rdBytes)-1] == model.DELIM_BYTE {
		rdBytes = rdBytes[:len(rdBytes)-1]
	}
	return rdBytes, err
}

func (c *TCPNetCient) ClientRecvMsg() (*model.Message, error) {
	msgP := &model.Message{}
	msgBytes, err := c.clientRecvData()
	log.Debugf("[Client] clientRecvData=%v", msgBytes)
	if err != nil {
		return nil, err
	}
	// Tips: json 反序列化的时候第二个参数必须是指针类型
	jsErr := json.Unmarshal(msgBytes, msgP)
	if jsErr != nil {
		return nil, jsErr
	}
	return msgP, nil
}

func (c *TCPNetCient) ClientChannelRecvMsg(msgChan chan *model.Message) error {
	go func() {
		for {
			msgP, err := c.ClientRecvMsg()
			if err != nil {
				break
			}
			msgChan <- msgP
		}
	}()
	return nil
}

func ClientTestMain() {
	//p := &TCPNetCient{}
	//cErr := p.Open(config.SERVER_IP_PORT)
	//if cErr != nil {
	//	log.Errorf("client fail")
	//}
	//
	//tcpMsgChan, err := p.ClientChannelRecvMsg()
	//if err != nil {
	//	log.Errorf("[Client] chanRecv failed")
	//	panic("[Client] channel error")
	//}
	//
	//cmdChan := p.CmdMsg()
	//
	//for {
	//	log.Infof("pleaseInput:\n")
	//	select {
	//	case tcpMsg := <-tcpMsgChan:
	//		fmt.Printf("[Client] recv msg %s\n", tcpMsg.ToString())
	//	case cmdMsg := <-cmdChan:
	//		fmt.Printf("[Client] recv cmd %s\n", cmdMsg)
	//		//p.clientSendData([]byte(cmdMsg))
	//		if strings.HasPrefix(cmdMsg, "Send") {
	//			randMsg := model.RandomGenMessage()
	//			err := p.ClientSendMsg(randMsg)
	//			if err != nil {
	//				log.Errorf("[Client] send msg err=%v", err)
	//			}
	//		}
	//	}
	//}
}
