# ORPC

ORPC is built for Zai Objects Transport:

API fixed, easy to tuning performance.

## Performance

The reasons to implement a new RPC protocol but not using GRPC/HTTP like others is we do care about performance.

### OTCP

In present, we only have implementation based on TCP. 

#### Protocol

Protocol is simple, and we can regard it as a user-defined version of HTTP:

1. Header
2. Body

We don't waste any byte in header for efficiency of network I/O.

No compression on body, because the body may have been compressed, and it makes data integrity checking getting complex.

#### Optimization

About what I've done, you could find in [otcp README](otcp/README.md).

## Data Integrity

ORPC will check the integrity of both the header and body.

## TLS

There is no need to use TLS, removing all parts of TLS help to tidy codes.
