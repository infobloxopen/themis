package main

import (
	"flag"
	"strings"
	"time"
)

type Config struct {
	Policy    string
	Content   string
	Addresses StringSet
	Timeout   time.Duration
	ChunkSize int
	ContentID string
	FromTag   string
	ToTag     string
}

type StringSet []string

func (s *StringSet) String() string {
	return strings.Join(*s, ", ")
}

func (s *StringSet) Set(v string) error {
	*s = append(*s, v)
	return nil
}

var config Config

func init() {
	flag.StringVar(&config.Policy, "p", "", "policy file to upload")
	flag.StringVar(&config.Content, "j", "", "JSON content to upload")
	flag.Var(&config.Addresses, "s", "server(s) to upload policy to")
	flag.DurationVar(&config.Timeout, "t", 5*time.Second, "connection timeout")
	flag.IntVar(&config.ChunkSize, "c", 64*1024, "size of chunk for splitting uploads")
	flag.StringVar(&config.ContentID, "id", "", "id of content to upload")
	flag.StringVar(&config.FromTag, "vf", "", "tag to update from (if not specified data to upload is full snapshot)")
	flag.StringVar(&config.ToTag, "vt", "", "new tag to set (if not specified data to upload is not updateable)")

	flag.Parse()
}
