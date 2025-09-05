# NATS Session Setuper

This is the host part for https://github.com/lzaugg/nats-session-go and it should be running before participants are trying to setup their env.

The examples from nats-session-go will work only, if these 2 NATS services are running:

* `GopherService.next-gopher`: returns a unique username (e.g. `gopher-01`). 
    It's very handy to have a unique username per participant. This is used during the setup only.

* `GopherService.ping`: a simple ping service. 
    Is used to check if the connection and the setup is working.

**Important**: The server is setup with token based auth, to simplify the setup process. 
    All participants will use the same token and there are no security measures against misuse (intentional or unintentional) of the message broker!

## Prerequisite

* have `go` ready (and do a `go mod tidy` initially)

* have a `.env` file in the project root folder with the `NATS_SERVER` set.
    `NATS_SERVER` is the NATS URL to connect to and has the pattern: `nats://<user>:<token>@<server>:<port>`.
    Example:

    ```bash
    NATS_SERVER=nats://root:reallystrongpassword@nats01.mye.ch:4222
    ```

## Start

```bash
go run main.go
```