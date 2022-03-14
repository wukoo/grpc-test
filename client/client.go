package main

import (
	"context"
	"fmt"
	"grpc-test/logger"
	"grpc-test/message"
	"grpc-test/pb/protogo"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

const (
	port        = "16666"
	tcpFlag     = true
	MaxSendSize = 10
	MaxRevSize  = 10
)

type MsgClient struct {
	stream protogo.MessageRPC_MessageChatClient
	stop   chan struct{}
	wg     *sync.WaitGroup
	logger *zap.SugaredLogger
}

func NewMsgClient(clientId string) *MsgClient {
	return &MsgClient{
		stream: nil,
		stop:   make(chan struct{}),
		wg:     new(sync.WaitGroup),
		logger: logger.NewLogger(clientId, "INFO"),
	}
}

func (c *MsgClient) StartClient() bool {
	c.logger.Infof("start message client")

	conn, err := c.NewClientConn()
	if err != nil {
		fmt.Printf("fail to create connection: %s", err)
		return false
	}

	stream, err := GetMSGClientStream(conn)
	if err != nil {
		fmt.Printf("fail to get connection stream: %s", err)
		return false
	}

	c.stream = stream

	sendTimes := 10000000

	//msg := message.MakeOneBMessage()
	//msg := message.MakeOneKBMessage()
	msg := message.MakeFourKBMessage()
	//msg := message.MakeOneMBMessage()
	stopChan := make(chan struct{})
	msgChan := make(chan *protogo.Message, 20000000)
	for j := 0; j < sendTimes/1000; j++ {
		go func() {
			for i := 0; i < 1000; i++ {
				msgChan <- msg
			}
		}()
	}
	c.wg.Add(2)
	startTime := time.Now()
	go c.sendMsgRoutine(msg, sendTimes, stopChan, msgChan)
	go c.revMsgRoutine(stopChan)
	c.wg.Wait()

	size := sendTimes * msg.Size() / 1000000
	sec := int(time.Since(startTime).Seconds())

	c.logger.Infof("finish sending %d messages in time: %ds, total size: %dMB, messages per sec: %d, "+
		"speed: %dMB/s, tcpFlag: %v", sendTimes, sec, size, sendTimes/sec, size/sec, tcpFlag)

	return true
}

func (c *MsgClient) sendMsgRoutine(msg *protogo.Message, sendTimes int, stopChan chan struct{},
	msgChan chan *protogo.Message) {
	defer func() {
		c.wg.Done()
		stopChan <- struct{}{}
	}()

	c.logger.Infof("start sending goroutine")

	for i := 0; i < sendTimes; i++ {
		msg1 := <-msgChan
		err := c.stream.Send(msg1)
		if i%5000000 == 0 {
			fmt.Printf("msg chan len:%d\n", len(msgChan))
		}
		if err != nil {
			c.logger.Errorf("fail to send one: %s", err)
			return
		}
	}
}

func (c *MsgClient) revMsgRoutine(stopChan chan struct{}) {
	defer c.wg.Done()
	c.logger.Infof("start receiving client message ")
	<-stopChan
	//for {
	//msg, revErr := c.stream.Recv()
	//
	//if revErr == io.EOF {
	//	clientLogger.Errorf("client receive eof and exit receive goroutine")
	//	close(c.stop)
	//	c.wg.Done()
	//	return
	//}
	//
	//if revErr != nil {
	//	clientLogger.Errorf("client receive err and exit receive goroutine %s", revErr)
	//	close(c.stop)
	//	c.wg.Done()
	//	return
	//}
	//
	//clientLogger.Infof("client receive [%v]", msg)

	//count++
	//}
}

// NewClientConn create rpc connection
func (c *MsgClient) NewClientConn() (*grpc.ClientConn, error) {

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxRevSize*1024*1024),
			grpc.MaxCallSendMsgSize(MaxSendSize*1024*1024),
		),
	}
	if tcpFlag {
		ip := "127.0.0.1"
		url := fmt.Sprintf("%s:%s", ip, port)

		return grpc.Dial(url, dialOpts...)
	} else {
		dialOpts = append(dialOpts, grpc.WithContextDialer(func(ctx context.Context, sock string) (net.Conn, error) {
			unixAddress, _ := net.ResolveUnixAddr("unix", "/tmp/grpc_test.sock")
			conn, err := net.DialUnix("unix", nil, unixAddress)
			return conn, err
		}))
		return grpc.Dial("/tmp/grpc_test.sock", dialOpts...)
	}

}

// GetMSGClientStream get rpc stream
func GetMSGClientStream(conn *grpc.ClientConn) (protogo.MessageRPC_MessageChatClient, error) {
	return protogo.NewMessageRPCClient(conn).MessageChat(context.Background())
}

func main() {
	clientId := "client1"
	client := NewMsgClient(clientId)
	client.StartClient()
}
