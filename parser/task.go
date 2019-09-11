package parser

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Task struct {
	Resource  string
	Attribute string
	Body      interface{}
}

// PrepareTask creates new task struct
func PrepareTask(data []byte) ([]*Task, error) {
	var t map[string]map[string]interface{}
	var tasks []*Task
	err := yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		return nil, err
	}

	for resource, value := range t {
		err = validateConfigValue(value)
		fmt.Printf("bear attack %v\n", err)
		if err != nil {
			break
		}
		for k, v := range value {
			var newTask Task
			newTask.Resource = resource
			newTask.Attribute = k
			newTask.Body = v
			tasks = append(tasks, &newTask)
			log.Infof("inside PrepareTask  %v \n", newTask)
		}
	}
	return tasks, err
}

// validateConfigValue tfv config only supports a slice of string and map type
func validateConfigValue(value interface{}) error {
	var err error
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Slice:
		log.Infof(">> checking config: found slice %+v\n", v.Slice(0, v.Cap()))
		for i := 0; i < v.Len(); i++ {
			err = validateConfigValue(v.Index(i).Interface())
		}
	case reflect.Map:
		log.Infof(">> checking config: found map %+v\n", v.MapKeys())
		for _, key := range v.MapKeys() {
			err = validateConfigValue(v.MapIndex(key).Interface())
		}
	case reflect.String:
		log.Infof(">> checking config: found string \n")
	default:
		err = fmt.Errorf("tfs doesn't handle %s type in the configuration file", v.Kind())
	}
	return err
}

// GenerateTasks creates tasks based on the config
func GenerateTasks(b []byte) ([]*Task, error) {

	tasks, err := PrepareTask(b)
	if err != nil {
		log.Errorf("error preparing task %v", err)
		return nil, err
	}

	return tasks, nil
}
