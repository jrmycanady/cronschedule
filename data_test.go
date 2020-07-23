package cronschedule_test

import "time"

type CronTestEntry struct {
	ID              int
	T               time.Time
	Schedule        string
	ExpectedResults []string
}

var CronTestData = []CronTestEntry{
	{
		ID:       0,
		T:        time.Date(2020, time.July, 23, 15, 28, 0, 0, time.Local),
		Schedule: "0 22 * * 1-5",
		ExpectedResults: []string{
			"2020-07-23 22:00:00",
			"2020-07-24 22:00:00",
			"2020-07-27 22:00:00",
			"2020-07-28 22:00:00",
			"2020-07-29 22:00:00",
		},
	},
	{
		ID:       1,
		T:        time.Date(2020, time.July, 23, 15, 29, 0, 0, time.Local),
		Schedule: "5 0 * 8 *",
		ExpectedResults: []string{
			"2020-08-01 00:05:00",
			"2020-08-02 00:05:00",
			"2020-08-03 00:05:00",
			"2020-08-04 00:05:00",
			"2020-08-05 00:05:00",
		},
	},
	{
		ID:       2,
		T:        time.Date(2020, time.July, 23, 15, 29, 0, 0, time.Local),
		Schedule: "15 14 1 * *",
		ExpectedResults: []string{
			"2020-08-01 14:15:00",
			"2020-09-01 14:15:00",
			"2020-10-01 14:15:00",
			"2020-11-01 14:15:00",
			"2020-12-01 14:15:00",
		},
	},
	{
		ID:       3,
		T:        time.Date(2020, time.July, 23, 15, 30, 0, 0, time.Local),
		Schedule: "23 0-20/2 * * *",
		ExpectedResults: []string{
			"2020-07-23 16:23:00",
			"2020-07-23 18:23:00",
			"2020-07-23 20:23:00",
			"2020-07-24 00:23:00",
			"2020-07-24 02:23:00",
		},
	},
	{
		ID:       4,
		T:        time.Date(2020, time.July, 23, 15, 32, 0, 0, time.Local),
		Schedule: "0 4 8-14 * *",
		ExpectedResults: []string{
			"2020-08-08 04:00:00",
			"2020-08-09 04:00:00",
			"2020-08-10 04:00:00",
			"2020-08-11 04:00:00",
			"2020-08-12 04:00:00",
		},
	},
	{
		ID:       5,
		T:        time.Date(2020, time.July, 23, 15, 32, 0, 0, time.Local),
		Schedule: "23 0-20/2 * * 3,2,4,5",
		ExpectedResults: []string{
			"2020-07-23 16:23:00",
			"2020-07-23 18:23:00",
			"2020-07-23 20:23:00",
			"2020-07-24 00:23:00",
			"2020-07-24 02:23:00",
		},
	},
	{
		ID:       6,
		T:        time.Date(2020, time.July, 23, 15, 34, 0, 0, time.Local),
		Schedule: "* 1 10 * *",
		ExpectedResults: []string{
			"2020-08-10 01:00:00",
			"2020-08-10 01:01:00",
			"2020-08-10 01:02:00",
			"2020-08-10 01:03:00",
			"2020-08-10 01:04:00",
		},
	},
	{
		ID:       7,
		T:        time.Date(2020, time.July, 23, 17, 16, 0, 0, time.Local),
		Schedule: "10/2 2 1-2,30,3 1-5,8 1-5",
		ExpectedResults: []string{
			"2020-08-01 02:10:00",
			"2020-08-01 02:12:00",
			"2020-08-01 02:14:00",
			"2020-08-01 02:16:00",
			"2020-08-01 02:18:00",
		},
	},
}
