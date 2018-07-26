# 12-Sharding

```
$ pdpserver -v 3 -l 127.0.0.1:5555 -c 127.0.0.1:5521 -storage 127.0.0.1:5531 -p shard-a.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5556 -c 127.0.0.1:5522 -storage 127.0.0.1:5532 -p shard-a.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5557 -c 127.0.0.1:5523 -storage 127.0.0.1:5533 -p shard-a.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5565 -c 127.0.0.1:5524 -storage 127.0.0.1:5534 -p shard-b.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5566 -c 127.0.0.1:5525 -storage 127.0.0.1:5535 -p shard-b.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5567 -c 127.0.0.1:5526 -storage 127.0.0.1:5536 -p shard-b.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5575 -c 127.0.0.1:5527 -storage 127.0.0.1:5537 -p shard-c.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5576 -c 127.0.0.1:5528 -storage 127.0.0.1:5538 -p shard-c.yaml
```

```
$ pdpserver -v 3 -l 127.0.0.1:5577 -c 127.0.0.1:5529 -storage 127.0.0.1:5539 -p shard-c.yaml
```
