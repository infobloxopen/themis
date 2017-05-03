package pdp

import "sort"

func sortSetOfStrings(v map[string]int) []string {
	pairs := newPairList(v)
	sort.Sort(pairs)

	list := make([]string, len(pairs))
	for i, pair := range pairs {
		list[i] = pair.value
	}

	return list
}

type pair struct {
	value string
	order int
}

type pairList []pair

func newPairList(v map[string]int) pairList {
	pairs := make(pairList, len(v))
	i := 0
	for k, v := range v {
		pairs[i] = pair{k, v}
		i++
	}

	return pairs
}

func (p pairList) Len() int {
	return len(p)
}

func (p pairList) Less(i, j int) bool {
	return p[i].order < p[j].order
}

func (p pairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
