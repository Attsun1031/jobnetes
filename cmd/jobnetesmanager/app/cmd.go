package app

import (
	"fmt"

	db2 "github.com/Attsun1031/jobnetes/dao/db"
	"github.com/Attsun1031/jobnetes/di"
	"github.com/Attsun1031/jobnetes/manager"
	"github.com/spf13/cobra"
)

func NewJobmanagerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "jobnetesmanager",
		Long: `This is jobnetesmanager for jobnetes`,
		Run:  run,
	}
	return cmd
}

func run(cmd *cobra.Command, args []string) {
	fmt.Println("jobnetesmanager run")
	manager.InitConfig()
	db := db2.Connect(manager.ManagerConfig.DbConfig)
	mgr := &manager.WorkflowManagerMain{
		Db:                                 db,
		WorkflowExecutionDao:               di.InjectWorkflowExecutionDao(),
		WorkflowExecutionProcessorRegistry: di.InjectWorkflowExecutionStateProcessorFactory(),
	}
	mgr.Run()
}