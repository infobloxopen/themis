package main

import (
	"bytes"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pdpserver/server"
)

func main() {
	log.Print("Creating server...")
	logBuffer := new(bytes.Buffer)
	logger := logrus.New()
	logger.Out = logBuffer
	logger.Level = logrus.ErrorLevel

	s := server.NewServer(
		server.WithLogger(logger),
	)

	log.Print("Loading policy...")
	if err := s.ReadPolicies(strings.NewReader(regexPrefixPolicy)); err != nil {
		log.Fatalf("Failed to load policy: %s", err)
	}

	log.Print("Preparing data...")
	req := &pb.Request{
		Attributes: []*pb.Attribute{
			{
				Id:    "x",
				Type:  "string",
				Value: "prefix-match-test",
			},
		},
	}

	timings := make([]timing, count)
	th := make(chan bool, threshold)

	log.Printf("Validating %d requests...", count)

	var wg sync.WaitGroup
	for i := range timings {
		th <- true
		wg.Add(1)
		go func(i int) {
			defer func() {
				wg.Done()
				<-th
			}()

			start := time.Now()
			res, err := s.Validate(context.Background(), req)
			end := time.Now()

			if err != nil {
				log.Fatalf("Failed to validate request: %s", err)
			}

			timings[i] = timing{
				s: start,
				r: end,
				e: checkResponse(res),
			}
		}(i)
	}

	wg.Wait()

	log.Printf("Dumping timings to %q...", output)
	dump(timings, output)
}
