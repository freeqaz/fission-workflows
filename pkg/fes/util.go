package fes

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

func validateAggregate(aggregate Aggregate) error {
	if len(aggregate.Id) == 0 {
		return errors.New("aggregate does not contain id")
	}

	if len(aggregate.Type) == 0 {
		return errors.New("aggregate does not contain type")
	}

	return nil
}

type DeepFoldMatcher struct {
	Expected string
}

func (df *DeepFoldMatcher) Match(target string) bool {
	return strings.EqualFold(df.Expected, target)
}

type ContainsMatcher struct {
	Substr string
}

func (cm *ContainsMatcher) Match(target string) bool {
	return strings.Contains(target, cm.Substr)
}

func NewAggregate(entityType string, entityID string) Aggregate {
	return Aggregate{
		Id:   entityID,
		Type: entityType,
	}
}

type EventOpts struct {
	Event
	Data      proto.Message
	Timestamp time.Time
}

func NewEvent(aggregate Aggregate, msg proto.Message) (*Event, error) {
	var data *any.Any
	if msg != nil {
		d, err := ptypes.MarshalAny(msg)
		if err != nil {
			return nil, err
		}
		data = d
	}
	return &Event{
		Aggregate: &aggregate,
		Data:      data,
		Timestamp: ptypes.TimestampNow(),
		Type:      reflect.Indirect(reflect.ValueOf(msg)).Type().Name(),
	}, nil
}
