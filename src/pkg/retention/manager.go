// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package retention

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/beego/beego/orm"
	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/retention/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

// Manager defines operations of managing policy
type Manager interface {
	// Create new policy and return ID
	CreatePolicy(p *policy.Metadata) (int64, error)
	// Update the existing policy
	// Full update
	UpdatePolicy(p *policy.Metadata) error
	// Delete the specified policy
	// No actual use so far
	DeletePolicyAndExec(ID int64) error
	// Get the specified policy
	GetPolicy(ID int64) (*policy.Metadata, error)
	// Create a new retention execution
	CreateExecution(execution *Execution) (int64, error)
	// Delete a new retention execution
	DeleteExecution(int64) error
	// Get the specified execution
	GetExecution(eid int64) (*Execution, error)
	// List executions
	ListExecutions(policyID int64, query *q.Query) ([]*Execution, error)
	// GetTotalOfRetentionExecs Count Retention Executions
	GetTotalOfRetentionExecs(policyID int64) (int64, error)
	// List tasks histories
	ListTasks(query ...*q.TaskQuery) ([]*Task, error)
	// GetTotalOfTasks Count Tasks
	GetTotalOfTasks(executionID int64) (int64, error)
	// Create a new retention task
	CreateTask(task *Task) (int64, error)
	// Update the specified task
	UpdateTask(task *Task, cols ...string) error
	// Update the status of the specified task
	// The status is updated only when (the statusRevision > the current revision)
	// or (the the statusRevision = the current revision and status > the current status)
	UpdateTaskStatus(taskID int64, status string, statusRevision int64) error
	// Get the task specified by the task ID
	GetTask(taskID int64) (*Task, error)
	// Get the log of the specified task
	GetTaskLog(taskID int64) ([]byte, error)
}

// DefaultManager ...
type DefaultManager struct {
}

// CreatePolicy Create Policy
func (d *DefaultManager) CreatePolicy(p *policy.Metadata) (int64, error) {
	p1 := &models.RetentionPolicy{}
	p1.ScopeLevel = p.Scope.Level
	p1.ScopeReference = p.Scope.Reference
	p1.TriggerKind = p.Trigger.Kind
	data, _ := json.Marshal(p)
	p1.Data = string(data)
	p1.CreateTime = time.Now()
	p1.UpdateTime = p1.CreateTime
	return dao.CreatePolicy(p1)
}

// UpdatePolicy Update Policy
func (d *DefaultManager) UpdatePolicy(p *policy.Metadata) error {
	p1 := &models.RetentionPolicy{}
	p1.ID = p.ID
	p1.ScopeLevel = p.Scope.Level
	p1.ScopeReference = p.Scope.Reference
	p1.TriggerKind = p.Trigger.Kind
	p.ID = 0
	data, _ := json.Marshal(p)
	p.ID = p1.ID
	p1.Data = string(data)
	p1.UpdateTime = time.Now()
	return dao.UpdatePolicy(p1, "scope_level", "trigger_kind", "data", "update_time")
}

// DeletePolicyAndExec Delete Policy
func (d *DefaultManager) DeletePolicyAndExec(id int64) error {
	return dao.DeletePolicyAndExec(id)
}

// GetPolicy Get Policy
func (d *DefaultManager) GetPolicy(id int64) (*policy.Metadata, error) {
	p1, err := dao.GetPolicy(id)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("no such Retention policy with id %v", id)
		}
		return nil, err
	}
	p := &policy.Metadata{}
	if err = json.Unmarshal([]byte(p1.Data), p); err != nil {
		return nil, err
	}
	p.ID = id
	if p.Trigger.Settings != nil {
		if _, ok := p.Trigger.References[policy.TriggerReferencesJobid]; ok {
			p.Trigger.References[policy.TriggerReferencesJobid] = int64(p.Trigger.References[policy.TriggerReferencesJobid].(float64))
		}
	}
	return p, nil
}

// CreateExecution Create Execution
func (d *DefaultManager) CreateExecution(execution *Execution) (int64, error) {
	exec := &models.RetentionExecution{}
	exec.PolicyID = execution.PolicyID
	exec.StartTime = execution.StartTime
	exec.DryRun = execution.DryRun
	exec.Trigger = execution.Trigger
	return dao.CreateExecution(exec)
}

// DeleteExecution Delete Execution
func (d *DefaultManager) DeleteExecution(eid int64) error {
	return dao.DeleteExecution(eid)
}

// ListExecutions List Executions
func (d *DefaultManager) ListExecutions(policyID int64, query *q.Query) ([]*Execution, error) {
	execs, err := dao.ListExecutions(policyID, query)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var execs1 []*Execution
	for _, e := range execs {
		e1 := &Execution{}
		e1.ID = e.ID
		e1.PolicyID = e.PolicyID
		e1.Status = e.Status
		e1.StartTime = e.StartTime
		e1.EndTime = e.EndTime
		e1.Trigger = e.Trigger
		e1.DryRun = e.DryRun
		execs1 = append(execs1, e1)
	}
	return execs1, nil
}

// GetTotalOfRetentionExecs Count Executions
func (d *DefaultManager) GetTotalOfRetentionExecs(policyID int64) (int64, error) {
	return dao.GetTotalOfRetentionExecs(policyID)
}

// GetExecution Get Execution
func (d *DefaultManager) GetExecution(eid int64) (*Execution, error) {
	e, err := dao.GetExecution(eid)
	if err != nil {
		return nil, err
	}
	e1 := &Execution{}
	e1.ID = e.ID
	e1.PolicyID = e.PolicyID
	e1.Status = e.Status
	e1.StartTime = e.StartTime
	e1.EndTime = e.EndTime
	e1.Trigger = e.Trigger
	e1.DryRun = e.DryRun
	return e1, nil
}

// CreateTask creates task record
func (d *DefaultManager) CreateTask(task *Task) (int64, error) {
	if task == nil {
		return 0, errors.New("nil task")
	}
	t := &models.RetentionTask{
		ExecutionID:    task.ExecutionID,
		Repository:     task.Repository,
		JobID:          task.JobID,
		Status:         task.Status,
		StatusCode:     task.StatusCode,
		StatusRevision: task.StatusRevision,
		StartTime:      task.StartTime,
		EndTime:        task.EndTime,
		Total:          task.Total,
		Retained:       task.Retained,
	}
	return dao.CreateTask(t)
}

// ListTasks lists tasks according to the query
func (d *DefaultManager) ListTasks(query ...*q.TaskQuery) ([]*Task, error) {
	ts, err := dao.ListTask(query...)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	tasks := make([]*Task, 0)
	for _, t := range ts {
		tasks = append(tasks, &Task{
			ID:             t.ID,
			ExecutionID:    t.ExecutionID,
			Repository:     t.Repository,
			JobID:          t.JobID,
			Status:         t.Status,
			StatusCode:     t.StatusCode,
			StatusRevision: t.StatusRevision,
			StartTime:      t.StartTime,
			EndTime:        t.EndTime,
			Total:          t.Total,
			Retained:       t.Retained,
		})
	}
	return tasks, nil
}

// GetTotalOfTasks Count tasks
func (d *DefaultManager) GetTotalOfTasks(executionID int64) (int64, error) {
	return dao.GetTotalOfTasks(executionID)
}

// UpdateTask updates the task
func (d *DefaultManager) UpdateTask(task *Task, cols ...string) error {
	if task == nil {
		return errors.New("nil task")
	}
	if task.ID <= 0 {
		return fmt.Errorf("invalid task ID: %d", task.ID)
	}
	return dao.UpdateTask(&models.RetentionTask{
		ID:             task.ID,
		ExecutionID:    task.ExecutionID,
		Repository:     task.Repository,
		JobID:          task.JobID,
		Status:         task.Status,
		StatusCode:     task.StatusCode,
		StatusRevision: task.StatusRevision,
		StartTime:      task.StartTime,
		EndTime:        task.EndTime,
		Total:          task.Total,
		Retained:       task.Retained,
	}, cols...)
}

// UpdateTaskStatus updates the status of the specified task
func (d *DefaultManager) UpdateTaskStatus(taskID int64, status string, statusRevision int64) error {
	if taskID <= 0 {
		return fmt.Errorf("invalid task ID: %d", taskID)
	}
	st := job.Status(status)
	return dao.UpdateTaskStatus(taskID, status, st.Code(), statusRevision)
}

// GetTask returns the task specified by task ID
func (d *DefaultManager) GetTask(taskID int64) (*Task, error) {
	if taskID <= 0 {
		return nil, fmt.Errorf("invalid task ID: %d", taskID)
	}
	task, err := dao.GetTask(taskID)
	if err != nil {
		return nil, err
	}
	return &Task{
		ID:             task.ID,
		ExecutionID:    task.ExecutionID,
		Repository:     task.Repository,
		JobID:          task.JobID,
		Status:         task.Status,
		StatusCode:     task.StatusCode,
		StatusRevision: task.StatusRevision,
		StartTime:      task.StartTime,
		EndTime:        task.EndTime,
		Total:          task.Total,
		Retained:       task.Retained,
	}, nil
}

// GetTaskLog gets the logs of task
func (d *DefaultManager) GetTaskLog(taskID int64) ([]byte, error) {
	task, err := d.GetTask(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task %d not found", taskID)
	}
	return cjob.GlobalClient.GetJobLog(task.JobID)
}

// NewManager ...
func NewManager() Manager {
	return &DefaultManager{}
}
