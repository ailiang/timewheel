package main

import (
	"fmt"
	"time"
	"tw/timewheel"
)

func main() {
	job := func(data interface{}) {
		fmt.Println(data)
	}
	tw := timewheel.NewTimeWheel(time.Second, 10, job)
	tw.Start()
	tw.AddTimer(22*time.Second, "22s")
	tw.AddTimer(1*time.Second, "1s")

	select {}
}
