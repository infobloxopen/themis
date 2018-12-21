package pkg

import "errors"

var errNoEndpoints = errors.New("no endpoints provided")

func (s *Schema) getFirstEndpoint() (string, *Endpoint, error) {
	for k, v := range s.Endpoints {
		return k, v, nil
	}

	return "", nil, errNoEndpoints
}

func (p *Endpoint) postProcess() error {
	goArgs, err := makeGoTypeList(p.Args)
	if err != nil {
		return err
	}

	p.goArgs = goArgs
	p.goArgPkgs = collectImports(goArgs...)

	goResult, goResultZero, err := getGoType(p.Result)
	if err != nil {
		return err
	}

	p.goResult = goResult
	p.goResultZero = goResultZero
	p.goResultPkg = collectImports(goResult)

	return nil
}
