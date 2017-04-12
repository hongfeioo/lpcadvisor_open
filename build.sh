#!/bin/bash
go build run.go  mylog.go
[ $? -ne 0 ] && exit 1

go build uploadCadvisorData.go pushDatas.go mylog.go getDatas.go dataFunc.go  SomeStruct.go  cadvisorApiv1.go
[ $? -ne 0 ] && exit 1

docker stop micadvisor ; docker rm micadvisor
docker build -t micadvisor ./
docker run     --volume=/:/rootfs:ro     --volume=/var/run:/var/run:rw     --volume=/sys:/sys:ro     --volume=/home/work/log/cadvisor/:/home/work/uploadCadviosrData/log     --volume=/data/docker/containers:/data/docker/containers:ro     --publish=18080:18080     --env Interval=60     --detach=true     --name=micadvisor     --net=host     --restart=always     micadvisor:latest
#tail -f  /home/work/log/cadvisor/*
