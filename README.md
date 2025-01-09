# Paddleball
Multi-session, fixed rate, network quality monitor.

Paddleball measures the network quality by sending streams of packets to a server
that bounces the packets back, analyzing the round-trip time and packet loss.

## Usage

```
./paddleball -h                      
Usage of ./paddleball:
  -V    print version info
  -b int
        payload size (default 384)
  -d string
        file name of a DuckDB database where to record all measurements (requires paddleball with DuckDB support compiled in)
  -e    gather extended stats, like percentiles
  -j    print in JSON format
  -k int
        server key
  -n int
        number of clients/servers to run (default 1)
  -r int
        client pps rate (default 10)
  -s    set server mode
  -t string
        tag to use in logging
```

## Building

Build static binary, without any dependencies. Binary is placed under build/ directory.

```
make build
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

## Storing all measurements in a DuckDB database

Paddleball can be compiled with DuckDB support. This is not done by default, because DuckDB requires support for loading system C libraries.

```
make buildquack
```

After that you can use -d switch, that will cause all measurements and the final aggregation to be stored in local DuckDB database.

```
paddleball -k 1 -e -r 100 -d test.duckdb -t test_20250109T0907 127.0.0.1:12234

$ duckdb test.duckdb
v1.1.3 19864453f7
Enter ".help" for usage hints.
D show tables;
┌─────────────┐
│    name     │
│   varchar   │
├─────────────┤
│ aggregation │
│ measurement │
└─────────────┘

D select * from aggregation;
┌──────────────────────┬────────────────────┬─────────────────┬────────────────┬──────────────────┬──────────────────┬────────────┬───┬────────────┬─────────┬──────────┬──────────────────────┬───────────────┬─────────────────┐
│      EventTime       │        Tag         │ ReceivedPackets │ DroppedPackets │ DuplicatePackets │ ReorderedPackets │ AverageRTT │ … │ HighestRTT │ P90RTT  │  P99RTT  │ PBQueueDroppedPack…  │ PBQueueLength │ PBQueueCapacity │
│ timestamp with tim…  │      varchar       │      int64      │     int64      │      int64       │      int64       │   double   │   │   double   │ double  │  double  │        int64         │     int32     │      int32      │
├──────────────────────┼────────────────────┼─────────────────┼────────────────┼──────────────────┼──────────────────┼────────────┼───┼────────────┼─────────┼──────────┼──────────────────────┼───────────────┼─────────────────┤
│ 2025-01-09 08:14:3…  │ test_20250109T0907 │           31499 │              0 │                0 │                0 │   0.761648 │ … │   7.684439 │ 1.16905 │ 1.325396 │                    0 │             0 │               0 │
├──────────────────────┴────────────────────┴─────────────────┴────────────────┴──────────────────┴──────────────────┴────────────┴───┴────────────┴─────────┴──────────┴──────────────────────┴───────────────┴─────────────────┤
│ 1 rows                                                                                                                                                                                                   14 columns (13 shown) │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Paddlequack

Check out the included paddlequack package, which provides a utility for recording all paddleball metrics into a DuckDB database for future analytics.
It is a separate binary, that ingests paddbleball JSON logs.
