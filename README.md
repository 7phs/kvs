# (One more in-memory) Key-Value Storages

* Hashed key
* Pre-allocated buffer to store values
* Cleaning dictionary and storages by scheduler
* Different type of storages: map, sync-map, partitioned-map and partitioned-sync-map

## Run

Environment vars:

* **LOG_LEVEL** - level of logging. Supported: DEBUG, INFO, WARNING. ERROR. Default, INFO
* **PORT** - number of port which a server listens. Default, 9889
* **EXPIRATION** - time of key's expiration. Default, 30m
* **MAINTENANCE** - interval of running scheduler to clean dictionary and storages. Default, 10m
* **PREALLOCATED** - size of a pre-allocated buffer to store a values in bytes. Default, 1048576
* **STORAGE_MODE** - mode of a dictionary storages. Supported: map, sync-map, partitioned-map and partitioned-sync-map. Default, partitioned-map

## Benchmark (2 wrk running simultaneously = POST + GET)

* MacBook Pro (15-inch, 2018); 2,6 GHz 6-Core Intel Core i7
* Key: random
* Key length: 3
* Value length: 8 - 3000
* EXPIRATION: 2s
* MAINTENANCE: 10s

### map

POST
```
Running 1m test @ http://localhost:9889/
  2 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     0.87ms  740.57us  38.55ms   95.14%
    Req/Sec     6.18k     1.26k   13.04k    86.33%
  738863 requests in 1.00m, 65.53MB read
Requests/sec:  12291.22
Transfer/sec:      1.09MB
```

GET
``` 
Running 1m test @ http://localhost:9889/
  6 threads and 64 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   560.82us  315.04us  32.04ms   89.64%
    Req/Sec    16.75k     2.40k   40.08k    69.88%
  6003983 requests in 1.00m, 1.56GB read
  Non-2xx or 3xx responses: 5480818
Requests/sec:  99896.13
Transfer/sec:     26.65MB
```

### partitioned-map

POST
```
Running 1m test @ http://localhost:9889/
  2 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   823.81us  437.48us  18.48ms   83.21%
    Req/Sec     6.42k     0.86k   11.15k    78.00%
  767386 requests in 1.00m, 68.06MB read
Requests/sec:  12774.75
Transfer/sec:      1.13MB
```

GET
``` 
Running 1m test @ http://localhost:9889/
  6 threads and 64 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   533.75us  249.92us  11.57ms   84.18%
    Req/Sec    17.50k     2.35k   59.82k    74.28%
  6270569 requests in 1.00m, 1.73GB read
  Non-2xx or 3xx responses: 5655050
Requests/sec: 104335.15
Transfer/sec:     29.46MB
```

### sync-map

POST
```
Running 1m test @ http://localhost:9889/
  2 threads and 16 connections
    Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     0.95ms    1.89ms  55.88ms   98.82%
    Req/Sec     6.41k     1.05k   12.89k    76.83%
  766286 requests in 1.00m, 67.96MB read
Requests/sec:  12761.76
Transfer/sec:      1.13MB
```

GET
``` 
Running 1m test @ http://localhost:9889/
  6 threads and 64 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   525.49us  241.98us  10.50ms   84.45%
    Req/Sec    17.79k     2.44k   57.44k    73.10%
  6378164 requests in 1.00m, 1.69GB read
  Non-2xx or 3xx responses: 5801650
Requests/sec: 106117.55
Transfer/sec:     28.77MB
```

### partitioned-sync-map

POST
```
Running 1m test @ http://localhost:9889/
  2 threads and 16 connections
    Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   845.01us  505.32us  20.99ms   87.45%
    Req/Sec     6.30k     0.96k   10.71k    72.17%
  752204 requests in 1.00m, 66.71MB read
Requests/sec:  12526.13
Transfer/sec:      1.11MB
```

GET
``` 
Running 1m test @ http://localhost:9889/
  6 threads and 64 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   537.30us  273.66us  12.31ms   87.81%
    Req/Sec    17.51k     2.47k   29.06k    70.44%
  6273634 requests in 1.00m, 1.69GB read
  Non-2xx or 3xx responses: 5683807
Requests/sec: 104537.16
Transfer/sec:     28.88MB
```
