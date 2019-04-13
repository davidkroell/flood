# flood
Is a OSI layer 2 attack to take down a switch by filling the MAC-Address table.

## Requirements
This software only runs on linux operating systems and requires minimum version of `go1.12`.

## Build
With an properly configured Go toolchain execute the following.

```bash
# on linux only
go get github.com/davidkroell/mac-flooding
cd $GOPATH/src/github.com/davidkroell/mac-flooding

go build -o flood main.go
```

## Usage
```bash
Usage of flood:
  -i int
    	Interface to send (default 1)
  -l	List available interfaces
  -n int
    	Amount of ethernet frames send (default 1)
  -t int
    	Number of threads to use (default 12)

```

## Disclaimer
This software is provided for educational use only.
The authors are not responsible for any misuse of the software.
Performing an attack without permission from the owner of the network is illegal.
Use at your own risk.
