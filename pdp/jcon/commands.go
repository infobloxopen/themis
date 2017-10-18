package jcon

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

func unmarshalCommand(d *json.Decoder, u *pdp.ContentUpdate) error {
	var op int
	opOk := false

	var path []string
	pathOk := false

	var entity *pdp.ContentItem
	entityOk := false

	err := UnmarshalObject(d, func(k string, d *json.Decoder) error {
		switch strings.ToLower(k) {
		case "op":
			if opOk {
				return newDuplicateCommandFieldError(k)
			}

			s, err := GetString(d, "operation")
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
			err := GetStringSequence(d, "path", func(s string) error {
				path = append(path, s)
				return nil
			})
			if err != nil {
				return err
			}

			pathOk = true
			return nil

		case "entity":
			if entityOk {
				return newDuplicateCommandFieldError(k)
			}

			var err error
			entity, err = unmarshalContentItem("", d)
			if err != nil {
				return err
			}

			entityOk = true
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

	if op == pdp.UOAdd && !entityOk {
		return newMissingCommandEntityError()
	}

	u.Append(op, path, entity)
	return nil
}

func unmarshalCommands(d *json.Decoder, u *pdp.ContentUpdate) error {
	ok, err := checkRootArrayStart(d)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	err = UnmarshalObjectArray(d, func(d *json.Decoder) error {
		return unmarshalCommand(d, u)
	}, "update")

	if err != nil {
		return err
	}

	return CheckEOF(d)
}
