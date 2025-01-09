package main

//
// Copyright (c) 2019 Tony Sarendal <tony@polarcap.org>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
//

import (
	"flag"
	"fmt"
	"github.com/influxdata/tdigest"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var version string // Populated at build time

func main() {
	modePtr := flag.Bool("s", false, "set server mode")
	keyPtr := flag.Int("k", 0, "server key")
	clntPtr := flag.Int("n", 1, "number of clients/servers to run")
	ratePtr := flag.Int("r", 10, "client pps rate")
	bytePtr := flag.Int("b", 384, "payload size")
	jsonPtr := flag.Bool("j", false, "print in JSON format")
	tagPtr := flag.String("t", "", "tag to use in logging")
	versPtr := flag.Bool("V", false, "print version info")
	quackPtr := flag.String("d", "", "file name of a DuckDB database where to record all measurements (requires paddleball with DuckDB support compiled in)")
	extendedStatsPtr := flag.Bool("e", false, "gather extended stats, like percentiles")
	flag.Parse()

	if *versPtr {
		fmt.Println("Version:", version)
		os.Exit(0)
	}

	slogSetup(*jsonPtr, *tagPtr)

	// start in server mode
	if *modePtr {
		if len(flag.Args()) != 1 {
			fatal("Please specify server port or ip:port as the final option")
		}

		baseipport := flag.Args()[0]
		var ipaddr, port string
		if parts := strings.Split(baseipport, ":"); len(parts) == 2 {
			ipaddr = parts[0]
			port = parts[1]
		} else {
			ipaddr = "0.0.0.0"
			port = baseipport
		}

		lport, err := strconv.Atoi(port)
		if err != nil {
			fatal("Invalid port: " + port)
		}
		if lport < 1 || lport > 65535 {
			fatal("Invalid port: " + port)
		}
		serverkey := int64(*keyPtr)
		if serverkey == 0 {
			serverkey = rand.Int63()
		}
		hport := lport + *clntPtr - 1
		for i := lport; i <= hport; i++ {
			ipport := ipaddr + ":" + strconv.Itoa(i)
			go server(ipport, serverkey, lport, hport)
		}
		if lport == hport {
			slog.Info("Starting in server mode", "key", serverkey, "lport", lport)
		} else {
			slog.Info("Starting in server mode", "key", serverkey, "lport", lport, "hport", hport)

		}
		<-(chan int)(nil) // wait forever
	}

	// client mode
	if len(flag.Args()) != 1 {
		fatal("Final and only argument must be IP:port")
	}
	if *keyPtr == 0 {
		fatal("Specify server key")
	}
	if *ratePtr < 1 {
		fatal("client rate below 1 pps not supported")
	}
	var quack *QuackStats
	if *quackPtr != "" {
		// Initialise quack (DuckDB resources)
		var err error
		quack, err = NewQuack(*quackPtr)
		if err != nil {
			slog.Error(err.Error())
			fatal("Error initialising DuckDB resources")
		}
		quack.Tag = *tagPtr
	}
	defer func() {
		// Close up quack resources
		if quack != nil {
			quack.Close()
		}
	}()

	// Global information and statistics
	global := packetStats{
		reportJSON: *jsonPtr,
	}
	if *extendedStatsPtr {
		// Initialize gathering quantiles for extended statistics
		global.quantiles = tdigest.New()
	}
	if quack != nil {
		global.quack = quack
	}
	// catch CTRL+C
	go trapper(&global)

	// start statistics engine
	rp := make(chan payload, (*ratePtr)*(*clntPtr)*2) // buffer return payload up to two second
	go statsEngine(rp, &global)
	// Send a probe to get server configuration
	if *jsonPtr {
		slog.Info("Starting probe", "target", flag.Args()[0])
	}
	p := newclient(65535)
	lport, hport := p.probe(flag.Args()[0], *keyPtr)
	if *jsonPtr {
		slog.Info("Ports active", "from", lport, "to", hport)
	}

	// Extract IP address from the IP:port string
	ip, err := net.ResolveUDPAddr("udp", flag.Args()[0])
	if err != nil {
		fatal(err.Error())
	}
	targetIP := ip.IP.String()
	// Set random port in the range, unless lport=hport (single port)
	targetPort := hport
	if hport != lport {
		targetPort = rand.Intn(hport-lport) + lport
	}

	// start the clients, staged over a second, iterating over server ports
	ticker := time.NewTicker(time.Duration(1000000/(*clntPtr)) * time.Microsecond)
	for i := 0; i < *clntPtr; i++ {
		c := newclient(i)
		go c.start(rp, targetIP, strconv.Itoa(targetPort), *keyPtr, *ratePtr, *bytePtr)
		if targetPort == hport {
			targetPort = lport
		} else {
			targetPort++
		}
		<-ticker.C
	}
	ticker.Stop()
	<-(chan int)(nil) // wait forever
}

func trapper(global *packetStats) {
	cs := make(chan os.Signal, 2)
	signal.Notify(cs, os.Interrupt, syscall.SIGTERM)
	<-cs

	fmt.Println()
	statsPrint(global, 0, 0, "globalstats")
	fmt.Println()
	if global.quack != nil {
		global.quack.Close()
	}
	os.Exit(0)
}

func fatal(s string) {
	slog.Error(s)
	os.Exit(1)
}
