# 12-Sharding

```
$ pdpserver -v 3 -l 127.0.0.1:5555 -c 127.0.0.1:5521 -storage 127.0.0.1:5531 -p policy-shard-a.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5565 -c 127.0.0.1:5524 -storage 127.0.0.1:5534 -p policy-shard-b.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5566 -c 127.0.0.1:5525 -storage 127.0.0.1:5535 -p policy-shard-b.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5567 -c 127.0.0.1:5526 -storage 127.0.0.1:5536 -p policy-shard-b.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5575 -c 127.0.0.1:5527 -storage 127.0.0.1:5537 -p policy-shard-c.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5576 -c 127.0.0.1:5528 -storage 127.0.0.1:5538 -p policy-shard-c.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5577 -c 127.0.0.1:5529 -storage 127.0.0.1:5539 -p policy-shard-c.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5555 -c 127.0.0.1:5521 -storage 127.0.0.1:5531 -p content-sharding.yaml -j content-shard-a.json
```

```
$ pdpserver -v 3 -l 127.0.0.1:5565 -content 127.0.0.1:5665 -c 127.0.0.1:5522 -storage 127.0.0.1:5532 -p content-sharding.yaml -j content-shard-b.json
```

```
$ pdpserver -v 3 -l 127.0.0.1:5566 -content 127.0.0.1:5666 -c 127.0.0.1:5523 -storage 127.0.0.1:5533 -p content-sharding.yaml -j content-shard-b.json
```

```
$ pdpserver -v 3 -l 127.0.0.1:5567 -content 127.0.0.1:5667 -c 127.0.0.1:5524 -storage 127.0.0.1:5534 -p content-sharding.yaml -j content-shard-b.json
```

```
$ pdpserver -v 3 -l 127.0.0.1:5575 -content 127.0.0.1:5675 -c 127.0.0.1:5525 -storage 127.0.0.1:5535 -p content-sharding.yaml -j content-shard-c.json
```

```
$ pdpserver -v 3 -l 127.0.0.1:5576 -content 127.0.0.1:5676 -c 127.0.0.1:5526 -storage 127.0.0.1:5536 -p content-sharding.yaml -j content-shard-c.json
```

```
$ pdpserver -v 3 -l 127.0.0.1:5577 -content 127.0.0.1:5677 -c 127.0.0.1:5527 -storage 127.0.0.1:5537 -p content-sharding.yaml -j content-shard-c.json
```
