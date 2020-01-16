# Multi-session, fixed rate, network quality monitor.

Paddleball measures the network quality by sending streams of packets to a server
that bounces the packets back, analyzing the round-trip time and packet loss.

## Starting the server with key 1984 and listing on port 2222:

paddleball -k 1984 -s 2222

Note that the server port is the final option. If server mode (-s) is specified without key or port,
key and port will be randomly chosen.

Server options
	-k <int>	server key
	-s		server mode



## Starting a client:

paddleball -k 1984 x.x.x.x:2222

Note that the ip:port is the final option.


Client options:
	-b <int>	payload size in bytes (not packet size, tcpdump is your friend)
	-j <string>	JSON output, for our logging system
	-k <int>	server key
	-n <int>	number of streams, default = 1
	-r <int>	pps rate per stream, default = 10

## Build and run a Docker image
Build the image
```bash
docker build -t paddleball .
```
Start the server
```bash
docker run -it --expose 2222 -p 2222:2222/udp paddleball -k 1984 -s 2222
```
Start the client on another host
```bash
docker run -it paddleball -k 1984 x.x.x.x:2222
```
### Start the server and client on the same host
If the server and client docker container are on the same host you need top put them on the same network to be able to communicate, first create the network then start the server and client
```bash
docker network create paddleballnet
docker run --name server --net paddleballnet -it --expose 2222 -p 2222:2222/udp paddleball -k 1984 -s 2222
docker run --name client --net paddleballnet -it paddleball -k 1984 server:2222
```