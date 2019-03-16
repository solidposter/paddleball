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
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	modePtr := flag.Bool("s", false, "set server mode")
	keyPtr := flag.Int("k", 0, "server key")
	clntPtr := flag.Int("n", 1, "number of clients to run")
	ratePtr := flag.Int("r", 10, "client pps rate")
	flag.Parse()

	// catch CTRL+C
	globalEngineInfo := engineInfo{}
	globalEngineInfo.minRtt = time.Duration(1*time.Hour)	// minRtt must not be zero
	go trapper(&globalEngineInfo)

	// start in server mode, flag.Args()[0] is port to listen on.
	if *modePtr {
		if len(flag.Args()) == 0 {
			server("0", *keyPtr)
		} else if len(flag.Args()) == 1 {
			server(flag.Args()[0],*keyPtr)
		} else {
			fmt.Println("Error, only the server port should follow the options.", flag.Args())
			os.Exit(1)
		}
	}

	// client mode
	if len(flag.Args()) == 0 {
		fmt.Println("Specify server:port")
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
	if len(flag.Args()) == 1 {
		fmt.Println("server address:", flag.Args()[0])
	} else {
		fmt.Println("Error, only server IP:port follow the options.", flag.Args())
		os.Exit(1)
	}

	// start statistics engine
	rp := make(chan payload, (*ratePtr)*(*clntPtr)*2 )	// buffer return payload up to two second
	globalEngineInfo.rate = *ratePtr
	globalEngineInfo.numClients = *clntPtr
	go statsEngine(rp, &globalEngineInfo)
	time.Sleep(20*time.Millisecond)		// give the statsengine time to init

	ticker := time.NewTicker(time.Duration(1000000/(*clntPtr)) * time.Microsecond)
	for i := 0; i < *clntPtr; i++ {
		go client(rp, i, flag.Args()[0], *keyPtr, *ratePtr)
		<- ticker.C
	}

//	for {
//		fmt.Println("statsenging channel capacity:", len(rp))
//		time.Sleep(time.Second)
//	}
	<-(chan int)(nil)	// wait forever
}

func trapper(globalEngineInfo *engineInfo) {
	cs := make(chan os.Signal)
	signal.Notify(cs, os.Interrupt, syscall.SIGTERM)
	<- cs
	fmt.Println("add some useful info about the total runtime")
	os.Exit(0)
}
