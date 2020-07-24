# cronschedule

cronschedule is a go module that provides the ability to parse an execution schedule specified in cron format and then generate
the next execution times based on any start time. It's specifically intended to be used as a component of job
scheduling and execution implementations.

## Usage

```go

// Parsing the schedule.
cronSchedule = "* * * * *"
schedule, err := cronschedule.Parse(cronSchedule)
if err != nil {
    // Handle errors. Generally thrown for formatting or invalid values.
    panic(err)
}

// Find the next 5 execution times based on the schedule after right now.
nextExecutions := schedule.NextExecutions(time.Now(), 5)
fmt.Println(nextExecutions)

// Find the next single time using the convenience method.
next := schedule.NextExecution(time.Now())
fmt.Println(next)

// Checking if an execution should run at specific time.
t := time.Date(2020,time.December,, 1, 1, 1, 1, 0, time.Locale)
if schedule.ShouldExecute(t) {
    fmt.Println("should execute at %v!", t)
}

// Checking if an execution should run right now.
if schedule.ShouldExecuteNow(time.Now()) {
    fmt.Println("should execute!")
}
```


## Support

* Only supports scheduling including all 5 fields separated by a single space.
* Per UNIX spec, utilizes an OR when both day_of_week and day_of_month are specified as anything but *.
* Text version of days, e.g. SUN-SAT, are _not_ currently supported.
* Text versions of months, e.g. JAN-DEC, are _not_ currently supported.
* Predefined schedules, e.g @yearly are _not_ supported.
* Years are not supported.
* Unsupported non-standard characters include [L, W, #, ?]
* _Does_ support / for intervals. Specifically the job will increment by the value of _b_ in _a_/_b_ starting with _a_.

### Day Of Month / Day Of Week Logic Table

|Day Of Month| Day Of Week |Output                                                                                  |
|------------|-------------|----------------------------------------------------------------------------------------|
|     *      |      *      |All days are included.                                                                  |
|     *      | Non * Value |Only days that match Day Of Week are included.                                          |
|Non * Value |      *      |Only days that match Day Of Month are included.                                         |
|Non * Value | Non * Value |All value that match Day Of Month or Day Of Year. Note: If * is included in either it can include all days and make the other irrelevant.|
