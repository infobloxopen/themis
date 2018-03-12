package pdp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// PolicyStorage is a storage for policies.
type PolicyStorage struct {
	tag      *uuid.UUID
	attrs    map[string]Attribute
	policies Evaluable
}

// NewPolicyStorage creates new policy storage with given root policy set
// or policy, symbol table (which maps attribute names to its definitions)
// and tag. Tag can be nil in which case policies can't be updated
// incrementally.
func NewPolicyStorage(p Evaluable, a map[string]Attribute, t *uuid.UUID) *PolicyStorage {
	return &PolicyStorage{tag: t, attrs: a, policies: p}
}

// Root returns root policy from the storage.
func (s *PolicyStorage) Root() Evaluable {
	return s.policies
}

// CheckTag checks if given tag matches to the storage tag. If the storage
// doesn't have any tag, no tag matches the storage and vice versa nil tag
// doesn't match any storage.
func (s *PolicyStorage) CheckTag(tag *uuid.UUID) error {
	if s == nil || s.tag == nil {
		return newUntaggedPolicyModificationError()
	}

	if tag == nil {
		return newMissingPolicyTagError()
	}

	if s.tag.String() != tag.String() {
		return newPolicyTagsNotMatchError(s.tag, tag)
	}

	return nil
}

// NewTransaction creates new transaction for given policy storage.
func (s *PolicyStorage) NewTransaction(tag *uuid.UUID) (*PolicyStorageTransaction, error) {
	err := s.CheckTag(tag)
	if err != nil {
		return nil, err
	}

	return &PolicyStorageTransaction{tag: *tag, attrs: s.attrs, policies: s.policies}, nil
}

// GetPath returns Policy/PolicySet at path if it exists, otherwise return nil, err
func (s *PolicyStorage) GetPath(path []string) (Iterable, error) {
	var (
		err error
		// [PolicySet, Policy, Rule] should always be iterable
		iter = s.policies.(Iterable)
	)
	if rootID, ok := s.policies.GetID(); len(path) == 0 || !ok || rootID != path[0] {
		return nil, fmt.Errorf("Invalid root id or hidden root")
	}
	for _, id := range path[1:] {
		iter, err = iter.FindNext(id)
		if err != nil {
			return nil, err
		}
	}
	return iter, nil
}

// GetSubtree obtains subtree of iter within depth in JSON format
func GetSubtree(iter Iterable, depth uint) string {
	if id, ok := iter.GetID(); ok {
		idInfo := fmt.Sprintf("\"id\":%s", strconv.Quote(id))
		if depth > 0 {
			nChild := iter.NextSize()
			cIdx := 0
			children := make([]string, nChild)
			for i := 0; i < nChild; i++ {
				child := GetSubtree(iter.GetNext(i), depth-1)
				if len(child) > 0 {
					children[cIdx] = child
					cIdx++
				}
			}
			if cIdx > 0 {
				return fmt.Sprintf("{%s,\"elems\":[%s]}", idInfo,
					strings.Join(children[:cIdx], ","))
			}
		} else if iter.NextSize() > 0 {
			return fmt.Sprintf("{%s,\"elems\":\"...\"}", idInfo)
		}
		return fmt.Sprintf("{%s}", idInfo)
	}
	return ""
}

type iterStackInfo struct {
	Iterable
	idx int
}

func (info *iterStackInfo) next() Iterable {
	var out Iterable
	if info.idx < info.NextSize() {
		out = info.GetNext(info.idx)
		info.idx++
	}
	return out
}

// PathQuery df searches for evaluable with id under subtree of root iter
func PathQuery(iter Iterable, id string) ([]string, Iterable, error) {
	rootID, ok := iter.GetID()
	if ok && rootID == id {
		return []string{}, iter, nil
	}
	// depth first search to avoid memory overhead
	// assumes relatively shallow tree
	stack := []iterStackInfo{{iter, 0}}
	for len(stack) > 0 {
		iter = stack[len(stack)-1].next()
		if nil == iter {
			// pop stack
			stack = stack[:len(stack)-1]
		} else {
			// found element matching id
			iterID, ok := iter.GetID()
			if ok && iterID == id {
				nOut := len(stack)
				out := make([]string, nOut)
				for i := 1; i < nOut; i++ {
					out[i-1], _ = stack[i].GetID()
				}
				out[nOut-1] = iterID
				return out, iter, nil
			}
			// push stack if possible
			if ok {
				stack = append(stack, iterStackInfo{iter, 0})
			}
		}
	}
	return nil, nil, fmt.Errorf("Element %s not found", strconv.Quote(id))
}

// Here set of supported update operations is defined.
const (
	// UOAdd stands for add operation (add or append item to a collection).
	UOAdd = iota
	// UODelete is delete operation (remove item from collection).
	UODelete
)

var (
	// UpdateOpIDs maps operation keys to operation ids.
	UpdateOpIDs = map[string]int{
		"add":    UOAdd,
		"delete": UODelete}

	// UpdateOpNames lists operation names in order of operation ids.
	UpdateOpNames = []string{
		"Add",
		"Delete"}
)

// PolicyUpdate encapsulates list of changes to particular policy storage.
type PolicyUpdate struct {
	oldTag uuid.UUID
	newTag uuid.UUID
	cmds   []*command
}

// NewPolicyUpdate creates empty update for policy storage and sets update tags.
// Policy storage must have oldTag so update can be applied. newTag will be set
// to storage after update.
func NewPolicyUpdate(oldTag, newTag uuid.UUID) *PolicyUpdate {
	return &PolicyUpdate{
		oldTag: oldTag,
		newTag: newTag,
		cmds:   []*command{}}
}

// Append inserts particular change to the end of changes list. Op is
// an operation (like add or delete), path identifies policy set, policy or rule
// to perform operation and entity to add (and ignored in case of delete
// operation).
func (u *PolicyUpdate) Append(op int, path []string, entity interface{}) {
	u.cmds = append(u.cmds, &command{op: op, path: path, entity: entity})
}

// String implements Stringer interface.
func (u *PolicyUpdate) String() string {
	if u == nil {
		return "no policy update"
	}

	lines := []string{fmt.Sprintf("policy update: %s - %s", u.oldTag, u.newTag)}
	if len(u.cmds) > 0 {
		lines = append(lines, "commands:")
		for _, cmd := range u.cmds {
			lines = append(lines, "- "+cmd.describe())
		}
	}

	return strings.Join(lines, "\n")
}

type command struct {
	op     int
	path   []string
	entity interface{}
}

func (c *command) describe() string {
	if c == nil {
		return "nop"
	}

	sop := "unknown"
	if c.op >= 0 && c.op < len(UpdateOpNames) {
		sop = UpdateOpNames[c.op]
	}

	qpath := []string{"."}
	if len(c.path) > 0 {
		qpath = make([]string, len(c.path))
		for i, s := range c.path {
			qpath[i] = strconv.Quote(s)
		}
	}

	opPath := strings.Join(qpath, "/")
	if nil != c.entity {
		if evaluable, ok := c.entity.(Evaluable); ok {
			// change to something on path
			return fmt.Sprintf("%s %s to\n  path: (%s)",
				sop, evaluable.describe(), opPath)
		}
	}

	// change directly to path
	return fmt.Sprintf("%s path (%s)", sop, opPath)
}

// PolicyStorageTransaction represents transaction for policy storage.
// Transaction aggregates updates and then can be committed to policy storage
// to make all the updates visible at once.
type PolicyStorageTransaction struct {
	tag      uuid.UUID
	attrs    map[string]Attribute
	policies Evaluable
	err      error
}

// Attributes returns symbol tables captured from policy storage on transaction
// creation.
func (t *PolicyStorageTransaction) Attributes() map[string]Attribute {
	return t.attrs
}

func (t *PolicyStorageTransaction) applyCmd(cmd *command) error {
	switch cmd.op {
	case UOAdd:
		return t.appendItem(cmd.path, cmd.entity)

	case UODelete:
		return t.del(cmd.path)
	}

	return newUnknownPolicyUpdateOperationError(cmd.op)
}

// Apply updates captured policies with given policy update.
func (t *PolicyStorageTransaction) Apply(u *PolicyUpdate) error {
	if t.err != nil {
		return newFailedPolicyTransactionError(t.tag, t.err)
	}

	if t.tag.String() != u.oldTag.String() {
		return newPolicyTransactionTagsNotMatchError(t.tag, u.oldTag)
	}

	for i, cmd := range u.cmds {
		err := t.applyCmd(cmd)
		if err != nil {
			t.err = err
			return bindErrorf(err, "command %d - %s", i, cmd.describe())
		}
	}

	t.tag = u.newTag
	return nil
}

// Commit creates new policy storage with updated policies. Each commit creates
// copy of storage with only its changes applied so applications must ensure
// that all pairs of NewTransaction and Commit for the same content id go
// sequentially.
func (t *PolicyStorageTransaction) Commit() (*PolicyStorage, error) {
	if t.err != nil {
		return nil, newFailedPolicyTransactionError(t.tag, t.err)
	}

	return &PolicyStorage{tag: &t.tag, attrs: t.attrs, policies: t.policies}, nil
}

func (t *PolicyStorageTransaction) appendItem(path []string, v interface{}) error {
	if len(path) <= 0 {
		p, ok := v.(Evaluable)
		if !ok {
			return newInvalidRootPolicyItemTypeError(v)
		}

		if _, ok := p.GetID(); !ok {
			return newHiddenRootPolicyAppendError()
		}

		t.policies = p
		return nil
	}

	ID := path[0]

	if pID, ok := t.policies.GetID(); ok && pID != ID {
		return newInvalidRootPolicyError(ID, pID)
	}

	p, err := t.policies.Append(path[1:], v)
	if err != nil {
		return err
	}

	t.policies = p
	return nil
}

func (t *PolicyStorageTransaction) del(path []string) error {
	if len(path) <= 0 {
		return newEmptyPathModificationError()
	}

	ID := path[0]

	if pID, ok := t.policies.GetID(); ok && pID != ID {
		return newInvalidRootPolicyError(ID, pID)
	}

	if len(path) > 1 {
		p, err := t.policies.Delete(path[1:])
		if err != nil {
			return err
		}

		t.policies = p
		return nil
	}

	t.policies = nil
	return nil
}
