package task

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeScheduler(t *testing.T) {
	scheduler := NewTimeScheduler("test-timer", 100, 60, 2)
	scheduler.Start()
	defer scheduler.Close()

	// Schedule a task to run after 500ms
	scheduler.Delay("task1", 500, func() {
		fmt.Println("Task1 executed at", time.Now())
	})

	// Schedule a task with absolute time
	absTime := time.Now().UnixMilli() + 10000
	scheduler.Add("task2", absTime, func() {
		fmt.Println("Task2 executed at", time.Now())
	})

	// Schedule a DelayTask
	task := NewDelayTask("task3", 1500, func() {
		fmt.Println("Task3 executed at", time.Now())
	})
	scheduler.AddTask(task)

	time.Sleep(20 * time.Second)
}
