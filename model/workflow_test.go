package model

import (
	"reflect"
	"testing"
)

func TestWorkflow_GetJobDef(t *testing.T) {
	workflow := &Workflow{
		Definition: `{"name":"test1","tasks":[{"type":"kube-job","name":"task-1","next":"task-2"}]}`,
	}
	jobDef := workflow.GetJobDef()
	expected := &JobDef{
		Name: "test1",
		Tasks: []Task{
			&KubeJobTask{Name: "task-1", NextTaskName: "task-2"},
		},
	}
	if !reflect.DeepEqual(expected, jobDef) {
		t.FailNow()
	}
}

func TestJobDef_GetStartTask(t *testing.T) {
	startTask := &KubeJobTask{Name: "task-1", NextTaskName: "task-2"}
	jobDef := &JobDef{
		Name: "test1",
		Tasks: []Task{
			startTask,
			&KubeJobTask{Name: "task-2", NextTaskName: "task-3"},
		},
	}
	if !reflect.DeepEqual(startTask, jobDef.GetStartTask()) {
		t.FailNow()
	}
}