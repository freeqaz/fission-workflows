package aggregates

import (
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

const (
	TypeTaskRun = "task"
)

type TaskRun struct {
	*fes.BaseEntity
	*types.TaskRun
}

func NewTaskRun(id string, fi *types.TaskRun) *TaskRun {
	tia := &TaskRun{
		TaskRun: fi,
	}

	tia.BaseEntity = fes.NewBaseEntity(tia, *NewTaskRunAggregate(id))

	return tia
}

func NewTaskRunAggregate(id string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   id,
		Type: TypeTaskRun,
	}
}

func (ti *TaskRun) ApplyEvent(event *fes.Event) error {

	eventData, err := fes.UnmarshalEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.TaskStarted:
		ti.TaskRun = &types.TaskRun{
			Metadata: types.NewObjectMetadata(m.GetSpec().TaskId),
			Spec:     m.GetSpec(),
			Status: &types.TaskRunStatus{
				Status:    types.TaskRunStatus_IN_PROGRESS,
				UpdatedAt: event.Timestamp,
			},
		}
	case *events.TaskSucceeded:
		ti.Status.Output = m.GetResult().Output
		ti.Status.Status = types.TaskRunStatus_SUCCEEDED
		ti.Status.UpdatedAt = event.Timestamp
	case *events.TaskFailed:
		// TODO validate event data
		if ti.Status == nil {
			ti.Status = &types.TaskRunStatus{}
		}
		ti.Status.Error = m.GetError()
		ti.Status.UpdatedAt = event.Timestamp
		ti.Status.Status = types.TaskRunStatus_FAILED
	case *events.TaskSkipped:
		ti.Status.Status = types.TaskRunStatus_SKIPPED
		ti.Status.UpdatedAt = event.Timestamp
	default:
		log.WithFields(log.Fields{
			"aggregate": ti.Aggregate(),
		}).Warnf("Skipping unimplemented event: %T", eventData)
	}
	return nil
}

func (ti *TaskRun) GenericCopy() fes.Entity {
	n := &TaskRun{
		TaskRun: ti.Copy(),
	}
	n.BaseEntity = ti.CopyBaseEntity(n)
	return n
}

func (ti *TaskRun) Copy() *types.TaskRun {
	return proto.Clone(ti.TaskRun).(*types.TaskRun)
}
