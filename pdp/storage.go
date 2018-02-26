package pdp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"

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
	if s.tag == nil {
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

// PolicyUpdateDetail stores detailed information for policy updates
type PolicyUpdateDetail map[string][]*Rule

// NewPolicyUpdateDetail maps visible policy ids in the update to its Rules
func NewPolicyUpdateDetail(update *PolicyUpdate) PolicyUpdateDetail {
	updatedPolicies := make(PolicyUpdateDetail)
	for _, cmd := range update.cmds {
		if policy, ok := cmd.entity.(*Policy); ok {
			if pid, ok := policy.GetID(); ok {
				updatedPolicies[pid] = policy.rules
			}
		}
	}
	return updatedPolicies
}

// GetDetail returns a description of the policy update where
// each line shows the policy id and policy's first nShow-1
// visible rule ids and the last visible rule id
func (detail PolicyUpdateDetail) GetDetail(nShow uint) string {
	policies := make([]string, 0, len(detail))
	for pid, rules := range detail {
		var (
			iRule   int
			ruleStr string
			pubIdx  uint
			nRules  = len(rules)
			ruleIDs = make([]string, nShow)
		)

		// find the first nShow-1 visible rules
		for iRule = 0; iRule < nRules && pubIdx < nShow-1; iRule++ {
			if ruleID, ok := rules[iRule].GetID(); ok {
				ruleIDs[pubIdx] = strconv.Quote(ruleID)
				pubIdx++
			}
		}
		// look for the last visible ruleID
		for j := nRules - 1; j > iRule && pubIdx < nShow; j++ {
			if ruleID, ok := rules[j].GetID(); ok {
				ruleIDs[pubIdx] = strconv.Quote(ruleID)
				pubIdx++
			}
		}
		// assert pubIdx <= nShow
		if pubIdx == nShow {
			ruleStr = strings.Join(ruleIDs[:nShow-1], ", ") +
				", ..., " + ruleIDs[nShow-1]
		} else {
			ruleStr = strings.Join(ruleIDs[:pubIdx], ", ")
		}
		policies = append(policies, fmt.Sprintf("policy %q rules(%s)", pid, ruleStr))
	}
	return strings.Join(policies, "\n")
}

// FilterLevel specifies that PolicyUpdateDetail will only reveal
// details when the log level is Debug
func (detail PolicyUpdateDetail) FilterLevel() logrus.Level {
	return logrus.DebugLevel
}

// String returns the less detailed information
func (detail PolicyUpdateDetail) String() string {
	// only show policy ids in case detail mode is not supported
	pids := make([]string, 0, len(detail))
	for pid := range detail {
		pids = append(pids, pid)
	}
	return strings.Join(pids, ", ")
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
