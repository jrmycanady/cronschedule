package cronschedule_test

import (
	"cronschedule"
	"fmt"
	"testing"
	"time"
)

func TestCronSchedule(t *testing.T) {

	s := "10-12 1,2,5,6 3/2 * *"
	schedule, err := cronschedule.Parse(s)
	if err != nil {
		t.Errorf("%s", err)
	}

	fmt.Println(schedule.PrettyString())

}

func TestNextExecutionTimes(t *testing.T) {
	scheduleStr := "1,2,3,4,5 1 23 1 1"
	schedule, err := cronschedule.Parse(scheduleStr)
	if err != nil {
		t.Fatalf("%s", err)
	}

	execTimes := schedule.NextExecutions(time.Now(), 20)

	for _, t := range execTimes {
		fmt.Println(t)
	}
}

func TestNextExecutionTimesV3(t *testing.T) {
	scheduleStr := "1,2,3,4,5 1 23 1 1"
	schedule, err := cronschedule.Parse(scheduleStr)
	if err != nil {
		t.Fatalf("%s", err)
	}

	execTimes := schedule.NextExecutionsV3(time.Now(), 20)

	for _, t := range execTimes {
		fmt.Println(t)
	}
}

func TestV1vsV3(t *testing.T) {
	scheduleStr := "0 1 23 1 1"
	schedule, err := cronschedule.Parse(scheduleStr)
	if err != nil {
		t.Fatalf("%s", err)
	}

	execTimes := schedule.NextExecutionsV3(time.Now(), 5)

	for _, t := range execTimes {
		fmt.Println(t)
	}

	execTimes = schedule.NextExecutions(time.Now(), 5)

	for _, t := range execTimes {
		fmt.Println(t)
	}
}

func TestNextExecutionV3Times(t *testing.T) {
	scheduleStr := "0 1 23 1 1"
	schedule, err := cronschedule.Parse(scheduleStr)
	if err != nil {
		t.Fatalf("%s", err)
	}

	execTimes := schedule.NextExecutionV3(time.Now())
	fmt.Println(execTimes)

}
