# flood
Is a OSI layer 2 attack to take down a switch by filling the MAC-Address table.

## Requirements
This software only runs on linux operating systems and requires minimum version of `go1.12`.

## Build
With an properly configured Go toolchain execute the following.

```bash
# on linux only
go get github.com/davidkroell/flood
cd $GOPATH/src/github.com/davidkroell/flood

go build -o flood main.go
```

## Usage
```bash
Usage of flood:
  -i string
    	Interface to send
  -n int
    	Amount of frames send (default 1)
  -s int
    	Seed for source MAC address
  -t int
    	Number of threads to use (default 12)
  -v	Print version
```

## Disclaimer
This software is provided for educational use only.
The authors are not responsible for any misuse of the software.
Performing an attack without permission from the owner of the network is illegal.
Use at your own risk.
