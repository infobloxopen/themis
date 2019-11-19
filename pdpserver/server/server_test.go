package server

import (
	"context"
	"strings"
	"testing"

	pb "github.com/infobloxopen/themis/pdp-service"
)

func TestValidate_hook_pre(t *testing.T) {

	var called bool
	hook := func(ctx context.Context) context.Context {
		called = true
		return ctx
	}

	s := NewServer(WithValidatePreHook(hook))

	msg := new(pb.Msg)

	_, err := s.Validate(nil, msg)
	if err != nil {
		t.Fatal(err)
	}

	if e := true; called != e {
		t.Errorf("hook was not called")
	}
}

func TestValidate_hook_post(t *testing.T) {

	var (
		called bool
		outMsg *pb.Msg
	)

	hook := func(ctx context.Context, msg *pb.Msg, err error) {
		called = true
		outMsg = msg
	}

	s := NewServer(WithValidatePostHook(hook))

	msg := new(pb.Msg)

	_, err := s.Validate(nil, msg)
	if err != nil {
		t.Fatal(err)
	}

	if e := true; called != e {
		t.Error("hook was not called")
	}

	if e := newMissingPolicyError().Error(); !strings.Contains(string(outMsg.Body), e) {
		t.Errorf("got: %s wanted: %s", string(outMsg.Body), e)
	}
}

func TestValidate_hook_context(t *testing.T) {

	var (
		preCalled bool
		called    bool
		hasKey    bool
		key       int = 1
	)

	preHook := func(ctx context.Context) context.Context {
		preCalled = true
		return context.WithValue(ctx, key, struct{}{})
	}

	hook := func(ctx context.Context, msg *pb.Msg, err error) {
		called = true
		_, ok := ctx.Value(key).(struct{})
		if ok {
			hasKey = true
		}
	}

	s := NewServer(WithValidatePreHook(preHook), WithValidatePostHook(hook))

	msg := new(pb.Msg)

	_, err := s.Validate(nil, msg)
	if err != nil {
		t.Fatal(err)
	}

	if e := true; hasKey != e {
		t.Error("key was not passed from pre hook to post hook")
	}

	if e := true; preCalled != e {
		t.Error("hook was not called")
	}

	if e := true; called != e {
		t.Error("hook was not called")
	}

}
