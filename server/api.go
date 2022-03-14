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

func (a *API) MessageChat(stream protogo.MessageRPC_MessageChatServer) error {
	index := a.index.Add(1)
	Logger.Infof("got a new stream, index: %d", index)
	stopChan := make(chan struct{})

	recvCount := 0
	a.wg.Add(2)
	startTime := time.Now()
	go a.revMsgRoutine(stream, index, &recvCount, stopChan)
	go a.sendMsgRoutine(stream, index, stopChan)

	a.wg.Wait()

	//size := recvCount * message.MakeOneBMessage().Size() / 1000000
	size := recvCount * message.MakeOneKBMessage().Size() / 1000000
	//size := recvCount * message.MakeOneMBMessage().Size() / 1000000
	sec := int(time.Since(startTime).Seconds())

	Logger.Infof("stream %d recv %d msgs in time %d, total size: %dMB, messages per sec: %d, speed %dMB/s, "+
		"tcpFlag: %v", index, recvCount, sec, size, recvCount/sec, size/sec, tcpFlag)

	return nil
}

func (a *API) revMsgRoutine(stream protogo.MessageRPC_MessageChatServer, index int32, count *int, stopChan chan struct{}) {
	defer func() {
		a.wg.Done()
		stopChan <- struct{}{}
	}()

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

func (a *API) sendMsgRoutine(stream protogo.MessageRPC_MessageChatServer, index int32, stopChan chan struct{}) {
	defer a.wg.Done()

	Logger.Infof("stream %d start sending goroutine", index)
	<-stopChan
}
