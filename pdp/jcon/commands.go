package jcon

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

func unmarshalCommand(d *json.Decoder, s pdp.Symbols, u *pdp.ContentUpdate) error {
	var op int
	opOk := false

	var path []string
	pathOk := false

	var entity interface{}
	entityType := updateEntityTypeUnknown

	err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		switch strings.ToLower(k) {
		case "op":
			if opOk {
				return newDuplicateCommandFieldError(k)
			}

			s, err := jparser.GetString(d, "operation")
			if err != nil {
				return err
			}

			op, opOk = pdp.UpdateOpIDs[strings.ToLower(s)]
			if !opOk {
				return newUnknownContentUpdateOperationError(s)
			}

			return nil

		case "path":
			if pathOk {
				return newDuplicateCommandFieldError(k)
			}
			path = []string{}
			err := jparser.GetStringSequence(d, func(idx int, s string) error {
				path = append(path, s)
				return nil
			}, "path")
			if err != nil {
				return err
			}

			pathOk = true
			return nil

		case "entity":
			if entityType != updateEntityTypeUnknown {
				return newDuplicateCommandFieldError(k)
			}

			var err error
			entityType, entity, err = unmarshalEntity(d, s)
			if err != nil {
				return err
			}

			return nil
		}

		return newUnknownCommadFieldError(k)
	}, "command")
	if err != nil {
		return err
	}

	if !opOk {
		return newMissingCommandOpError()
	}

	if !pathOk {
		return newMissingCommandPathError()
	}

	if op == pdp.UOAdd || op == pdp.UOAppendShard {
		if entityType == updateEntityTypeUnknown {
			return newMissingCommandEntityError()
		}

		if op == pdp.UOAdd && entityType != updateEntityTypeContentItem ||
			op == pdp.UOAppendShard && entityType != updateEntityTypeShard {
			return fmt.Errorf("Entity type %d doesn't match update operation %d", entityType, op)
		}
	}

	u.Append(op, path, entity)
	return nil
}

func unmarshalCommands(d *json.Decoder, s pdp.Symbols, u *pdp.ContentUpdate) error {
	ok, err := jparser.CheckRootArrayStart(d)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	err = jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
		if err := unmarshalCommand(d, s, u); err != nil {
			return bindErrorf(err, "%d", idx)
		}
		return nil
	}, "update")

	if err != nil {
		return err
	}

	return jparser.CheckEOF(d)
}

const (
	updateEntityTypeUnknown = iota
	updateEntityTypeContentItem
	updateEntityTypeShard
)

var (
	contentItemFields = map[string]struct{}{
		"type":     {},
		"sharding": {},
		"keys":     {},
		"data":     {},
	}

	shardFields = map[string]struct{}{
		"min":     {},
		"max":     {},
		"servers": {},
	}
)

func unmarshalEntity(d *json.Decoder, s pdp.Symbols) (int, interface{}, error) {
	err := jparser.CheckObjectStart(d, "entity")
	if err != nil {
		return updateEntityTypeUnknown, nil, err
	}

	updateEntityType := updateEntityTypeUnknown
	item := &contentItem{s: s}
	shard := new(shardJSON)

	err = jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		switch updateEntityType {
		case updateEntityTypeUnknown:
			field := strings.ToLower(k)

			if _, ok := contentItemFields[field]; ok {
				updateEntityType = updateEntityTypeContentItem
				return item.unmarshal(k, d)
			}

			if _, ok := shardFields[field]; ok {
				updateEntityType = updateEntityTypeShard
				return shard.unmarshal(k, d)
			}

			return fmt.Errorf("field %q doesn't match to content item or shard entity", k)

		case updateEntityTypeContentItem:
			return item.unmarshal(k, d)

		case updateEntityTypeShard:
			return shard.unmarshal(k, d)
		}

		return fmt.Errorf("invalid entity type %d", updateEntityType)
	}, "entity")
	if err != nil {
		return updateEntityTypeUnknown, nil, err
	}

	var entity interface{}
	switch updateEntityType {
	default:
		return updateEntityTypeUnknown, nil, fmt.Errorf("invalid entity type %d", updateEntityType)

	case updateEntityTypeContentItem:
		entity, err = item.get()

	case updateEntityTypeShard:
		entity, err = shard.get()
	}

	return updateEntityType, entity, err
}
