package cronschedule_test

import (
	"cronschedule"
	"fmt"
	"testing"
)

func TestCronSchedule(t *testing.T) {

	s := "10-12 1,2,5,6 3/2 * *"
	schedule, err := cronschedule.Parse(s)
	if err != nil {
		t.Errorf("%s", err)
	}

	fmt.Println(schedule.PrettyString())

}
