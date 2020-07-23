package cronschedule_test

import (
	"cronschedule"
	"testing"
	"time"
)

func TestNextExecution(t *testing.T) {
	for _, param := range CronTestData {
		schedule, err := cronschedule.Parse(param.Schedule)
		if err != nil {
			t.Errorf("%d|failed to build schedule for %s: %s", param.ID, param.Schedule, err)
			continue
		}

		nextTimes := schedule.NextExecutions(param.T, 5)

		for i := 1; i < 5; i++ {
			expected, err := time.ParseInLocation("2006-01-02 15:04:05", param.ExpectedResults[i], time.Local)
			if err != nil {
				t.Errorf("%d|failed to parse time %s: %s", param.ID, param.ExpectedResults[i], err)
				continue
			}
			if nextTimes[i] != expected {
				t.Errorf("%d|times do not match, expected %v received %v", param.ID, param.ExpectedResults[i], nextTimes[i])
				continue
			}
		}
	}
}

func BenchmarkNextExecution(*testing.B) {
	for _, param := range CronTestData {
		schedule, err := cronschedule.Parse(param.Schedule)
		if err != nil {
			return
		}

		_ = schedule.NextExecution(param.T)

	}
}

func BenchmarkNextExecutions(*testing.B) {
	for _, param := range CronTestData {
		schedule, err := cronschedule.Parse(param.Schedule)
		if err != nil {
			return
		}

		_ = schedule.NextExecutions(param.T, 200)

	}
}
