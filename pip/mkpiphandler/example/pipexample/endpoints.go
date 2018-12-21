// Package pipexample is a generated PIP server handler package. DO NOT EDIT.
package pipexample

import (
	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/strtree"
	"net"
)

type Endpoints interface {
	Set(int64, domain.Name) (*strtree.Tree, error)
	List(int64, domain.Name) ([]string, error)
	Default(string, net.IP) (*net.IPNet, error)
}
