package pkg

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	goPkgNetName        = "\"net\""
	goPkgDomainName     = "\"github.com/infobloxopen/go-trees/domain\""
	goPkgStrtreeName    = "\"github.com/infobloxopen/go-trees/strtree\""
	goPkgIPTreeName     = "\"github.com/infobloxopen/go-trees/iptree\""
	goPkgDomainTreeName = "\"github.com/infobloxopen/go-trees/domaintree\""
)

const (
	goPkgNetMask = 1 << iota
	goPkgDomainMask
	goPkgStrtreeMask
	goPkgIPTreeMask
	goPkgDomainTreeMask
)

var (
	goPkgMasks = map[int]string{
		goPkgNetMask:        goPkgNetName,
		goPkgDomainMask:     goPkgDomainName,
		goPkgStrtreeMask:    goPkgStrtreeName,
		goPkgIPTreeMask:     goPkgIPTreeName,
		goPkgDomainTreeMask: goPkgDomainTreeName,
	}

	goPkgMap = map[string]int{
		goTypeNetIP:      goPkgNetMask,
		goTypeNetIPNet:   goPkgNetMask,
		goTypeDomainName: goPkgDomainMask,
		goTypeStrtree:    goPkgStrtreeMask,
		goTypeIPTree:     goPkgIPTreeMask,
		goTypeDomainTree: goPkgDomainTreeMask,
	}
)

func collectImports(types ...string) int {
	out := 0
	for _, t := range types {
		if m, ok := goPkgMap[t]; ok {
			out |= m
		}
	}

	return out
}

func makeImports(pkgs int, imports ...string) []string {
	for m, n := range goPkgMasks {
		if pkgs&m != 0 {
			imports = append(imports, n)
		}
	}

	return imports
}

func fixImports(dir string, files ...string) error {
	cmd := exec.Command("gofmt", append([]string{"-s", "-w"}, files...)...)
	cmd.Dir = dir

	b, err := cmd.CombinedOutput()
	if err != nil {
		s := strings.Split(string(b), "\n")
		return fmt.Errorf("can't adjust format for %q: %s", dir, s[0])
	}

	return nil
}
