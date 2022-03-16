package main

import (
	"grpc-test/message"
	"grpc-test/pb/protogo"
	"io"
	"sync"
	"time"

	"go.uber.org/atomic"
)

type API struct {
	index *atomic.Int32
	wg    *sync.WaitGroup
}

func NewAPI() *API {
	return &API{
		index: atomic.NewInt32(0),
		wg:    new(sync.WaitGroup),
	}
}

func (a *API) MessageTwoDirection(server protogo.MessageRPC_MessageTwoDirectionServer) error {
	index := a.index.Add(1)
	Logger.Infof("got a new stream, index: %d", index)

	sendTimes := 10000000
	recvCount := 0
	a.wg.Add(2)
	startTime := time.Now()

	go a.msgTwoDirectionServerRecvMsgRoutine(server, index, &recvCount)
	go a.msgTwoDirectionServerSendMsgRoutine(server, index, sendTimes)

	a.wg.Wait()

	//size := recvCount * message.MakeOneBMessage().Size() / 1000000
	//size := recvCount * message.MakeOneKBMessage().Size() / 1000000
	size := recvCount * message.MakeFourKBMessage().Size() / 1000000
	//size := recvCount * message.MakeOneMBMessage().Size() / 1000000
	sec := int(time.Since(startTime).Seconds())

	Logger.Infof("stream %d recv %d msgs in time %d, total size: %dMB, messages per sec: %d, speed %dMB/s, "+
		"send %d messages in time:%d, messages per sec: %d, speed %dMB/s, tcpFlag: %v",
		index, recvCount, sec, size, recvCount/sec, size/sec, sendTimes, sec, sendTimes/sec, size/sec, tcpFlag)

	return nil
}

func (a *API) MessageToServer(server protogo.MessageRPC_MessageToServerServer) error {
	index := a.index.Add(1)
	Logger.Infof("got a new stream, index: %d", index)

	recvCount := 0
	a.wg.Add(1)
	startTime := time.Now()

	go a.msgToServerServerRecvMsgRoutine(server, index, &recvCount)

	a.wg.Wait()

	//size := recvCount * message.MakeOneBMessage().Size() / 1000000
	//size := recvCount * message.MakeOneKBMessage().Size() / 1000000
	size := recvCount * message.MakeFourKBMessage().Size() / 1000000
	//size := recvCount * message.MakeOneMBMessage().Size() / 1000000
	sec := int(time.Since(startTime).Seconds())

	Logger.Infof("stream %d recv %d msgs in time %d, total size: %dMB, messages per sec: %d, speed %dMB/s, "+
		"tcpFlag: %v", index, recvCount, sec, size, recvCount/sec, size/sec, tcpFlag)

	return nil
}

func (a *API) MessageToClient(server protogo.MessageRPC_MessageToClientServer) error {
	index := a.index.Add(1)
	Logger.Infof("got a new stream, index: %d", index)

	sendTimes := 10000000
	recvCount := 0
	a.wg.Add(1)
	startTime := time.Now()

	go a.msgToClientServerSendMsgRoutine(server, index, sendTimes)

	a.wg.Wait()

	//size := recvCount * message.MakeOneBMessage().Size() / 1000000
	//size := recvCount * message.MakeOneKBMessage().Size() / 1000000
	size := recvCount * message.MakeFourKBMessage().Size() / 1000000
	//size := recvCount * message.MakeOneMBMessage().Size() / 1000000
	sec := int(time.Since(startTime).Seconds())

	Logger.Infof("stream %d recv %d msgs in time %d, total size: %dMB, messages per sec: %d, speed %dMB/s, "+
		"send %d messages in time:%d, messages per sec: %d, speed %dMB/s, tcpFlag: %v",
		index, recvCount, sec, size, recvCount/sec, size/sec, sendTimes, sec, sendTimes/sec, size/sec, tcpFlag)

	return nil
}

func (a *API) msgToServerServerRecvMsgRoutine(stream protogo.MessageRPC_MessageToServerServer,
	index int32, count *int) {
	defer a.wg.Done()

	Logger.Infof("stream %d start receiving cdm message", index)
	for {
		_, revErr := stream.Recv()
		if revErr == io.EOF {
			Logger.Infof("stream %d receive eof and exit receive goroutine", index)
			return
		}

		if revErr != nil {
			Logger.Infof("stream %d receive err and exit receive goroutine, error: %s", index, revErr)
			return
		}
		*count++
	}
}

func (a *API) msgToClientServerSendMsgRoutine(stream protogo.MessageRPC_MessageToClientServer,
	index int32, sendTimes int) {
	defer a.wg.Done()

	Logger.Infof("stream %d start sending goroutine", index)
	msg := message.MakeFourKBMessage()
	for i := 0; i < sendTimes; i++ {
		err := stream.Send(msg)
		if err != nil {
			Logger.Errorf("fail to send one: %s", err)
			return
		}
	}
}

func (a *API) msgTwoDirectionServerRecvMsgRoutine(stream protogo.MessageRPC_MessageTwoDirectionServer,
	index int32, count *int) {
	defer a.wg.Done()

	Logger.Infof("stream %d start receiving cdm message", index)
	for {
		_, revErr := stream.Recv()
		if revErr == io.EOF {
			Logger.Infof("stream %d receive eof and exit receive goroutine", index)
			return
		}

		if revErr != nil {
			Logger.Infof("stream %d receive err and exit receive goroutine, error: %s", index, revErr)
			return
		}
		*count++
	}
}

func (a *API) msgTwoDirectionServerSendMsgRoutine(stream protogo.MessageRPC_MessageTwoDirectionServer,
	index int32, sendTimes int) {
	defer a.wg.Done()

	Logger.Infof("stream %d start sending goroutine", index)
	msg := message.MakeFourKBMessage()
	for i := 0; i < sendTimes; i++ {
		err := stream.Send(msg)
		if err != nil {
			Logger.Errorf("fail to send one: %s", err)
			return
		}
	}
}
