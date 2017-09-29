package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
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
	limit := float64(c.limit)
	if total >= 0.85*limit && s.gcPercent > 5 {
		s.gcPercent = 5
		log.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Critical memory pressure. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total >= 0.8*limit && s.gcPercent > 10 {
		s.gcPercent = 10
		log.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Hard memory pressure. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total >= 0.7*limit && s.gcPercent > 20 {
		s.gcPercent = 20
		log.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Moderate memory pressure. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total >= 0.6*limit && s.gcPercent > 30 {
		s.gcPercent = 30
		log.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Light memory pressure. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total >= 0.5*limit && s.gcPercent > 50 {
		s.gcPercent = 50
		log.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Half of memory in use. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total < 0.3*limit && s.gcPercent != s.gcMax {
		s.gcPercent = s.gcMax
		log.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("No memory pressure. Returning GC target percentage to maximum")
		debug.SetGCPercent(s.gcPercent)
	}

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

	s := float32(size) / 1024
	if s < 1024 {
		return fmt.Sprintf("%.2f KB", s)
	}

	s /= 1024
	if s < 1024 {
		return fmt.Sprintf("%.2f MB", s)
	}

	s /= 1024
	if s < 1024 {
		return fmt.Sprintf("%.2f GB", s)
	}

	s /= 1024
	return fmt.Sprintf("%.2f TB", s)
}
