package message

import (
	"grpc-test/pb/protogo"
	"strings"
)

var oneB string
var oneKB string
var fourKB string
var oneMB string

func init() {
	oneB = "a"

	sb := strings.Builder{}
	for i := 0; i < 1024; i++ {
		sb.WriteString(oneB)
	}
	oneKB = sb.String()

	sbFourKB := strings.Builder{}
	for i := 0; i < 4; i++ {
		sbFourKB.WriteString(oneKB)
	}
	fourKB = sbFourKB.String()

	sbMB := strings.Builder{}
	for i := 0; i < 1024; i++ {
		sbMB.WriteString(oneKB)
	}
	oneMB = sbMB.String()
}

func MakeOneBMessage() *protogo.Message {
	return &protogo.Message{
		Msg: oneB,
	}
}

func MakeOneKBMessage() *protogo.Message {
	return &protogo.Message{
		Msg: oneKB,
	}
}

func MakeFourKBMessage() *protogo.Message {
	return &protogo.Message{
		Msg: fourKB,
	}
}

func MakeOneMBMessage() *protogo.Message {
	return &protogo.Message{
		Msg: oneMB,
	}
}
