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
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var ( // Populated at build time.
	version string
	date    string
)

func main() {
	modePtr := flag.Bool("s", false, "set server mode")
	keyPtr := flag.Int("k", 0, "server key")
	clntPtr := flag.Int("n", 1, "number of clients/servers to run")
	ratePtr := flag.Int("r", 10, "client pps rate")
	bytePtr := flag.Int("b", 384, "payload size")
	jsonPtr := flag.String("j", "text", "print in JSON format with (not the word text)")
	versPtr := flag.Bool("v", false, "print version info")
	flag.Parse()

	if *versPtr {
		fmt.Println("Version:", version)
		fmt.Println("Date:", date)
		os.Exit(0)
	}

	// start in server mode
	if *modePtr {
		if len(flag.Args()) != 1 {
			fmt.Println("Please specify server base port as the final option")
			os.Exit(1)
		}
		port := flag.Args()[0]
		lport, err := strconv.Atoi(port)
		if err != nil {
			fmt.Println("Invalid port", port)
			os.Exit(1)
		}
		if lport < 1 || lport > 65535 {
			fmt.Println("Invalid port", port)
			os.Exit(1)
		}
		serverkey := int64(*keyPtr)
		if serverkey == 0 {
			serverkey = rand.Int63()
		}
		hport := lport + *clntPtr - 1
		fmt.Printf("Starting in server mode on port %v", lport)
		for i := lport; i <= hport; i++ {
			go server(strconv.Itoa(i), serverkey, lport, hport)
		}
		if lport == hport {
			fmt.Printf(" with key %v\n", serverkey)
		} else {
			fmt.Printf("-%v with key %v\n", hport, serverkey)
		}
		<-(chan int)(nil) // wait forever
	}

	// client mode
	if len(flag.Args()) != 1 {
		fmt.Println("Final and only argument must be IP:port")
		os.Exit(1)
	}
	if *keyPtr == 0 {
		fmt.Println("Specify server key")
		os.Exit(1)
	}
	if *ratePtr < 1 {
		fmt.Println("client rate below 1 pps not supported")
		os.Exit(1)
	}

	// Global information and statistics
	global := packetStats{}
	// catch CTRL+C
	go trapper(&global)

	// start statistics engine
	rp := make(chan payload, (*ratePtr)*(*clntPtr)*2) // buffer return payload up to two second
	go statsEngine(rp, &global, *jsonPtr)
	// Send a probe to get server configuration
	if *jsonPtr == "text" {
		fmt.Printf("Starting probe of %v", flag.Args()[0])
	}
	p := newclient(65535)
	lport, hport := p.probe(flag.Args()[0], *keyPtr)
	if *jsonPtr == "text" {
		fmt.Printf(" ports %v-%v active\n", lport, hport)
	}

	// Extract IP address from the IP:port string
	ip, err := net.ResolveUDPAddr("udp", flag.Args()[0])
	if err != nil {
		log.Panic(err)
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
	<-(chan int)(nil) // wait forever
}

func trapper(global *packetStats) {
	cs := make(chan os.Signal, 2)
	signal.Notify(cs, os.Interrupt, syscall.SIGTERM)
	<-cs

	fmt.Println()
	statsPrint(global, "text", 0, 0) // no need for JSON here
	fmt.Println()
	os.Exit(0)
}
