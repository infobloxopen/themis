package pep

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	pdp "github.com/infobloxopen/themis/pdp-service"
)

const (
	policySet = `# Policy set for benchmark
attributes:
  k1: string
  k2: string
  k3: domain
  x: string

policies:
  alg: FirstApplicableEffect
  policies:
  - target:
    - equal:
      - attr: k1
      - val:
          type: string
          content: "Left"
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/first"
          path:
          - attr: k2
          type: string
      default: DefaultPolicy

    policies:
    - id: DefaultPolicy
      alg: FirstApplicableEffect
      rules:
      - effect: Deny
        obligations:
        - x: DefaultPolicy

    - id: P1
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P1.DefaultRule
      - id: First
        effect: Permit
        obligations:
        - x: P1.First
      - id: Second
        effect: Permit
        obligations:
        - x: P1.Second

    - id: P2
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P2.DefaultRule
      - id: Second
        effect: Permit
        obligations:
        - x: P2.Second
      - id: Third
        effect: Permit
        obligations:
        - x: P2.Third

    - id: P3
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P3.DefaultRule
      - id: Third
        effect: Permit
        obligations:
        - x: P3.Third
      - id: Fourth
        effect: Permit
        obligations:
        - x: P3.Fourth

    - id: P4
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P4.DefaultRule
      - id: Fourth
        effect: Permit
        obligations:
        - x: P4.Fourth
      - id: Fifth
        effect: Permit
        obligations:
        - x: P4.Fifth

    - id: P5
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P5.DefaultRule
      - id: Fifth
        effect: Permit
        obligations:
        - x: P5.Fifth
      - id: First
        effect: Permit
        obligations:
        - x: P5.First

  - target:
    - equal:
      - attr: k1
      - val:
          type: string
          content: "Right"
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/first"
          path:
          - attr: k2
          type: string
      default: DefaultPolicy

    policies:
    - id: DefaultPolicy
      alg: FirstApplicableEffect
      rules:
      - effect: Deny
        obligations:
        - x: DefaultPolicy

    - id: P1
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P1.DefaultRule
      - id: First
        effect: Permit
        obligations:
        - x: P1.First
      - id: Second
        effect: Permit
        obligations:
        - x: P1.Second

    - id: P2
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P2.DefaultRule
      - id: Second
        effect: Permit
        obligations:
        - x: P2.Second
      - id: Third
        effect: Permit
        obligations:
        - x: P2.Third

    - id: P3
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P3.DefaultRule
      - id: Third
        effect: Permit
        obligations:
        - x: P3.Third
      - id: Fourth
        effect: Permit
        obligations:
        - x: P3.Fourth

    - id: P4
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P4.DefaultRule
      - id: Fourth
        effect: Permit
        obligations:
        - x: P4.Fourth
      - id: Fifth
        effect: Permit
        obligations:
        - x: P4.Fifth

    - id: P5
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x: P5.DefaultRule
      - id: Fifth
        effect: Permit
        obligations:
        - x: P5.Fifth
      - id: First
        effect: Permit
        obligations:
        - x: P5.First

  - alg: FirstApplicableEffect
    rules:
    - effect: Deny
      obligations:
      - x: Root Deny
`

	benchmarkContent = `{
    "id": "content",
    "items": {
        "first": {
            "keys": ["string"],
            "type": "string",
            "data": {
                "First": "P1",
                "Second": "P2",
                "Third": "P3",
                "Fourth": "P4",
                "Fifth": "P5",
                "Sixth": "P6",
                "Seventh": "P7"
            }
        },
        "second": {
            "keys": ["domain"],
            "type": "list of strings",
            "data": {
                "first.example.com": ["First", "Third"],
                "second.example.com": ["Second", "Fourth"],
                "third.example.com": ["Third", "Fifth"],
                "first.test.com": ["Fourth", "Sixth"],
                "second.test.com": ["Fifth", "Seventh"],
                "third.test.com": ["Sixth", "First"],
                "first.example.com": ["Seventh", "Second"],
                "second.example.com": ["Firth", "Fourth"],
                "third.example.com": ["Second", "Fifth"],
                "first.test.com": ["Third", "Sixth"],
                "second.test.com": ["Fourth", "Seventh"],
                "third.test.com": ["Fifth", "First"]
            }
        }
    }
}`
)

var (
	directionOpts = []string{
		"Left",
		"Right",
	}

	policySetOpts = []string{
		"First",
		"Second",
		"Third",
		"Fourth",
		"Fifth",
		"Sixth",
		"Seventh",
	}

	domainOpts = []string{
		"first.example.com",
		"second.example.com",
		"third.example.com",
		"first.test.com",
		"second.test.com",
		"third.test.com",
		"first.example.com",
		"second.example.com",
		"third.example.com",
		"first.test.com",
		"second.test.com",
		"third.test.com",
	}

	requests []*pdp.Request
)

func init() {
	requests = make([]*pdp.Request, 2000000)
	for i := range requests {
		requests[i] = &pdp.Request{
			Attributes: []*pdp.Attribute{
				{
					Id:    "k1",
					Type:  "string",
					Value: directionOpts[rand.Intn(len(directionOpts))],
				},
				{
					Id:    "k2",
					Type:  "string",
					Value: policySetOpts[rand.Intn(len(policySetOpts))],
				},
				{
					Id:    "k3",
					Type:  "domain",
					Value: domainOpts[rand.Intn(len(domainOpts))],
				},
			},
		}
	}

}

func BenchmarkNoBatch(b *testing.B) {
	benchmark(1, 0, 0, b)
}

func BenchmarkBatch(b *testing.B) {
	benchmark(100, 100, 16, b)
}

func benchmark(threadCount int, batchInterval, batchLimit uint, b *testing.B) {
	b.Logf("Threads: %d, Interval: %d, Limit: %d", threadCount, batchInterval, batchLimit)
	ok := true
	tmpYAST, tmpJCon, server, err := startServer(policySet)
	if err != nil {
		b.Fatalf("startServer() failed: %s", err)
	}

	client, err := startClient(batchInterval, batchLimit)
	if err != nil {
		b.Fatalf("startClient() failed: %s", err)
	}

	defer func() {
		if !ok {
			b.Logf("BenchmarkPolicySet failed")
		}
		client.Close()
		if err := killProcess(server); err != nil {
			b.Logf("cannot stop PDP server: %s", err)
		}
		os.Remove(tmpYAST)
		os.Remove(tmpJCon)
	}()

	ok = b.Run("BenchmarkPolicySet", func(b *testing.B) {
		var wg sync.WaitGroup
		for i := 0; i < threadCount; i++ {
			wg.Add(1)
			go func() {
				count := b.N / threadCount
				for n := 0; n < count; n++ {
					in := requests[n%len(requests)]

					out, err := client.Validate(in)
					if err != nil {
						b.Fatalf("unexpected error: %#v", err)
					}

					if (out.Effect != pdp.DENY &&
						out.Effect != pdp.PERMIT &&
						out.Effect != pdp.NOTAPPLICABLE) ||
						out.Reason != "Ok" {
						b.Fatalf("unexpected response: %#v", out)
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func startServer(p string) (string, string, *os.Process, error) {
	tmpYAST, err := makeTempFile(p, "policy")
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot create policy file: %s", err)
	}

	tmpJCon, err := makeTempFile(benchmarkContent, "content")
	if err != nil {
		os.Remove(tmpYAST)
		return "", "", nil, fmt.Errorf("cannot create content file: %s", err)
	}

	process, err := startProcess("../build/pdpserver", "-p", tmpYAST, "-j", tmpJCon)
	if err != nil {
		os.Remove(tmpYAST)
		os.Remove(tmpJCon)
		return "", "", nil, fmt.Errorf("cannot start PDP server: %s", err)
	}

	time.Sleep(time.Second / 10)

	return tmpYAST, tmpJCon, process, nil
}

func startClient(batchInterval, batchLimit uint) (Client, error) {
	client := NewBalancedClient([]string{"127.0.0.1:5555"}, batchInterval, batchLimit)
	err := client.Connect()
	time.Sleep(time.Second / 10)
	return client, err
}

func makeTempFile(s, prefix string) (string, error) {
	f, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", err
	}

	name := f.Name()

	if _, err := f.Write([]byte(s)); err != nil {
		f.Close()
		os.Remove(name)
		return name, err
	}

	if err := f.Close(); err != nil {
		os.Remove(name)
		return name, err
	}

	return name, nil
}

func startProcess(name string, argv ...string) (*os.Process, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return nil, fmt.Errorf("cannot find '%q': %s", name, err)
	}
	attr := &os.ProcAttr{Files: []*os.File{nil, nil, nil}}
	proc, err := os.StartProcess(path, append([]string{path}, argv...), attr)
	if err != nil {
		return nil, fmt.Errorf("cannot start process '%s': %s", path, err)
	}
	return proc, nil
}

func killProcess(proc *os.Process) error {
	if proc == nil {
		return nil
	}

	err := proc.Signal(os.Interrupt)
	if err != nil {
		return fmt.Errorf("cannot send interrupt signal: %s", err)
	}

	_, err = proc.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait to exit: %s", err)
	}

	proc = nil

	return nil
}
