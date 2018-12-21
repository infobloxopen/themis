package test

import (
	"fmt"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pip/client"
	"github.com/infobloxopen/themis/pip/pipcli/global"
)

func command(conf *global.Config) error {
	n := conf.N
	if n <= 0 {
		n = len(conf.Requests)
	}

	for i := 0; i < n; i++ {
		if i > 0 {
			fmt.Println()
		}
		r := conf.Requests[i%len(conf.Requests)]

		v, err := conf.Client.Get(r.Path, r.Args)
		if err != nil {
			if err == client.ErrNotConnected {
				return err
			}

			fmt.Printf("- err: %q\n", err)
		} else {
			s, err := v.Serialize()
			if err != nil {
				fmt.Printf("- err: %q\n", err)
			} else {
				switch t := v.GetResultType(); t {
				default:
					fmt.Printf("- type: %q\n  content: %q\n", t, s)

				case pdp.TypeSetOfStrings, pdp.TypeSetOfNetworks, pdp.TypeSetOfDomains, pdp.TypeListOfStrings:
					fmt.Printf("- type: %q\n  content: [%s]\n", t, s)
				}
			}
		}
	}

	return nil
}
