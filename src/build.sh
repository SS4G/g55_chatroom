#!/bin/zsh
# cd /Users/bytedance/Desktop/my_proj/g55_chatroom/src
# go mod init g55.com/chat
export GO111MODULE=on
export GOPATH="/Users/bytedance/go:/Users/bytedance/Desktop/my_proj/g55_chatroom"
#go build -o main_run *.go
go build -o server_run ./server_main.go
go build -o client_run ./client_main.go
go build -o db_run ./db_main.go

if test $? -eq 0;
then
    echo "compile done"
    #./main_run
else
    echo "compile error"
fi
chmod 777 server_run client_run db_run
