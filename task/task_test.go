package task

import (
	"fmt"
	"testing"
)

func TestTask_Run(t *testing.T) {

	task := New(4, func(v interface{}) {

		num := v.(int)
		fmt.Println("deal :", 1000000+num)

	})
	defer task.Close()

	for i := 0; i < 100; i++ {

		if i%2 == 0 {

			err := task.SendMsg(i)
			if err != nil {

				fmt.Println("task sendMsg err", err)
			}

			fmt.Println("send :", i)
		} else {

			err := task.SendFn(func() {

				fmt.Println("100fn")
			})

			if err != nil {

				fmt.Println("task sendFn err", err)
			}

			fmt.Println("sendfn :", i)
		}

	}

	task.Wait()
}
