//+build linux

//Copyright (C) 2019  David Kröll
//
//This program is free software: you can redistribute it and/or modify
//it under the terms of the GNU General Public License as published by
//the Free Software Foundation, either version 3 of the License, or
//(at your option) any later version.
//
//This program is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//GNU General Public License for more details.

package main

import (
	"flag"
	"fmt"
	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/raw"
	"net"
	"os/user"
	"sync"
	"time"
)

var num = flag.Int("n", 1, "Amount of frames send")
var ifaceName = flag.String("i", "", "Interface to send")
var numThreads = flag.Int("t", 12, "Number of threads to use")
var seed = flag.Int("s", 0, "Seed for source MAC address")
var versionFlag = flag.Bool("v", false, "Print version")

const etherType = 0xbeef
const version = "flood v0.2.0"

func main() {
	flag.Parse()
	if !prerequisitesSatisfied() {
		return
	}

	iface, err := net.InterfaceByName(*ifaceName)
	if err != nil {
		fmt.Println("No such network interface")
		return
	}

	conn, err := raw.ListenPacket(iface, etherType, nil)
	if err != nil {
		fmt.Printf("cannot open connection: %v", err)
		return
	}

	var wg sync.WaitGroup

	// init channels
	ch := make(chan *ethernet.Frame)
	stats := make(chan int)

	// create sender goroutines
	for i := 0; i < *numThreads; i++ {
		wg.Add(1)
		// pass in Done() method from waitgroup
		go frameWriter(conn, ch, stats, wg.Done)
	}

	// init stat vars
	framesSend := 0
	bytesWritten := 0
	startTime := time.Now()

	// stat collecting goroutine
	go func() {
		// no need for waitgroup here,
		// goroutine gets automatically killed when any sender goroutine exits
		for bytes := range stats {
			bytesWritten += bytes
			framesSend++
		}
	}()

	for i := 1; i <= *num; i++ {
		f := &ethernet.Frame{
			Destination: ethernet.Broadcast,
			Source: net.HardwareAddr{
				byte(*seed),
				0x00,
				// hacky method for power to 2 numbers
				byte(i / (24 << 1)), uint8(i / (16 << 1)), uint8(i / (8 << 1)), uint8(i),
			},
			EtherType: etherType,
		}
		ch <- f
	}

	// close channel
	close(ch)
	wg.Wait() // wait for goroutines quit

	fmt.Println("Execution summary:")
	fmt.Printf("%d frames send\n", framesSend)
	fmt.Printf("%d bytes written\n", bytesWritten)
	fmt.Printf("Took a total time of %v\n", time.Since(startTime))
}

func frameWriter(c net.PacketConn, ch <-chan *ethernet.Frame, stats chan<- int, doneCall func()) {
	for f := range ch {
		// get frame and marshall it to binary
		b, err := f.MarshalBinary()
		if err != nil {
			fmt.Printf("failed to marshal ethernet frame: %v", err)
		}

		// only necessary for WriteTo() method, does not change the frame
		addr := &raw.Addr{
			HardwareAddr: ethernet.Broadcast,
		}

		// write frame
		n, err := c.WriteTo(b, addr)
		if err != nil {
			fmt.Printf("Cannot write to connection: %v", err)
		}

		// send to channel
		stats <- n
	}
	doneCall()
}

func prerequisitesSatisfied() bool {
	if *versionFlag {
		fmt.Println(version)
		return false
	}

	u, _ := user.Current()
	if u.Uid != "0" {
		fmt.Println("This program requires root (UID 0) access")
		return false
	}

	// require iface index flag
	if *ifaceName == "" {
		flag.Usage()
		return false
	}

	if *seed < 0 || *seed > 255 {
		fmt.Printf("Seed (%d) must be between 0 and 255\n", *seed)
		return false
	}
	return true
}
