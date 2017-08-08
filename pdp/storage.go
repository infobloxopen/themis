package pdp

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

type PolicyStorage struct {
	tag      *uuid.UUID
	attrs    map[string]Attribute
	policies Evaluable
}

func NewPolicyStorage(p Evaluable, a map[string]Attribute, t *uuid.UUID) *PolicyStorage {
	return &PolicyStorage{tag: t, attrs: a, policies: p}
}

func (s *PolicyStorage) Attributes() map[string]Attribute {
	attrs := make(map[string]Attribute)
	for ID, a := range s.attrs {
		attrs[ID] = a
	}

	return attrs
}

func (s *PolicyStorage) Root() Evaluable {
	return s.policies
}

func (s *PolicyStorage) newTransaction(tag *uuid.UUID) (*policyStorageTransaction, error) {
	if s.tag == nil {
		return nil, newUntaggedPolicyModificationError()
	}

	if tag == nil {
		return nil, newMissingPolicyTagError()
	}

	if !uuid.Equal(*s.tag, *tag) {
		return nil, newPolicyTagsNotMatchError(s.tag, tag)
	}

	return &policyStorageTransaction{attrs: s.attrs, policies: s.policies}, nil
}

const (
	UOAdd = iota
	UODelete
)

var (
	UpdateOpIDs = map[string]int{
		"add":    UOAdd,
		"delete": UODelete}

	UpdateOpNames = []string{
		"Add",
		"Delete"}
)

type PolicyUpdate struct {
	oldTag uuid.UUID
	newTag uuid.UUID
	cmds   []*command
}

func NewPolicyUpdate(oldTag, newTag uuid.UUID) (*PolicyUpdate, error) {
	return &PolicyUpdate{
		oldTag: oldTag,
		newTag: newTag,
		cmds:   []*command{}}, nil
}

func (u *PolicyUpdate) Append(op int, path []string, entity interface{}) {
	u.cmds = append(u.cmds, &command{op: op, path: path, entity: entity})
}

type command struct {
	op     int
	path   []string
	entity interface{}
}

func (c *command) MarshalJSON() ([]byte, error) {
	b := []byte("{")
	var err error

	b, err = appendJSONTag(b, "op")
	if err != nil {
		return nil, err
	}

	b, err = appendJSONString(b, UpdateOpNames[c.op])
	if err != nil {
		return nil, err
	}

	b = append(b, ',')

	b, err = appendJSONTag(b, "path")
	if err != nil {
		return nil, err
	}

	b, err = appendJSONStringArray(b, c.path)
	if err != nil {
		return nil, err
	}

	if c.op == UOAdd && c.entity != nil {
		if entity, ok := c.entity.(*ContentItem); ok {
			b = append(b, ',')

			b, err = appendJSONTag(b, "entity")
			if err != nil {
				return nil, err
			}

			e, err := json.Marshal(entity)
			if err != nil {
				return nil, err
			}

			b = append(b, e...)
		}
	}

	return append(b, '}'), nil
}

type policyStorageTransaction struct {
	attrs    map[string]Attribute
	policies Evaluable
}

func (t *policyStorageTransaction) commit(tag *uuid.UUID) (*PolicyStorage, error) {
	if tag == nil {
		return nil, newMissingPolicyTagError()
	}

	return &PolicyStorage{tag: tag, attrs: t.attrs, policies: t.policies}, nil
}

func (t *policyStorageTransaction) appendItem(path []string, v interface{}) error {
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

func (t *policyStorageTransaction) del(path []string) error {
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
