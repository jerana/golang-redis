# golang-redis
This code package having  both minial functionality redis server  and its client. Both Server and client code and their executable(cross-compiled in darwin/amd64). If you have installed golang in your platform then you build both server and its redis_cli (client) from your platform itself.

## Running both server and clinet without compliation
1. server binary reside at ../golang-redis/ as server 
2. client binary at ../golang-redis/redis_cli/ as redis_cli
By default without given configuration, server run at "localhost" and listen at port "3333" but user can change this configuration by passing host and  port configuation during server run . 

i.e run server with help flag : ./server -h , this will give you how to sepecify  host and port configuration to server 
 ./server -h
Usage of ./server:
  -host string
    	Remote redis server HostIPs (default "localhost")
  -port string
    	Remote Redis server Listen port (default "3333")



3. By default client as redis_cli try to connect server at "localhost" and port "3333" but again you can specify server connection inforamtion during client startup :
./redis_cli -h
Usage of ./redis_cli:
  -host string
    	Remote redis server HostIPs (default "localhost")
  -port string
    	Remote Redis server Listen port (default "3333")


4. Start server  by running excuatable : ./server as follow:
./server
Listening on localhost:3333


5. open another shell and run client binary to connect server  ./redis_cli
 ./redis_cli
Connecting to redis Server at: localhost:3333
redis-cli>

6. start sending redis command request to server :  Make sure , all commands should be written in UpperCase letter 
```
redis-cli> SET x 1
OK
redis-cli> SET y 2
OK
redis-cli> GET x
1
redis-cli> GET y
2
redis-cli> SET x [1,3,4,]
OK
redis-cli> GET x
[1,3,4,]
redis-cli> DEL x
1
redis-cli> GET x
(nil)
redis-cli>
```

## Build and Running locally
Make sure that you have Go installed, and that it supports go modules.

```bash 
1. go into ../golang-redis directory 

2. go run server.go
```
Now go into another shell or window and complied and run client 
```
1/ go into ../golang-redis/redis_cli and do < go build > this will build resdis_cli binary and then you can run this cli client as ./redis_cli 
 ./redis_cli
Connecting to redis Server at: localhost:3333
redis-cli>


```

Allowed commands are `GET`, `SET`, `DEL`.