# ORPC

ORPC is built for Zai Objects Transport:

API fixed, easy to tuning performance.

## Performance

The reasons to implement a new RPC protocol but not using GRPC/HTTP like others is we do care about performance.

### OTCP

In present, we only have implementation based on TCP. 

### OUDP

It's based on UDP. 

The RTT testing made by `netperf` shows TCP maybe better or UDP cannot be better:

**UDP:**

```shell
➜  ~ netperf -H 10.188.33.13 -l 10 -t UDP_RR -v 2 -- -o min_latency,mean_latency,max_latency,stddev_latency,transaction_rate
```
|Minimum Latency Microseconds|Mean Latency Microseconds|Maximum Latency Microseconds|Stddev Latency Microseconds|Transaction Rate Tran/s|
|----|----|----|----|----|
|128|152.50|3186|20.99|6522.789|

**TCP:**

```
➜  ~ netperf -H 10.188.33.13 -l 10 -t TCP_RR -v 2 -- -o min_latency,mean_latency,max_latency,stddev_latency,transaction_rate
```
|Minimum Latency Microseconds|Mean Latency Microseconds|Maximum Latency Microseconds|Stddev Latency Microseconds|Transaction Rate Tran/s|
|----|----|----|----|----|
|122|145.88|919|15.11|6817.025|

I can't see the significant difference between TCP and UDP.

Before making a UDP version, we should do more researching.

#### Protocol

Protocol is simple, and we can regard it as a user-defined version of HTTP:

1. Header
2. Body

We don't waste any byte in header for efficiency of network I/O.

No compression on body, because the body may have been compressed, and it makes data integrity checking getting complex.

Protocol must be not modified unless could get huge benefit.

#### Optimization

About what I've done, you could find in [otcp README](otcp/README.md).

## Data Integrity

ORPC will check the integrity of both the header and body.

## TLS

There is no need to use TLS, removing all parts of TLS help to tidy codes.

## Why not make it general?

The network layer of ORPC is designed for performance sensitive application. It maybe a good news if we could use it
somewhere else.

The main issue is that it really hard to add/remove handler because the protocol limitation.
