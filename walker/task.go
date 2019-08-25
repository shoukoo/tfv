package walker

/**
walker package is not completed!!

The goal is walk the Terraform files and verify the value of each attribue.
**/

type Task struct {
	Resource     string
	AttibuteKeys map[string][]string
	Errors       []string
	Error        error
	Scores       map[string]bool
}

func NewTask(res string, attibutes map[string][]string) *Task {
	return &Task{
		Resource:     res,
		AttibuteKeys: attibutes,
	}
}
