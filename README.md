
# TCP Echo Server

## Features
- 4-byte big-endian length-prefixed messages  
- Correct partial reads/writes  
- Concurrency-limited connection handling  
- Graceful shutdown on signal  
- Max frame size check  

## Run
**Server**
```sh
go run ./cmd/server
```

**Client**
```sh
go run ./cmd/client
```

## Protocol
```
[4 bytes length][payload]
```

