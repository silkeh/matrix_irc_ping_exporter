# Matrix/IRC ping bot and Prometheus exporter
This little project rolls three things into one:

- A Matrix bot that responds to ping messages.
- An IRC bot that responds to ping messages.
- A prometheus exporter that sends a message from the Matrix bot to the IRC bot,
  and measures the latency.
  
## Usage

### Matrix
The Matrix bot responds to the following command:

```
!ping [message]
```

The response is human-readable, with the metadata set in the message.
This command mirrors the functionality of [maubot/echo][].

### IRC
The IRC bot responds to ping commands of the following format:

```
ping [id] [unix time in ns]
```

The response is of the format:

```
pong <id> <unix time in ns> [delay in ns] [human-readable delay]
```

The default `id` is `unixnano`.

### Prometheus


## Installation
Download and build the program using:

```
go get -u github.com/silkeh/matrix_irc_ping_exporter/cmd/matrix_irc/ping_exporter
```

See `config.dist.yaml` for an example configuration.

[maubot/echo]: https://github.com/maubot/echo
