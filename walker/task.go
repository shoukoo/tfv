package walker

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Task struct {
	Resource     string
	AttibuteKeys map[string][]string
}

func PrepareTask(data []byte) ([]*Task, error) {
	var t map[string]map[string][]string
	var tasks []*Task
	err := yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		return nil, err
	}
	for key, value := range t {
		for k, v := range value {
			var newTask Task
			newTask.Resource = key
			newTask.AttibuteKeys = make(map[string][]string)
			newTask.AttibuteKeys[k] = v
			tasks = append(tasks, &newTask)
			log.Infof("inside PrepareTask  %v \n", newTask)
		}

	}
	return tasks, nil

}
