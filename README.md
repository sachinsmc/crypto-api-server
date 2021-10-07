# Crypto API server

```
GOOS=linux GOARCH=amd64 go build -o crypto-api-server-linux .

GOOS=linux GOARCH=amd64 go build -o crypto-api-server-mac . 

GOOS=windows GOARCH=amd64 go build -o crypto-api-server-windows.exe .
```

# How to run

To run from macos just run binary, you check other untested binaries

`$ ./crypto-api-server`

To run source code : 

`$ go run main.go`



# Used libraries

    1. Gorilla WebSocket : implementation of the WebSocket
    2. Gorilla Mux : HTTP request multiplexer, implements a request router and dispatcher for matching incoming requests to their respective handler
    3. juju/errors :  provides an easy way to annotate errors without losing the original error context
    4. juju/testing : This package provides additional base test suites to be used with gocheck# crypto-api-server
