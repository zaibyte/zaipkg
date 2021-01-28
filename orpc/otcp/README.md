# otcp

otcp implementation built on TCP.

Based on [goproc](https://github.com/valyala/gorpc) with these modifications:

1. No multi-methods supports, only has three methods: Put Object, Get Object, Delete Object.

2. Implement End-to-End checksum.

3. Add read/write a deadline on each read/write.

4. Add a header.

5. Import xlog for logging.

6. Replacing Gob encoding by binary encoding.

7. Remove batch supports

8. Remove public Async API

9. Client/Server reader will wait for a certain time to get header, if timeout it'll retry, avoiding hang.

10. Add needed comments to explain the logic.

11. Remove statistics.

## Limitation

otcp is not designed for big data chunk stream. Each request read/write will wait for the Done chan, then begins to work.
The max size of object in Zai is 4MB, so it's okay to use this model.

## Performance Tuning

The origin has tried its best to make things non-blocking.

It's hard to compare directly, because the features of otcp is limited. Under almost "same" benchmark test,
otcp gets 3x better than gorpc. (Both of them are using same configs, including buffer size, client connections, flush delay)

### Done

1. Binary encoding/decoding

2. Reuse memory (pool for tiny bytes slice)

...

### TODO

Combine requests, improving throughput.

### Given up

#### Conn Deadline

In Go standard library, net.Conn use Time.Until to get duration.

Which means if the Time has no monotonic time it will call time.Now() again,
so it's meaningless to call tsc.UnixNano() outside.

