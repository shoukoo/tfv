package main

import "testing"

//Helper
func errorMsg(t *testing.T, expected string, result interface{}) {
	t.Errorf("Expecting %v but got %v", expected, result)
}

// TestReadConfig
func TestReadConfig(t *testing.T) {

	_, err := readConfig("invalid")

	if err == nil {
		errorMsg(t, "an error", err)
	}

	_, err = readConfig("test/tf.yaml")

	if err != nil {
		errorMsg(t, "no error", err)
	}

}

// TestGetTask
func TestGetTask(t *testing.T) {

	b, err := readConfig("test/tf.yaml")

	if err != nil {
		errorMsg(t, "to find test/tf.yml", err)
	}

	tasks, err := getTasks(b)

	if len(tasks) != 2 {
		errorMsg(t, "2 tasks", len(tasks))
	}

}
