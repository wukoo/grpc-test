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

//func MakeOneBMessage() *protogo.Message {
//	return &protogo.Message{
//		Msg: oneB,
//	}
//}

func MakeOneKBMessage() *protogo.Message {
	para := make(map[string][]byte)
	para["ddddd"] = []byte("ddddddd")
	para["dddfd"] = []byte("ddddddd")
	para["dddsdd"] = []byte("ddddddd")
	para["ddggggdd"] = []byte("ddddddd")
	para["fffffff"] = []byte("ddddddd")
	para["aaaaaaa"] = []byte("ddddddd")
	para["uuuu"] = []byte("ddddddd")
	para["xxxxx"] = []byte("ddddddd")
	para["mmmmmmm"] = []byte("ddddddd")
	para["ggggg"] = []byte("ddddddd")
	para["hhhhhhhhhh"] = []byte("ddddddd")

	return &protogo.Message{
		TxId:       "xxxxxfffaasdasdfasdfasdfasdfasdfasdfasdfasdfasasdfasdfasdf",
		Type:       0,
		ResultCode: 0,
		Payload:    []byte(oneKB),
		Message:    "",
		TxRequest: &protogo.TxRequest{
			TxId:            "xxxxxfffaasdasdfasdfasdfasdfasdfasdfasdfasdfasasdfasdfasdf",
			ContractName:    "safasdfasdfasdf",
			ContractVersion: "1.0.0",
			Method:          "ddddddd",
			Parameters:      para,
			TxContext: &protogo.TxContext{
				CurrentHeight:       0,
				WriteMap:            nil,
				ReadMap:             nil,
				OriginalProcessName: "",
			},
		},
		TxResponse: nil,
	}
}

//func MakeFourKBMessage() *protogo.Message {
//	return &protogo.Message{
//		Msg: fourKB,
//	}
//}
//
//func MakeOneMBMessage() *protogo.Message {
//	return &protogo.Message{
//		Msg: oneMB,
//	}
//}
