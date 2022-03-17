package main

import (
	"context"
	"fmt"
	"grpc-test/logger"
	"grpc-test/message"
	"grpc-test/pb/protogo"
	"io"
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
	wg     *sync.WaitGroup
	logger *zap.SugaredLogger
}

func NewMsgClient(logger *zap.SugaredLogger) *MsgClient {
	return &MsgClient{
		wg:     new(sync.WaitGroup),
		logger: logger,
	}
}

func (c *MsgClient) StartToServerClient() bool {
	c.logger.Infof("start message client")

	conn, err := c.NewClientConn()
	if err != nil {
		fmt.Printf("fail to create connection: %s", err)
		return false
	}

	stream, err := GetMSGToServerClientStream(conn)
	if err != nil {
		fmt.Printf("fail to get connection stream: %s", err)
		return false
	}
	sendTimes := 10000000

	//msg := message.MakeOneBMessage()
	msg := message.MakeOneKBMessage()
	//msg := message.MakeFourKBMessage()
	//msg := message.MakeOneMBMessage()
	msgChan := make(chan *protogo.Message, 20000000)
	for j := 0; j < sendTimes/1000; j++ {
		go func() {
			for i := 0; i < 1000; i++ {
				msgChan <- msg
			}
		}()
	}
	c.wg.Add(1)
	startTime := time.Now()
	go c.toServerClientSendMsgRoutine(stream, sendTimes, msgChan)
	c.wg.Wait()

	size := sendTimes * msg.Size() / 1000000
	sec := int(time.Since(startTime).Seconds())

	c.logger.Infof("finish sending %d messages in time: %ds, total size: %dMB, messages per sec: %d, "+
		"speed: %dMB/s, tcpFlag: %v", sendTimes, sec, size, sendTimes/sec, size/sec, tcpFlag)

	return true
}

func (c *MsgClient) StartToClientClient() bool {
	c.logger.Infof("start message client")

	conn, err := c.NewClientConn()
	if err != nil {
		fmt.Printf("fail to create connection: %s", err)
		return false
	}

	stream, err := GetMSGToClientClientStream(conn)
	if err != nil {
		fmt.Printf("fail to get connection stream: %s", err)
		return false
	}

	sendTimes := 10000000

	//msg := message.MakeOneBMessage()
	msg := message.MakeOneKBMessage()
	//msg := message.MakeFourKBMessage()
	//msg := message.MakeOneMBMessage()
	c.wg.Add(1)
	startTime := time.Now()
	go c.toClientClientRecvMsgRoutine(stream, sendTimes)
	c.wg.Wait()

	size := sendTimes * msg.Size() / 1000000
	sec := int(time.Since(startTime).Seconds())

	c.logger.Infof("finish received %d messages in time: %ds, total size: %dMB, messages per sec: %d, "+
		"speed: %dMB/s, tcpFlag: %v", sendTimes, sec, size, sendTimes/sec, size/sec, tcpFlag)

	return true
}

func (c *MsgClient) StartTwoDirectionClient() bool {
	c.logger.Infof("start message client")

	conn, err := c.NewClientConn()
	if err != nil {
		fmt.Printf("fail to create connection: %s", err)
		return false
	}

	stream, err := GetMSGTwoDirectionClientStream(conn)
	if err != nil {
		fmt.Printf("fail to get connection stream: %s", err)
		return false
	}

	sendTimes := 10000000

	//msg := message.MakeOneBMessage()
	msg := message.MakeOneKBMessage()
	//msg := message.MakeFourKBMessage()
	//msg := message.MakeOneMBMessage()
	msgChan := make(chan *protogo.Message, 20000000)
	for j := 0; j < sendTimes/1000; j++ {
		go func() {
			for i := 0; i < 1000; i++ {
				msgChan <- msg
			}
		}()
	}
	c.wg.Add(2)
	go c.twoDirectionClientSendMsgRoutine(stream, sendTimes, msgChan)
	go c.twoDirectionClientRecvMsgRoutine(stream, sendTimes)
	c.wg.Wait()

	return true
}

func (c *MsgClient) toServerClientSendMsgRoutine(stream protogo.MessageRPC_MessageToServerClient, sendTimes int,
	msgChan chan *protogo.Message) {
	defer c.wg.Done()

	//msg1 := message.MakeOneKBMessage()
	c.logger.Infof("start sending goroutine")

	for i := 0; i < sendTimes; i++ {
		msg1 := <-msgChan
		err := stream.Send(msg1)
		if i%5000000 == 0 {
			fmt.Printf("msg chan len:%d\n", len(msgChan))
		}
		if err != nil {
			c.logger.Errorf("fail to send one: %s", err)
			return
		}
	}
}

func (c *MsgClient) toClientClientRecvMsgRoutine(stream protogo.MessageRPC_MessageToClientClient, recvTimes int) {
	defer c.wg.Done()
	c.logger.Infof("start receiving client message ")
	for i := 0; i < recvTimes; i++ {
		_, revErr := stream.Recv()

		if revErr == io.EOF {
			c.logger.Errorf("client receive eof and exit receive goroutine")
			return
		}

		if revErr != nil {
			c.logger.Errorf("client receive err and exit receive goroutine %s", revErr)
			return
		}
	}
}

func (c *MsgClient) twoDirectionClientSendMsgRoutine(stream protogo.MessageRPC_MessageTwoDirectionClient, sendTimes int,
	msgChan chan *protogo.Message) {
	startTime := time.Now()
	msg := message.MakeOneKBMessage()
	size := sendTimes * msg.Size() / 1000000
	defer func() {
		sec := int(time.Since(startTime).Seconds())
		c.logger.Infof("finish tow direction sending %d messages in time: %ds, total size: %dMB, messages per sec: %d, "+
			"speed: %dMB/s, tcpFlag: %v", sendTimes, sec, size, sendTimes/sec, size/sec, tcpFlag)
		c.wg.Done()
	}()

	c.logger.Infof("start sending goroutine")

	for i := 0; i < sendTimes; i++ {
		msg1 := <-msgChan
		err := stream.Send(msg1)
		if i%5000000 == 0 {
			fmt.Printf("msg chan len:%d\n", len(msgChan))
		}
		if err != nil {
			c.logger.Errorf("fail to send one: %s", err)
			return
		}
	}
}

func (c *MsgClient) twoDirectionClientRecvMsgRoutine(stream protogo.MessageRPC_MessageTwoDirectionClient, recvTimes int) {
	startTime := time.Now()
	msg1 := message.MakeOneKBMessage()
	size := recvTimes * msg1.Size() / 1000000
	defer func() {
		sec := int(time.Since(startTime).Seconds())
		c.logger.Infof("finish tow direction sending %d messages in time: %ds, total size: %dMB, messages per sec: %d, "+
			"speed: %dMB/s, tcpFlag: %v", recvTimes, sec, size, recvTimes/sec, size/sec, tcpFlag)
		c.wg.Done()
	}()
	c.logger.Infof("start receiving client message ")
	for i := 0; i < recvTimes; i++ {
		_, revErr := stream.Recv()
		if revErr == io.EOF {
			c.logger.Errorf("client receive eof and exit receive goroutine")
			return
		}

		if revErr != nil {
			c.logger.Errorf("client receive err and exit receive goroutine %s", revErr)
			return
		}
	}
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

// GetMSGToServerClientStream get rpc stream
func GetMSGToServerClientStream(conn *grpc.ClientConn) (protogo.MessageRPC_MessageToServerClient, error) {
	return protogo.NewMessageRPCClient(conn).MessageToServer(context.Background())
}

// GetMSGToClientClientStream get rpc stream
func GetMSGToClientClientStream(conn *grpc.ClientConn) (protogo.MessageRPC_MessageToClientClient, error) {
	return protogo.NewMessageRPCClient(conn).MessageToClient(context.Background())
}

// GetMSGTwoDirectionClientStream get rpc stream
func GetMSGTwoDirectionClientStream(conn *grpc.ClientConn) (protogo.MessageRPC_MessageTwoDirectionClient, error) {
	return protogo.NewMessageRPCClient(conn).MessageTwoDirection(context.Background())
}

func main() {
	log := logger.NewLogger("client", "INFO")
	for i := 0; i < 4; i++ {
		client := NewMsgClient(log)
		go client.StartToClientClient()
	}
	for i := 0; i < 4; i++ {
		client := NewMsgClient(log)
		go client.StartToServerClient()
	}
	time.Sleep(1000000000000)
	for i := 0; i < 4; i++ {
		client := NewMsgClient(log)
		go client.StartTwoDirectionClient()
	}
	time.Sleep(1000000000000)
}
