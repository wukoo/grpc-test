gen-pb:
	cd pb/proto && protoc -I=. --gogofaster_out=plugins=grpc:../protogo --gogofaster_opt=paths=source_relative message.proto