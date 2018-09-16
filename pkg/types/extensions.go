package types

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// Types other than specified in protobuf
const (
	InputMain    = "default"
	InputBody    = "body"
	InputHeaders = "headers"
	InputQuery   = "query"
	InputMethod  = "method"
	InputParent  = "_parent"

	typedValueShortMaxLen = 32
	WorkflowAPIVersion    = "v1"
)

// InvocationEvent
var invocationFinalStates = []WorkflowInvocationStatus_Status{
	WorkflowInvocationStatus_ABORTED,
	WorkflowInvocationStatus_SUCCEEDED,
	WorkflowInvocationStatus_FAILED,
}

var taskFinalStates = []TaskRunStatus_Status{
	TaskRunStatus_FAILED,
	TaskRunStatus_ABORTED,
	TaskRunStatus_SKIPPED,
	TaskRunStatus_SUCCEEDED,
}

//
// TypedValue
//

// Prints a short description of the Value
func (tv TypedValue) Short() string {
	var val string
	if len(tv.Value) > typedValueShortMaxLen {
		val = fmt.Sprintf("%s[..%d..]", tv.Value[:typedValueShortMaxLen], len(tv.Value)-typedValueShortMaxLen)
	} else {
		val = fmt.Sprintf("%s", tv.Value)
	}

	return fmt.Sprintf("<Type=\"%s\", Val=\"%v\">", tv.Type, strings.Replace(val, "\n", "", -1))
}

func (tv *TypedValue) SetLabel(k string, v string) *TypedValue {
	if tv == nil {
		return tv
	}
	if tv.Labels == nil {
		tv.Labels = map[string]string{}
	}
	tv.Labels[k] = v

	return tv
}

func (tv *TypedValue) GetLabel(k string) (string, bool) {
	if tv == nil {
		return "", false
	}

	if tv.Labels == nil {
		tv.Labels = map[string]string{}
	}
	v, ok := tv.Labels[k]
	return v, ok
}

//
// Error
//

func (m *Error) Error() string {
	return m.Message
}

//
// WorkflowInvocation
//

func (m *WorkflowInvocation) ID() string {
	return m.GetMetadata().GetId()
}

func (m *WorkflowInvocation) TaskRun(id string) (*TaskRun, bool) {
	ti, ok := m.Status.Tasks[id]
	return ti, ok
}

func (m *WorkflowInvocation) TaskRuns() []*TaskRun {
	var tasks []*TaskRun
	for id := range m.Status.Tasks {
		task, _ := m.TaskRun(id)
		tasks = append(tasks, task)
	}
	return tasks
}

//
// WorkflowInvocationStatus
//

func (m *WorkflowInvocationStatus) ToTaskStatus() *TaskRunStatus {
	var statusMapping = map[WorkflowInvocationStatus_Status]TaskRunStatus_Status{
		WorkflowInvocationStatus_UNKNOWN:     TaskRunStatus_UNKNOWN,
		WorkflowInvocationStatus_SCHEDULED:   TaskRunStatus_SCHEDULED,
		WorkflowInvocationStatus_IN_PROGRESS: TaskRunStatus_IN_PROGRESS,
		WorkflowInvocationStatus_SUCCEEDED:   TaskRunStatus_SUCCEEDED,
		WorkflowInvocationStatus_FAILED:      TaskRunStatus_FAILED,
		WorkflowInvocationStatus_ABORTED:     TaskRunStatus_ABORTED,
	}

	return &TaskRunStatus{
		Status:    statusMapping[m.Status],
		Error:     m.Error,
		UpdatedAt: m.UpdatedAt,
		Output:    m.Output,
	}
}

func (m WorkflowInvocationStatus) Finished() bool {
	for _, event := range invocationFinalStates {
		if event == m.GetStatus() {
			return true
		}
	}
	return false
}

func (m WorkflowInvocationStatus) Successful() bool {
	return m.GetStatus() == WorkflowInvocationStatus_SUCCEEDED
}

//
// TaskRun
//

func (m *TaskRun) ID() string {
	return m.GetMetadata().GetId()
}

//
// TaskRunSpec
//

func (m *TaskRunSpec) ToWorkflowSpec() *WorkflowInvocationSpec {
	return &WorkflowInvocationSpec{
		WorkflowId: m.FnRef.ID,
		Inputs:     m.Inputs,
	}
}

//
// TaskRunStatus
//

func (ti TaskRunStatus) Finished() bool {
	for _, event := range taskFinalStates {
		if event == ti.Status {
			return true
		}
	}
	return false
}

func (ti TaskRunStatus) Successful() bool {
	return ti.GetStatus() == TaskRunStatus_SUCCEEDED
}

//
// Task
//

func (m *Task) ID() string {
	return m.GetMetadata().GetId()
}

//
// TaskSpec
//

func (m *TaskSpec) Input(key string, val *TypedValue) *TaskSpec {
	if len(m.Inputs) == 0 {
		m.Inputs = map[string]*TypedValue{}
	}
	m.Inputs[key] = val

	return m
}

func (m *TaskSpec) Parent() (string, bool) {
	var parent string
	var present bool
	for id, params := range m.Requires {
		if params.Type == TaskDependencyParameters_DYNAMIC_OUTPUT {
			present = true
			parent = id
			break
		}
	}
	return parent, present
}

func (m *TaskSpec) Require(taskID string, opts ...*TaskDependencyParameters) *TaskSpec {
	if m.Requires == nil {
		m.Requires = map[string]*TaskDependencyParameters{}
	}
	var params *TaskDependencyParameters
	if len(opts) > 0 {
		params = opts[0]
	}

	m.Requires[taskID] = params
	return m
}

func (m *TaskSpec) Overlay(overlay *TaskSpec) *TaskSpec {
	nt := proto.Clone(m).(*TaskSpec)
	nt.Await = overlay.Await
	nt.Requires = overlay.Requires
	return nt
}

//
// Workflow
//

func (m *Workflow) ID() string {
	return m.GetMetadata().GetId()
}

// Note: this only retrieves the statically, top-level defined tasks
func (m *Workflow) Task(id string) (*Task, bool) {
	var ok bool
	spec, ok := m.Spec.Tasks[id]
	if !ok {
		return nil, false
	}
	var status *TaskStatus
	if m.Status.Tasks != nil {
		status, ok = m.Status.Tasks[id]
	}
	if !ok {
		status = &TaskStatus{
			UpdatedAt: ptypes.TimestampNow(),
		}
	}

	return &Task{
		Metadata: &ObjectMetadata{
			Id:        id,
			CreatedAt: m.Metadata.CreatedAt,
		},
		Spec:   spec,
		Status: status,
	}, true
}

// Note: this only retrieves the statically top-level defined tasks
func (m *Workflow) Tasks() []*Task {
	var tasks []*Task
	for id := range m.Spec.Tasks {
		task, _ := m.Task(id)
		tasks = append(tasks, task)
	}
	return tasks
}

//
// WorkflowSpec
//

func (m *WorkflowSpec) TaskIds() []string {
	var ids []string
	for k := range m.Tasks {
		ids = append(ids, k)
	}
	return ids
}

func (m *WorkflowSpec) SetDescription(s string) *WorkflowSpec {
	m.Description = s
	return m
}

func (m *WorkflowSpec) SetOutput(taskID string) *WorkflowSpec {
	m.OutputTask = taskID
	return m
}

func (m *WorkflowSpec) AddTask(id string, task *TaskSpec) *WorkflowSpec {
	if m.Tasks == nil {
		m.Tasks = map[string]*TaskSpec{}
	}
	m.Tasks[id] = task
	return m
}

//
// WorkflowStatus
//

func (m *WorkflowStatus) Ready() bool {
	return m.Status == WorkflowStatus_READY
}

func (m *WorkflowStatus) Failed() bool {
	return m.Status == WorkflowStatus_FAILED
}

func (m *WorkflowStatus) AddTaskStatus(id string, t *TaskStatus) {
	if m.Tasks == nil {
		m.Tasks = map[string]*TaskStatus{}
	}
	m.Tasks[id] = t
}
