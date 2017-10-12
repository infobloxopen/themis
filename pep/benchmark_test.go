package pep

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/infobloxopen/themis/pdp-service"
)

const (
	oneStageBenchmarkPolicySet = `# Policy set for benchmark
attributes:
  k3: domain
  x: string

policies:
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
    - x: DefaultRule
  - id: First
    effect: Permit
    obligations:
    - x: First
  - id: Second
    effect: Permit
    obligations:
    - x: Second
  - id: Third
    effect: Permit
    obligations:
    - x: Third
  - id: Fourth
    effect: Permit
    obligations:
    - x: Fourth
  - id: Fifth
    effect: Permit
    obligations:
    - x: Fifth
`

	twoStageBenchmarkPolicySet = `# Policy set for benchmark 2-level nesting policy
attributes:
  k2: string
  k3: domain
  x: string

policies:
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
`

	threeStageBenchmarkPolicySet = `# Policy set for benchmark 3-level nesting policy
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

type decisionRequest struct {
	Direction string `pdp:"k1"`
	Policy    string `pdp:"k2"`
	Domain    string `pdp:"k3,domain"`
}

type decisionResponse struct {
	Effect string `pdp:"Effect"`
	Reason string `pdp:"Reason"`
	X      string `pdp:"x"`
}

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

	decisionRequests []decisionRequest
	rawRequests      []pb.Request
)

func init() {
	decisionRequests = make([]decisionRequest, 2000000)
	for i := range decisionRequests {
		decisionRequests[i] = decisionRequest{
			Direction: directionOpts[rand.Intn(len(directionOpts))],
			Policy:    policySetOpts[rand.Intn(len(policySetOpts))],
			Domain:    domainOpts[rand.Intn(len(domainOpts))],
		}
	}

	rawRequests = make([]pb.Request, 2000000)
	for i := range rawRequests {
		rawRequests[i] = pb.Request{
			Attributes: []*pb.Attribute{
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

func benchmarkPolicySet(name, p string, b *testing.B) {
	ok := true
	tmpYAST, tmpJCon, pdp, c := startPDPServer(p, b)
	defer func() {
		c.Close()

		_, errDump, _ := pdp.kill()
		if !ok && len(errDump) > 0 {
			b.Logf("PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(tmpYAST)
		os.Remove(tmpJCon)
	}()

	ok = b.Run(name, func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			in := decisionRequests[n%len(decisionRequests)]

			var out decisionResponse
			c.ModalValidate(in, &out)

			if (out.Effect != "DENY" && out.Effect != "PERMIT" && out.Effect != "NOTAPPLICABLE") ||
				out.Reason != "Ok" {
				b.Fatalf("unexpected response: %#v", out)
			}
		}
	})
}

func BenchmarkOneStagePolicySet(b *testing.B) {
	benchmarkPolicySet("BenchmarkOneStagePolicySet", oneStageBenchmarkPolicySet, b)
}

func BenchmarkTwoStagePolicySet(b *testing.B) {
	benchmarkPolicySet("BenchmarkTwoStagePolicySet", twoStageBenchmarkPolicySet, b)
}

func BenchmarkThreeStagePolicySet(b *testing.B) {
	benchmarkPolicySet("BenchmarkThreeStagePolicySet", threeStageBenchmarkPolicySet, b)
}

func BenchmarkThreeStagePolicySetRaw(b *testing.B) {
	ok := true
	tmpYAST, tmpJCon, pdp, c := startPDPServer(threeStageBenchmarkPolicySet, b)
	defer func() {
		c.Close()

		_, errDump, _ := pdp.kill()
		if !ok && len(errDump) > 0 {
			b.Logf("PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(tmpYAST)
		os.Remove(tmpJCon)
	}()

	ok = b.Run("BenchmarkThreeStagePolicySetRaw", func(b *testing.B) {
		var out pb.Response
		for n := 0; n < b.N; n++ {
			in := rawRequests[n%len(rawRequests)]

			c.ModalValidate(in, &out)

			if (out.Effect != pb.Response_DENY &&
				out.Effect != pb.Response_PERMIT &&
				out.Effect != pb.Response_NOTAPPLICABLE) ||
				out.Reason != "Ok" {
				b.Fatalf("unexpected response: %#v", out)
			}
		}
	})
}

func BenchmarkThreeStagePolicySetStreams(b *testing.B) {
	ok := true
	tmpYAST, tmpJCon, pdp, c := startPDPServer(threeStageBenchmarkPolicySet, b, WithStreams(100))
	defer func() {
		c.Close()

		_, errDump, _ := pdp.kill()
		if !ok && len(errDump) > 0 {
			b.Logf("PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(tmpYAST)
		os.Remove(tmpJCon)
	}()

	ok = b.Run("BenchmarkThreeStagePolicySetStreams", func(b *testing.B) {
		var out pb.Response
		for n := 0; n < b.N; n++ {
			in := rawRequests[n%len(rawRequests)]

			c.StreamValidate(in, &out)

			if (out.Effect != pb.Response_DENY &&
				out.Effect != pb.Response_PERMIT &&
				out.Effect != pb.Response_NOTAPPLICABLE) ||
				out.Reason != "Ok" {
				b.Fatalf("unexpected response: %#v", out)
			}
		}
	})
}

func startPDPServer(p string, b *testing.B, opts ...Option) (string, string, *proc, Client) {
	tmpYAST, err := makeTempFile(p, "policy")
	if err != nil {
		b.Fatalf("can't create policy file: %s", err)
	}

	tmpJCon, err := makeTempFile(benchmarkContent, "content")
	if err != nil {
		os.Remove(tmpYAST)
		b.Fatalf("can't create content file: %s", err)
	}

	pdp, err := newProc("pdpserver", "-p", tmpYAST, "-j", tmpJCon)
	if err != nil {
		os.Remove(tmpYAST)
		os.Remove(tmpJCon)
		b.Fatalf("can't start PDP server: %s", err)
	}

	time.Sleep(time.Second)

	c := NewClient(opts...)
	err = c.Connect("127.0.0.1:5555")
	if err != nil {
		os.Remove(tmpYAST)
		os.Remove(tmpJCon)

		_, errDump, _ := pdp.kill()
		if len(errDump) > 0 {
			b.Fatalf("can't connect to PDP server: %s\nPDP server dump:\n%s", err, strings.Join(errDump, "\n"))
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	return tmpYAST, tmpJCon, pdp, c
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

type proc struct {
	name string

	proc *os.Process

	out *pipe
	err *pipe
}

func newProc(name string, argv ...string) (*proc, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return nil, fmt.Errorf("can't find %q: %s", name, err)
	}

	p := &proc{name: name}

	outPipe, err := newPipe(name, "stdout")
	if err != nil {
		p.cleanup()
		return nil, fmt.Errorf("can't create stdout pipe for %q: %s", name, err)
	}

	p.out = outPipe

	errPipe, err := newPipe(name, "stderr")
	if err != nil {
		p.cleanup()
		return nil, fmt.Errorf("can't create stderr pipe for %q: %s", name, err)
	}

	p.err = errPipe

	attr := &os.ProcAttr{Files: []*os.File{nil, p.out.w, p.err.w}}
	prc, err := os.StartProcess(path, append([]string{path}, argv...), attr)
	if err != nil {
		p.cleanup()
		return nil, fmt.Errorf("can't start process %q: %s", name, err)
	}

	p.proc = prc
	return p, nil
}

func (p *proc) cleanup() ([]string, []string) {
	var (
		out []string
		err []string
	)

	if p.out != nil {
		out = p.out.cleanup()
		p.out = nil
	}

	if p.err != nil {
		err = p.err.cleanup()
		p.err = nil
	}

	return out, err
}

func (p *proc) kill() ([]string, []string, error) {
	if p == nil {
		return nil, nil, nil
	}

	if p.proc != nil {
		err := p.proc.Signal(os.Interrupt)
		if err != nil {
			return p.out.dump(), p.err.dump(), fmt.Errorf("can't send interrupt signal to %q: %s", p.name, err)
		}

		_, err = p.proc.Wait()
		if err != nil {
			return p.out.dump(), p.err.dump(), fmt.Errorf("failed to wait for %q to exit: %s", p.name, err)
		}

		p.proc = nil
	}

	out, err := p.cleanup()
	return out, err, nil
}

type noLogLine struct {
	name   string
	substr string
	step   int
}

func (err *noLogLine) Error() string {
	return fmt.Sprintf("no expected log line for %q (%q) at %d step", err.name, err.substr, err.step)
}

func (p *proc) waitErrLine(s string) error {
	ok, err := p.err.wait(s, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error on waiting for %q logs: %s", p.name, err)
	}

	if !ok {
		return &noLogLine{
			name:   p.name,
			substr: s,
		}
	}

	return nil
}

type pipe struct {
	name string
	kind string

	r *os.File
	w *os.File

	err error

	storage *storage
	lookup  *lookup

	sync.WaitGroup
	sync.Mutex
}

func newPipe(name, kind string) (*pipe, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	p := &pipe{
		name:    name,
		kind:    kind,
		r:       r,
		w:       w,
		storage: newStorage(),
		lookup:  newLookup(),
	}
	p.scan()

	return p, nil
}

func (p *pipe) scan() {
	s := bufio.NewScanner(p.r)
	p.Add(1)

	go func() {
		defer func() {
			p.lookup.done()
			p.Done()
		}()

		for s.Scan() {
			p.storage.put(s.Text())
			p.lookup.iteration(p.storage)
		}

		if err := s.Err(); err != nil {
			p.Lock()
			p.err = err
			p.Unlock()
		}
	}()
}

func (p *pipe) setErr(err error) {
	p.Lock()
	defer p.Unlock()

	p.err = err
}

func (p *pipe) getErr() error {
	p.Lock()
	defer p.Unlock()

	err := p.err
	return err
}

func (p *pipe) wait(s string, timeout time.Duration) (bool, error) {
	if err := p.getErr(); err != nil {
		return false, err
	}

	return p.lookup.wait(s, timeout, p.storage), nil
}

func (p *pipe) dump() []string {
	if p == nil || p.storage == nil {
		return nil
	}

	p.storage.Lock()
	defer p.storage.Unlock()
	return p.storage.lines
}

func (p *pipe) cleanup() []string {
	if p.w != nil {
		p.w.Close()
		p.w = nil
		p.Wait()
	}

	if p.r != nil {
		p.r.Close()
		p.r = nil
	}

	return p.storage.lines
}

type storage struct {
	lines []string
	count int

	sync.Mutex
}

func newStorage() *storage {
	return &storage{lines: []string{}}
}

func (s *storage) put(line string) {
	s.Lock()
	defer s.Unlock()

	s.lines = append(s.lines, line)
}

func (s *storage) scan(f func(string) bool) bool {
	s.Lock()
	defer s.Unlock()

	for i, line := range s.lines[s.count:] {
		if f(line) {
			s.count += i + 1
			return true
		}
	}

	s.count = len(s.lines)
	return false
}

type lookup struct {
	ch     chan int
	found  bool
	substr *string

	sync.Mutex
}

func newLookup() *lookup {
	return &lookup{ch: make(chan int)}
}

func (lu *lookup) iteration(s *storage) {
	lu.Lock()
	defer lu.Unlock()

	if lu.substr == nil {
		return
	}

	s.scan(func(line string) bool {
		if strings.Contains(line, *lu.substr) {
			lu.found = true

			select {
			default:
			case lu.ch <- 0:
			}

			return true
		}

		return false
	})
}

func (lu *lookup) done() {
	close(lu.ch)
}

func (lu *lookup) start(substr string, s *storage) bool {
	lu.Lock()
	defer lu.Unlock()

	lu.found = false

	found := lu.history(substr, s)
	if found {
		return true
	}

	lu.substr = &substr
	return false
}

func (lu *lookup) stop() bool {
	lu.Lock()
	defer lu.Unlock()

	found := lu.found
	lu.substr = nil

	return found
}

func (lu *lookup) history(substr string, s *storage) bool {
	return s.scan(func(line string) bool { return strings.Contains(line, substr) })
}

func (lu *lookup) wait(substr string, timeout time.Duration, s *storage) bool {
	found := lu.start(substr, s)
	if found {
		return true
	}

	select {
	case <-time.After(timeout):
	case <-lu.ch:
	}

	return lu.stop()
}
