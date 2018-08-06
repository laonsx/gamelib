package task

import (
	"fmt"
	"testing"
)

func TestTaskRun(t *testing.T) {

	task := New(32, func(v interface{}) {

		num := v.(int)
		fmt.Println("deal :", 1000+num)

	})
	defer task.Close()

	for i := 0; i < 100; i++ {

		task.Send(i)
		fmt.Println("send :", i)
	}

	task.Wait()
}
