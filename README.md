# Gossip Gloomers

[Gossip Gloomers](https://fly.io/dist-sys/) challenge is a series of distributed systems challenges by Fly.io.
This repository contains my solutions to these challenges in Go using [maelstrom-go](https://github.com/jepsen-io/maelstrom/blob/main/demo/go/node.go) library.
I am doing these challenges to sharpen my Go skills and learn about distributed systems basics.

## Prerequisites

### Go

```bash
brew install go
```

### OpenJDK, graphviz, gnuplot

```bash
brew install openjdk graphviz gnuplot
# add openjdk binaries to PATH
echo 'export PATH="/opt/homebrew/opt/openjdk/bin:$PATH"' >> ~/.zshrc
```

### Maelstrom

The solutions are validated using [maelstrom](https://github.com/jepsen-io/maelstrom) workbench.
Download [Maelstrom v0.2.3](https://github.com/jepsen-io/maelstrom/releases/tag/v0.2.3) and unpack it to run `maelstrom`
binary from the folder directly.

## Challenges

### maelstrom-echo

This is a warmup [challenge](https://fly.io/dist-sys/1/) where node returns back the same message as it received.
[Solution](https://github.com/anmarchenko/gossip-gloomers/blob/370f569235aed2b95992185ee9e0bebb07ae1548/maelstrom-echo/main.go) is straightforward: return the same body but replace message type with "echo_ok".

Run it:

```bash
cd maelstrom-echo
go install .

# use your maelstrom location
cd ~/maelstrom
./maelstrom test -w echo --bin ~/go/bin/maelstrom-echo --node-count 1 --time-limit 10
```

### maelstrom-unique-ids

In [this challenge](https://fly.io/dist-sys/2/) you'll need to implement a globally-unique ID generation system.
In my [solution](https://github.com/anmarchenko/gossip-gloomers/blob/370f569235aed2b95992185ee9e0bebb07ae1548/maelstrom-unique-ids/main.go) I used [Snowflake ID](https://en.wikipedia.org/wiki/Snowflake_ID) algorithm for ID generation.

Snowflake ID is a 64-bit integer that contains:

- 41 bits for timestamp (in milliseconds since arbitrary chosen epoch)
- 10 bits for machine id (I used NodeID provided by maelstrom-go library)
- 12 bits are for machine sequence number to avoid collisions when generating multiple IDs per millisecond

JS pitfall I fell into when solving this challenge: javascript (and thus JSON) decimals are floats and 64 bit integers cannot be represented in JSON precisely. As maelstrom uses JSON as data format, the ID we return must be a string to avoid precision issues

Run it:

```bash
cd maelstrom-unique-ids
go install .

# use your maelstrom location
cd ~/maelstrom
./maelstrom test -w unique-ids --bin ~/go/bin/maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
```

### maelstrom-broadcast

[This challenge](https://fly.io/dist-sys/3a/) is about implementing a broadcast system that gossips messages between all nodes in the cluster.

The system must handle the following RPC calls:

- **broadcast** - requests that a value be broadcast out to all nodes in the cluster. The node should store the "message" value locally so it can be read later
- **read** - requests that a node return all values that it has seen
- **topology** - informs the node of who its neighboring nodes are

There is another RPC call I added in order to gossip messages between nodes:

- **propagate** - requests that a value will be added to a set of messages seen by the node

I implemented a simple [solution](https://github.com/anmarchenko/gossip-gloomers/blob/be07dd76ed4205c30667b27c08bc8ed546f030a9/maelstrom-broadcast/main.go) for this challenge: the messages are being propagated to every node in a cluster once a node receives a new message:

```mermaid
graph TB:
  A(0)-->B(1)
  A(0)-->C(2)
  A(0)-->N(N)
```

This is not an optimized solution but it solves the problem in this case.

Run it:

```bash
cd maelstrom-broadcast
go install .

# use your maelstrom location
cd ~/maelstrom
./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10
```

### maelstrom-counter

TBD
