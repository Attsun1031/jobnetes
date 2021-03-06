package workflowstate

import (
	"time"

	"fmt"

	"errors"

	"github.com/Attsun1031/jobnetes/dao"
	"github.com/Attsun1031/jobnetes/manager/taskexecutor"
	"github.com/Attsun1031/jobnetes/manager/taskpoller"
	"github.com/Attsun1031/jobnetes/model"
	"github.com/Attsun1031/jobnetes/utils/log"
	"github.com/jinzhu/gorm"
	"k8s.io/client-go/kubernetes"
)

type RunningStateProcessor struct {
	WorkflowExecutionDao dao.WorkflowExecutionDao
	TaskExecutionDao     dao.TaskExecutionDao
	TaskExecutorFactory  taskexecutor.Factory
	TaskPollerFactory    taskpoller.Factory
	KubeClient           kubernetes.Interface
}

func (processor *RunningStateProcessor) ToNextState(execution *model.WorkflowExecution, db *gorm.DB) (bool, error) {
	// check current running task
	succeededTasks, err := checkRunningTask(processor, execution, db)
	if err != nil {
		return false, err
	}

	// start next task
	runNextTask, err := runNextTask(processor, execution, succeededTasks, db)
	if err != nil {
		return false, err
	}
	// if start next task, return false
	if runNextTask {
		return false, nil
	}

	// if all task not completed, return false
	uncompletedTasks, err := processor.TaskExecutionDao.FindUncompletedByWorkflowId(execution.ID, db)
	if err != nil {
		return false, err
	}
	if len(uncompletedTasks) > 0 {
		return false, nil
	}

	// if all task completed, end workflow
	err = endWorkflow(execution, processor, db)
	if err != nil {
		return false, err
	}
	return true, nil
}

func checkRunningTask(processor *RunningStateProcessor, execution *model.WorkflowExecution, db *gorm.DB) ([]*model.TaskExecution, error) {
	// load tasks
	tes, err := processor.TaskExecutionDao.FindUncompletedByWorkflowId(execution.ID, db)
	if err != nil {
		return nil, err
	}
	jobDef := execution.GetJobDef()
	succeededTasks := make([]*model.TaskExecution, 0)
	for _, te := range tes {
		poller, err := processor.TaskPollerFactory.GetTaskPoller(te, jobDef)
		if err != nil {
			log.Logger.Error(err)
			continue
		}
		changed, err := poller.Poll(te, db)
		if err != nil {
			log.Logger.Error(err)
			continue
		}
		if changed {
			if te.IsSucceeded() {
				succeededTasks = append(succeededTasks, te)
			}
		}
	}
	return succeededTasks, nil
}

func runNextTask(processor *RunningStateProcessor, execution *model.WorkflowExecution, succeededTasks []*model.TaskExecution, db *gorm.DB) (bool, error) {
	// start next task
	runNextTask := false
	jobDef := execution.GetJobDef()
	for _, te := range succeededTasks {
		next := jobDef.GetNextTask(te)
		if next == nil {
			log.Logger.Infof("No next task. TaskID=%v", te.ID)
			continue
		}
		runNextTask = true

		log.Logger.Infof("Run next task. CurrentTaskID=%v NextTaskName=%v", te.ID, next.GetName())
		executor, err := processor.TaskExecutorFactory.GetTaskExecutor(next)
		if err != nil {
			return runNextTask, err
		}
		var executed *model.TaskExecution
		executed, err = executor.Execute(execution, db, te.Output, te.ParentTaskExecutionID, te.ID)
		if err != nil {
			return runNextTask, errors.New(fmt.Sprintf("Failed to request task. ExecutionName=%s cause=%s", next.GetName(), err))
		}
		te.NextTaskExecution = executed
		err = processor.TaskExecutionDao.Update(te, db)
		if err != nil {
			return runNextTask, errors.New(fmt.Sprintf("Failed to set NextTaskExecution to successed task. SucceededTaskId=%v NextTaskId=%v cause=%s", te.ID, executed.ID, err))
		}
	}
	return runNextTask, nil
}

func endWorkflow(execution *model.WorkflowExecution, processor *RunningStateProcessor, db *gorm.DB) error {
	completedTasks, err := processor.TaskExecutionDao.FindCompletedByWorkflowId(execution.ID, db)
	if err != nil {
		return err
	}

	hasFailedTask := false

	// failed task exists?
	for _, te := range completedTasks {
		if te.IsFailed() {
			hasFailedTask = true
			break
		}
	}

	endedAt := time.Now()
	execution.EndedAt = &endedAt
	if hasFailedTask {
		execution.Status = model.WfFailed
	} else {
		execution.Status = model.WfSuccess
	}
	return processor.WorkflowExecutionDao.Update(execution, db)
}
