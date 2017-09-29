package main

import (
	"fmt"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
)

type memCheckConfig struct {
	limit uint64
	reset float64
	soft  float64
	frag  float64
	back  float64
}

func (s *server) checkMemory(c *memCheckConfig) {
	if c.limit <= 0 {
		return
	}

	now := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	total := float64(m.Sys)
	if total >= c.reset {
		log.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit)}).Fatal("Memory usage is too high. Exiting...")
	}

	if total >= c.soft {
		if s.softMemWarn == nil {
			tmp := now
			s.softMemWarn = &tmp

			log.WithFields(log.Fields{
				"allocated": fmtMemSize(m.Sys),
				"limit":     fmtMemSize(c.limit)}).Warn("Memory usage essentially increased")
		} else if now.Sub(*s.softMemWarn) > time.Minute {
			*s.softMemWarn = now

			log.WithFields(log.Fields{
				"allocated": fmtMemSize(m.Sys),
				"limit":     fmtMemSize(c.limit)}).Warn("Memory usage remains high")
		}
	} else {
		s.softMemWarn = nil
	}

	if float64(m.HeapInuse-m.HeapAlloc)/total >= c.frag {
		if s.fragMemWarn == nil {
			tmp := now
			s.fragMemWarn = &tmp

			log.WithFields(log.Fields{
				"allocated":    fmtMemSize(m.Sys),
				"in-use":       fmtMemSize(m.HeapAlloc),
				"in-use-spans": fmtMemSize(m.HeapInuse)}).Warn("Amount of fragmented memory essentially increased")
		} else if now.Sub(*s.fragMemWarn) > time.Minute {
			*s.fragMemWarn = now

			log.WithFields(log.Fields{
				"allocated":    fmtMemSize(m.Sys),
				"in-use":       fmtMemSize(m.HeapAlloc),
				"in-use-spans": fmtMemSize(m.HeapInuse)}).Warn("Amount of fragmented memory remains high")
		}
	} else {
		s.fragMemWarn = nil
	}

	if (total-float64(m.HeapAlloc))/total >= c.back {
		if s.backMemWarn == nil {
			tmp := now
			s.backMemWarn = &tmp

			log.WithFields(log.Fields{
				"allocated": fmtMemSize(m.Sys),
				"in-use":    fmtMemSize(m.HeapAlloc)}).Warn("Amount of unused memory essentially increased")
		} else if now.Sub(*s.backMemWarn) > time.Minute {
			*s.backMemWarn = now

			log.WithFields(log.Fields{
				"allocated": fmtMemSize(m.Sys),
				"in-use":    fmtMemSize(m.HeapAlloc)}).Warn("Amount of unused memory remains high")
		}
	} else {
		s.backMemWarn = nil
	}
}

func fmtMemSize(size uint64) string {
	if size < 1024 {
		return fmt.Sprintf("%d", size)
	}

	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%d KB", size)
	}

	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%d MB", size)
	}

	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%d GB", size)
	}

	size /= 1024
	return fmt.Sprintf("%d TB", size)
}
