
protoc  --go_out=. --go-grpc_out=. post.proto

mongodb+srv://dbvitor:<password>@cryptocluster.fdk7o.mongodb.net/myFirstDatabase?retryWrites=true&w=majority

go run crypto_server/server.go
go run crypto_client/client.go

go mod init github.com/Vitordb/crypto
go mod tidy
