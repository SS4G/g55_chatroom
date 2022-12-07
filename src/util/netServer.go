package util

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"g55.com/chat/model"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

// 网络管道 底层可以收发数据
type NetServer interface {
	// 打开关闭
	// 打开对应 ${ip}:${port} 的端口
	Open(ip_port string) error
	// 接受链接
	Accept() (string, error)
	// 关闭客户端
	Close() error
	// 测试连通性
	Ping(target_client string) bool

	// 收发数据
	// 向对应的客户端发送数据
	serverSendData(raddr string, data []byte) (int, error)
	// 发送消息
	ServerSendMsg(raddr string, data *model.Message) error

	// 接受来自服务端的消息 底层需要使用bufio对消息按照协议解析
	serverRecvData(raddr string) ([]byte, error)
	//
	ServerRecvMsg(raddr string) (*model.Message, error)
	// 从channel中接受来自服务端 解析好的msg
	SeverChannelRecvMsg() (<-chan *model.Message, error)

	// 获取检查链接
	GetClientsAddrs() []string
	PingClients() map[string]bool
}

type TCPNetServer struct {
	port           string //127.0.0.1:${port}
	listener       *net.TCPListener
	clientConnsMap *sync.Map //map[string]net.Conn // TODO: 需要使用 sync.Map
}

func (s *TCPNetServer) Open(ip_port string) error {
	s.port = ip_port
	s.clientConnsMap = &sync.Map{} //make(map[string]net.Conn)
	tcpAddr, rsvErr := net.ResolveTCPAddr("tcp", s.port)
	if rsvErr != nil {
		log.Errorf("[Server] TCP resolve laddr Err\n")
	}
	// 监听 对应的端口 这里用的是laddr 因为Listen不可能监听远端的ip 端口 所以这里能用到的只有端口
	listener, lisErr := net.ListenTCP("tcp", tcpAddr)
	s.listener = listener
	if lisErr != nil {
		return lisErr
	}
	return nil
}

func (s *TCPNetServer) Accept() (raddr string, accErr error) {
	conn, accErr := s.listener.Accept()
	if accErr != nil {
		log.Errorf("[Server] accept fail\n")
		return "", accErr
	}
	raddr = conn.RemoteAddr().String()
	log.Infof("[Server] remoteAddr Got %s", raddr)
	s.clientConnsMap.Store(raddr, conn)
	return raddr, nil
}

// net.Conn 底层实际上是TCPConn* 所以是指针类型可以直接复制的
func (s *TCPNetServer) getClientConn(remoteAddr string) net.Conn {
	if v, ok := s.clientConnsMap.Load(remoteAddr); ok {
		if connP, vok := v.(net.Conn); vok {
			return connP
		}
	}
	return nil
}

func (c *TCPNetServer) serverSendData(raddr string, data []byte) (int, error) {
	tcpConn := c.getClientConn(raddr)
	if tcpConn != nil {
		writeN, err := tcpConn.Write(data)
		if err != nil {
			log.Errorf("[Server] write bytes Error")
			panic("server write bytes Error")
		}
		if writeN != len(data) {
			log.Errorf("[Server] write bytes NumError")
			return writeN, errors.New("client write Error")
		}
		return writeN, nil
	}
	return 0, errors.New(fmt.Sprintf("raddr %s not Found", raddr))
}

func (c *TCPNetServer) ServerSendMsg(raddr string, msg *model.Message) error {
	_, err := c.serverSendData(raddr, msg.ToBytes())
	return err
}

// 从接受端读取字节
func (c *TCPNetServer) severRecvData(raddr string) ([]byte, error) {
	tcpConn := c.getClientConn(raddr)
	if tcpConn != nil {
		log.Debugf("[Server] wating for reading bufio")
		bufReader := bufio.NewReader(tcpConn)
		rdBytes, err := bufReader.ReadBytes(model.DELIM_BYTE)

		if err != nil {
			log.Errorf("[Server] bufio err=%v", err)
		}
		//log.Infof("[Server] A recv bytes=%d", len(rdBytes))
		//log.Infof("[Server] A recv string=%s", string(rdBytes))
		if len(rdBytes) > 0 && rdBytes[len(rdBytes)-1] == model.DELIM_BYTE {
			rdBytes = rdBytes[:len(rdBytes)-1]
		}
		log.Debugf("[Server] recv bytes=%d", len(rdBytes))
		log.Debugf("[Server] recv string=%s", string(rdBytes))

		return rdBytes, err
	}
	return nil, errors.New(fmt.Sprintf("raddr %s not Found", raddr))
}

// 从接受端读取字节
//func (c *TCPNetServer) severSimpleRecvData(raddr string) ([]byte, error) {
//	tcpConn := c.getClientConn(raddr)
//	if tcpConn != nil {
//		buffer := make([]byte, 4096)
//		readN, err := tcpConn.Read(buffer)
//		log.Infof("[Server] recv bytes %d\n", readN)
//		//rdBytes, err := bufReader.ReadBytes(model.DELIM_BYTE)
//		//log.Infof("[Server] recv bytes=%d", len(rdBytes))
//		return buffer, err
//	}
//	return nil, errors.New(fmt.Sprintf("raddr %s not Found", raddr))
//}

// 从接受端读取字节并解析成对应的 Message
func (c *TCPNetServer) ServerRecvMsg(raddr string) (*model.Message, error) {
	msgP := &model.Message{}
	msgBytes, err := c.severRecvData(raddr)
	if err != nil {
		return nil, err
	}
	// Tips: json 反序列化的时候第二个参数必须是指针类型
	jsErr := json.Unmarshal(msgBytes, msgP)
	if jsErr != nil {
		return nil, jsErr
	}
	// 标记来源addr
	msgP.SrcAddr = raddr
	return msgP, nil
}

// 从对应地址制定的ip处读取
func (c *TCPNetServer) ServerChannelRecvMsg(raddr string, recvChannel chan *model.Message) {
	//msgChan := make(chan *model.Message, 128)
	go func() {
		// 退出时关闭对应的链接 并从字典中删除
		defer func() {
			c.getClientConn(raddr).Close()
			c.clientConnsMap.Delete(raddr)
		}() // Notice 需要注意这里需要defer语句调用
		for {
			msgP, err := c.ServerRecvMsg(raddr)
			if err != nil {
				log.Errorf("[Server] ServerRecvMsg err=%v", err)
				break
			}
			// TODO:  捕获err
			// 不管从什么渠道收到的信息都写入到recvChannel中 需要中央处理逻辑根据From和To来做转发
			recvChannel <- msgP
		}
	}()
	return
}

func ServerTestMain() {
	//tcpServerP := &TCPNetServer{}
	//err := tcpServerP.Open(config.SERVER_IP_PORT)
	//	recvCh := make(chan *model.Message, 100)
	//	if err != nil {
	//	log.Errorf("a")
	//}

	//go AcceptConnectionProcess(tcpServerP, recvCh)
	//go MainServerProcess(recvCh)
	//time.Sleep(time.Second * 1000)
}
