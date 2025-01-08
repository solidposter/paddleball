# Paddleball
Multi-session, fixed rate, network quality monitor.

Paddleball measures the network quality by sending streams of packets to a server
that bounces the packets back, analyzing the round-trip time and packet loss.

## Usage

```
./paddleball -h                      
Usage of ./paddleball:
  -V	print version info
  -b int
    	payload size (default 384)
  -e	gather extended stats, like percentiles
  -j	print in JSON format
  -k int
    	server key
  -n int
    	number of clients/servers to run (default 1)
  -r int
    	client pps rate (default 10)
  -s	set server mode
  -t string
    	tag to use in logging
```

## Basic use
Example server running on port 10000 (single port):
paddleball -s -k 1984 10000

Example client using above server:
paddleball -k 1984 192.168.1.100:10000

## High performance test
Paddleball scales across multiple cores with the -n option

Example server using 4 ports to load over 4 cores:
paddleball -s -k 1984 -n 4 10000

Example client hitting it with 400 25pps clients (10kpps):
paddleball -k 1984 -r 25 -n 400 192.168.100:10000

## Paddlequack

Check out the included paddlequack package, which provides a utility for recording all paddleball metrics into a DuckDB database for future analytics.
