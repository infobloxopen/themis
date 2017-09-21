# PEPCLI - Policy Enforcement Point Command Line Interface
PEPCLI implements simple PEP as well as a tool to measure performance of PDP server.

# Decision requests
The utility can make decision requests with help of `test` command. There is a bunch of [examples](../examples) in the repository. Additionally, PEPCLI can reach server on remote machine if its address is specified like below:
```
pepcli -s 192.0.2.1 -i requests.yaml test
```

Requests can be run in loop with `-n` option. If specified number is greater than number of requests in input file, pepcli repeats queries in a loop until desired number is reached. For example command below repeates all queries 3 times if requests.yaml file contains only two of them:
```
pepcli -s 192.0.2.1 -i requests.yaml -n 6 test
```

Specify option `-o` to redirect responses from stdout to a file:
```
pepcli -s 192.0.2.1 -i requests.yaml -n 6 -o responses.yaml test
```

# Performance test
Command `perf` allows to measure PDP server performance. For example to send 10000 requests sequentially and measure timings of requests run:
```
pepcli -i requests.yaml -n 10000 -o test.json perf
```

The commands puts measurement results into test.json file. This file contains just send and receive timestamps and can be analysed and visualised with grinder utility from [DNS Tools](https://github.com/infobloxopen/dnstools/tree/master/mig/analyser):
```
python grinder.py -o test.html -t Test -d sequential test.json
```

Option `-p` makes pepcli run requests in parallel and limits number of parallel requests:
- `-p -1` - runs all requests in parallel;
- `-p 0` - runs all requests sequentially;
- `-p <number>` - runs at most number requests in parallel.

Option `-l` sets limit to request rate by adding pause between requests creation. The pause time is 1s/&lt;rate-limit&gt; so actual limit appears lower.

You can make series of measurements with different parallel numbers and then visualise summary of all runs with grinder utility:
```
pepcli -i requests.yaml -n 100000 -o test-1.json perf -p 1
pepcli -i requests.yaml -n 100000 -o test-2.json perf -p 2
pepcli -i requests.yaml -n 100000 -o test-3.json perf -p 3
pepcli -i requests.yaml -n 100000 -o test-4.json perf -p 4
pepcli -i requests.yaml -n 100000 -o test-5.json perf -p 5
python grinder.py -o test.html -t Test ./
```

The latest command collects all test-&lt;number&gt;.json files and creates report based on all of them.
