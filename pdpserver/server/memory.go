package server

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	log "github.com/Sirupsen/logrus"
)

// MemLimits structure contains memory limit levels to manage GC
type MemLimits struct {
	limit uint64
	reset float64
	soft  float64
	frag  float64
	back  float64
}

// MakeMemLimits fills MemLimits structure with given parameters
func MakeMemLimits(limit uint64, reset, soft, back, frag float64) (MemLimits, error) {
	m := MemLimits{limit: limit}
	if m.limit > 0 {
		if reset < 0 || reset > 100 {
			return MemLimits{}, fmt.Errorf("reset limit should be in range 0 - 100 but got %f", reset)
		}

		if soft < 0 || soft > 100 {
			return MemLimits{}, fmt.Errorf("soft limit should be in range 0 - 100 but got %f", soft)
		}

		if soft >= reset {
			return MemLimits{},
				fmt.Errorf("reset limit should be higher than soft limit "+
					"but got %f >= %f", soft, reset)
		}

		m.reset = reset / 100 * float64(m.limit)
		m.soft = soft / 100 * float64(m.limit)

		if back < 0 || back > 100 {
			return MemLimits{}, fmt.Errorf("back percentage should be in range 0 - 100 but got %f", back)
		}
		m.back = back / 100

		if frag < 0 || frag > 100 {
			return MemLimits{},
				fmt.Errorf("fragmentation warning percentage should be in range 0 - 100 "+
					"but got %f", frag)
		}
		m.frag = frag / 100
	}

	return m, nil
}

func (s *Server) checkMemory(c *MemLimits) {
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
		s.opts.logger.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Critical memory pressure. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total >= 0.8*limit && s.gcPercent > 10 {
		s.gcPercent = 10
		s.opts.logger.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Hard memory pressure. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total >= 0.7*limit && s.gcPercent > 20 {
		s.gcPercent = 20
		s.opts.logger.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Moderate memory pressure. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total >= 0.6*limit && s.gcPercent > 30 {
		s.gcPercent = 30
		s.opts.logger.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Light memory pressure. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total >= 0.5*limit && s.gcPercent > 50 {
		s.gcPercent = 50
		s.opts.logger.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("Half of memory in use. Decreasing GC target percentage")
		debug.SetGCPercent(s.gcPercent)
	} else if total < 0.3*limit && s.gcPercent != s.gcMax {
		s.gcPercent = s.gcMax
		s.opts.logger.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit),
			"gc%":       s.gcPercent}).Warn("No memory pressure. Returning GC target percentage to maximum")
		debug.SetGCPercent(s.gcPercent)
	}

	if total >= c.reset {
		s.opts.logger.WithFields(log.Fields{
			"allocated": fmtMemSize(m.Sys),
			"limit":     fmtMemSize(c.limit)}).Fatal("Memory usage is too high. Exiting...")
	}

	if total >= c.soft {
		if s.softMemWarn == nil {
			tmp := now
			s.softMemWarn = &tmp

			s.opts.logger.WithFields(log.Fields{
				"allocated": fmtMemSize(m.Sys),
				"limit":     fmtMemSize(c.limit)}).Warn("Memory usage essentially increased")
		} else if now.Sub(*s.softMemWarn) > time.Minute {
			*s.softMemWarn = now

			s.opts.logger.WithFields(log.Fields{
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

			s.opts.logger.WithFields(log.Fields{
				"allocated":    fmtMemSize(m.Sys),
				"in-use":       fmtMemSize(m.HeapAlloc),
				"in-use-spans": fmtMemSize(m.HeapInuse)}).Warn("Amount of fragmented memory essentially increased")
		} else if now.Sub(*s.fragMemWarn) > time.Minute {
			*s.fragMemWarn = now

			s.opts.logger.WithFields(log.Fields{
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

			s.opts.logger.WithFields(log.Fields{
				"allocated": fmtMemSize(m.Sys),
				"in-use":    fmtMemSize(m.HeapAlloc)}).Warn("Amount of unused memory essentially increased")
		} else if now.Sub(*s.backMemWarn) > time.Minute {
			*s.backMemWarn = now

			s.opts.logger.WithFields(log.Fields{
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
