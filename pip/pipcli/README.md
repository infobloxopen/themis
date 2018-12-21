# PIPCLI (Policy Information Point Command Line Interface)

PIPCLI is a command line interface client for unified PIP servers. The client can send information requests to a PIP server (or servers) and print responses or measure performance of given servers.

## Usage

Usage of pipcli:
```
$ pipcli [GLOBAL OPTIONS] command [OPTIONS]
```
Global options:
- **-i** - path to input YAML file with requests (default "requests.yaml");
- **-n** - number of requests to sent (default - all requests from input file);
- **-a** - destination address (default "localhost:5600" for "tcp\*" network and "/var/run/pip.socket" for "unix");
- **-net** - kind of destination network (default "tcp", accepts "tcp", "tcp4", "tcp6" or "unix");
- **-round-robin** - use round-robin balancer (works only for "tcp\*" networks);
- **-hot-spot** - use hot-spot balancer (works only for "tcp\*" networks);
- **-s** - list of servers used to initialize balancer (for example -s 192.0.2.1:5600 -s 192.0.2.2:5600 -s 192.0.2.3:5600);
- **-dns** - use discovery by default DNS resolver (works only for "tcp\*" networks);
- **-k8s** - use discovery by in-cluster kubernetes tools (works only for "tcp\*" networks within kubernetes cluster);
- **-max-request-size** - limits request size in bytes (default 10240);
- **-max-queue** - number of requests client can send in parallel (default 100);
- **-buffer-size** - size of input and output buffers (default 1048576);
- **-conn-timeout** - connection timeout (default 30s);
- **-write-interval** - duration after which data from write buffer are sent to network even if write buffer isn't full (default 50µs);
- **-resp-timeout** - response timeout (default 1s);
- **-check-interval** - inteval of response timeout checks (default 50µs).

Commands:
- **test** - sends information requests to PIP and print resposnes (the command has no options);
- **perf** - measures request roundtrip timings.

Options of **perf** command:
- **w** - number of workers to make requests in parallel (default 100);

## Testing PIP server

Run PIP server in one terminal. For example it can be PIPJCon with content from examples/07-selector/content.json:
```
$ pipjcon -j content.json
INFO[0000] PIP JCon server
INFO[0000] opening content                               content=content.json
INFO[0000] parsing content                               content=content.json
INFO[0000] opening service port                          address="localhost:5600" network=tcp
```

And PIPCLI with **test** command in other terminal:
```
$ pipcli test
- type: "Set of Networks"
  content: ["192.0.2.16/28","192.0.2.32/28"]

- type: "Set of Networks"
  content: ["2001:db8:3000::/40","2001:db8:4000::/40"]
```

By default PIPCLI takes requests from requests.yaml which for given example is following:
```yaml
- path: "content/domain-addresses"
  args:
  - good
  - type: domain
    content: example.com

- path: "content/domain-addresses"
  args:
  - bad
  - type: domain
    content: test.com
```

## Measuring Performance

With **perf** command PIPCLI sends informational requests and records timestamps before each request has been sent and after corresponding response has been received. It can try to send several requests in parallel depending on number of workers. Example below shows measurements for 5 requests with 3 workers:
```
$ pipcli -n 5 perf -w 3
{
  "sends": [
    1542295971206358000,
    1542295971206362000,
    1542295971206372000,
    1542295971209660000,
    1542295971209831000
  ],
  "receives": [
    1542295971209660000,
    1542295971209831000,
    1542295971210040000,
    1542295971210685000,
    1542295971210688000
  ],
  "pairs": [
    [
      1542295971206358000,
      1542295971210688000,
      4330000
    ],
    [
      1542295971206362000,
      1542295971210685000,
      4323000
    ],
    [
      1542295971206372000,
      1542295971209660000,
      3288000
    ],
    [
      1542295971209660000,
      1542295971209831000,
      171000
    ],
    [
      1542295971209831000,
      1542295971210040000,
      209000
    ]
  ]
}
```

The returned JSON object contains sorted list of sending timestamps which allows to estimated number of send requests per second. List of receive timestamps ("receives") is sorted list of responses arrival times. Generally, "sends" and "receives" lists can contain timestamps for different requests at the same position. So "pairs" contain send and receive times grouped by request (third number here is roundtrip time or difference between receive and send timestamps). If PIP returns error or undefined value for some request its receive time isn't recorded and "pairs" list contains only send timestamp for such request.
