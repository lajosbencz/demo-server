# demo-server

## Run from binary

```bash
./demo-server -help

Usage of ./demo-server:
  -f string
        Persist resource state to this file (leave empty to disable)
  -h string
        Host part of address to listen on (default "localhost")
  -p int
        Port part of address to listen on (default 8080)
  -s    Enable HTTPS with self-signed certificate
```

## Run from container
```bash
docker run -v ${PWD}:/etc/demo-server -p 8080:8080 ghcr.io/lajosbencz/demo-server:latest
```
