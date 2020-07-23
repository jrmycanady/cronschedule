// Package cronschedule provides the functionality to parse and request execution times for a schedule provided in the
// cron format.
package cronschedule

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CronFieldValueRegex parses a single value fround in a cron field. Each field can have multiple values separated by
// commas. This regex specifically parses the values. Multiple values would need to be split first and then each value
// provided to the regex to parse a multiple value field.
// Group Index IDs
// 1 - [*] wildcard
// 2 - [*/#] wildcard with interval
// 3 - [#-#] numerical value range
// 4 - [#-#/#] Numerical value range with interval
// 5 - [#/#] Interval with start value
// 6 - [#] Numerical value
const CronFieldValueRegex = `(^\*$)|(^\*\/\d*$)|(^\d*-\d*$)|(^\d*-\d*\/\d*$)|(^\d*\/\d*$)|(^\d*$)`

var re = regexp.MustCompile(CronFieldValueRegex)

const FieldMinuteMin int = 0
const FieldMinuteMax int = 59
const FieldHourMin int = 0
const FieldHourMax int = 23
const FieldDayOfMonthMin int = 1
const FieldDayOfMonthMax int = 31
const FieldMonthMin int = 1
const FieldMonthMax int = 12
const FieldDayOfTheWeekMin int = 0
const FieldDayOfTheWeekMax int = 6

// Schedule is a cron schedule that has been parsed. It contains all the values for each field that are specified by the
// cron schedule.
type Schedule struct {
	Minutes      map[int]int
	MinutesSlice []int
	MinutesStr   []string

	Hours      map[int]int
	HoursSlice []int
	HoursStr   []string

	DaysOfMonth      map[int]int
	DaysOfMonthSlice []int
	DaysOfMonthStr   []string

	Months      map[int]int
	MonthsSlice []int
	MonthsStr   []string

	DaysOfTheWeek    map[int]int
	DaysOfWeekSlice  []int
	DaysOfTheWeekStr []string

	ScheduleStr string
}

// PrettyString generates a multi line string containing the schedule and values within it.
func (s *Schedule) PrettyString() string {
	prettyString := ""
	prettyString += fmt.Sprintf("Cron Schedule:     [%s]\n", s.ScheduleStr)
	prettyString += fmt.Sprintf("Minute:            %s => [%#v]\n", s.MinutesStr, sortMapKeys(s.Minutes))
	prettyString += fmt.Sprintf("Hour:              %s => [%#v]\n", s.HoursStr, sortMapKeys(s.Hours))
	prettyString += fmt.Sprintf("Days Of The Month: %s => [%#v]\n", s.DaysOfMonthStr, sortMapKeys(s.DaysOfMonth))
	prettyString += fmt.Sprintf("Month:             %s => [%#v]\n", s.MonthsStr, sortMapKeys(s.Months))
	prettyString += fmt.Sprintf("Day Of The Week:   %s => [%#v]\n", s.DaysOfTheWeekStr, sortMapKeys(s.DaysOfTheWeek))

	return prettyString
}

// ShouldExecute returns true if the schedule specifies it should execute at time t.
func (s *Schedule) ShouldExecute(t time.Time) bool {
	if _, ok := s.Minutes[t.Minute()]; !ok {
		return false
	}

	if _, ok := s.Hours[t.Hour()]; !ok {
		return false
	}

	if _, ok := s.Months[int(t.Month())]; !ok {
		return false
	}

	// Per POSIX spec the day of week and day of month are ORed...
	_, dayOfMonthOK := s.DaysOfMonth[t.Day()]
	_, dayOfWeekOK := s.DaysOfTheWeek[int(t.Weekday())]
	if !dayOfWeekOK && !dayOfMonthOK {
		return false
	}

	return true
}

// ShouldExecuteNow is the same as ShouldExecute but it uses the current time.
func (s *Schedule) ShouldExecuteNow() bool {
	return s.ShouldExecute(time.Now())
}

// computeStartValues computes the starting values for generating the closest schedule time for t. If the schedule
// directly aligns with t then the values related to t would be returned. In general t + 1second is generally provided
// as the result of t would always be in the past as seconds would be assumed to be zero.
func (s *Schedule) computeStartValues(t time.Time) (year int, monthIdx int, hourIdx int, minuteIdx int, day int) {
	tYear := t.Year()
	tMonth := t.Month()
	tDay := t.Day()
	tHour := t.Hour()
	tMinute := t.Minute()

	monthIdx = 0
	hourIdx = 0

	// Finding what the correct start month should be by looking at all valid months in the schedule.
	for monthIdx < len(s.MonthsSlice) {

		if s.MonthsSlice[monthIdx] > int(tMonth) {
			// The month found is now larger than the start month so the new start value would be this month and
			// the same year. All other field would start at zero.
			return tYear, monthIdx, 0, 0, 1
		}

		if s.MonthsSlice[monthIdx] == int(tMonth) {
			// Found the exact month so we need to lookup everything else.

			// Validate the day is a good stating point.
			_, dayOfMonthOK := s.DaysOfMonth[tDay]
			t := time.Date(tYear, tMonth, tDay, 0, 0, 0, 0, time.Local)
			_, dayOfWeekOK := s.DaysOfTheWeek[int(t.Weekday())]

			if dayOfWeekOK || dayOfMonthOK {

				// The day of week is valid so process hours.
				for hourIdx < len(s.HoursSlice) {

					if s.HoursSlice[hourIdx] > tHour {
						// The hour current index hour is past the provided out so send it along with a reset minute.
						return tYear, monthIdx, hourIdx, 0, tDay
					}

					if s.HoursSlice[hourIdx] == tHour {
						// The hour is correct so find the next minute.

						for minuteIdx < len(s.MinutesSlice) {
							if s.MinutesSlice[minuteIdx] >= tMinute {
								return tYear, monthIdx, hourIdx, minuteIdx, tDay
							}
						}
					}

					hourIdx++
				}
			}
			// The day of week was not valid so trying the next day.
			nextDay := tDay + 1
			if nextDay <= daysPerMonth(time.Month(s.MonthsSlice[monthIdx]), tYear) {
				return tYear, monthIdx, 0, 0, nextDay
			}
			// The next day loops to a new month so doing nothing.
		}

		monthIdx++
	}
	// The current month, nor a month after the current was found in the current year. Start the search at the beginning
	// of the next year.
	return tYear + 1, 0, 0, 0, 1

}

// NextExecutions returns a slice containing the times when the schedule should execute next.
func (s *Schedule) NextExecutions(t time.Time, count int) []time.Time {
	// execTimes will store all the resulting execution times found.
	execTimes := make([]time.Time, 0, count)

	t.Add(1 * time.Minute)
	// Computing the starting values for the generation algorithm.
	year, monthIdx, hourIdx, minuteIdx, day := s.computeStartValues(t.Add(1 * time.Minute))

	// Generating the next run time until total count is reached. Generation is performed by simply processing the
	// permutations of the known values. Days are an outlier due to the OR nature of day of the month and day of the week.
	var numFound int = 0

permutation:
	for numFound <= count {

		// Processing each supported month.
		for monthIdx < len(s.MonthsSlice) {
			month := s.MonthsSlice[monthIdx]

			// Processing the days.
			daysInMonth := daysPerMonth(time.Month(month), year)
			for day <= daysInMonth {

				_, dayOfMonthOK := s.DaysOfMonth[day]
				t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
				_, dayOfWeekOK := s.DaysOfTheWeek[int(t.Weekday())]
				if dayOfMonthOK || dayOfWeekOK {
					// Processing the hours.
					for hourIdx < len(s.HoursSlice) {
						hour := s.HoursSlice[hourIdx]

						for minuteIdx < len(s.MinutesSlice) {
							minute := s.MinutesSlice[minuteIdx]

							execT := time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local)
							execTimes = append(execTimes, execT)
							numFound++

							// Checking if we have the correct number and breaking early if so.  Waiting would result in
							// more than count returned.
							if numFound == count {
								break permutation
							}

							minuteIdx++
						}

						minuteIdx = 0
						hourIdx++
					}
				}

				hourIdx = 0
				minuteIdx = 0
				day++
			}

			// Starting at the first hour:minute:day of the next month.
			day = 1
			hourIdx = 0
			minuteIdx = 0
			monthIdx++
		}

		// Starting at the next month:hour:minute:day of the next year.
		monthIdx = 0
		day = 1
		hourIdx = 0
		minuteIdx = 0
		year++
	}
	return execTimes

}

// NextExecution returns the next time the schedule should be executed. It is a convenience method to return the next
// immediate execution time. It leverages NextExecutions() which should be used if multiple values are needed.
func (s *Schedule) NextExecution(t time.Time) time.Time {
	execTimes := s.NextExecutions(t, 1)
	return execTimes[0]
}

// daysPerMonth returns the number of days in the month for the year specified.
func daysPerMonth(month time.Month, year int) int {
	switch month {
	case time.January, time.March, time.May, time.July, time.August, time.October, time.December:
		return 31
	case time.April, time.June, time.September, time.November:
		return 30
	case time.February:
		leapTime := time.Date(year, time.December, 31, 0, 0, 0, 0, time.Local)
		if leapTime.YearDay() > 365 {
			return 29
		} else {
			return 28
		}
	default:
		panic("unknown month")
	}
}

// sortMapKeys sorts the keys of an int keyed map and returns a slice of the sorted keys.
func sortMapKeys(m map[int]int) []int {
	list := make([]int, 0, len(m))
	for k := range m {
		list = append(list, k)
	}
	sort.Ints(list)
	return list
}

// AddMinutes add the minutes listed to the schedule. Invalid values will be ignored.
func (s *Schedule) AddMinutes(minutes []int) {
	for _, i := range minutes {
		if i < FieldMinuteMin || i > FieldMinuteMax {
			continue
		}

		if _, ok := s.Minutes[i]; ok {
			s.Minutes[i] += 1
		} else {
			s.Minutes[i] = 1
		}
	}
}

// AddHours add the hours listed to the schedule. Invalid values will be ignored.
func (s *Schedule) AddHours(hours []int) {
	for _, i := range hours {
		if i < FieldHourMin || i > FieldHourMax {
			continue
		}

		if _, ok := s.Hours[i]; ok {
			s.Hours[i] += 1
		} else {
			s.Hours[i] = 1
		}
	}
}

// AddDaysOfMonth add the days of the month listed to the schedule. Invalid values will be ignored.
func (s *Schedule) AddDaysOfMonth(daysOfMonth []int) {
	for _, i := range daysOfMonth {
		if i < FieldDayOfMonthMin || i > FieldDayOfMonthMax {
			continue
		}

		if _, ok := s.DaysOfMonth[i]; ok {
			s.DaysOfMonth[i] += 1
		} else {
			s.DaysOfMonth[i] = 1
		}
	}
}

// AddMonths add the months listed to the schedule. Invalid values will be ignored.
func (s *Schedule) AddMonths(months []int) {
	for _, i := range months {
		if i < FieldMonthMin || i > FieldMonthMax {
			continue
		}

		if _, ok := s.Months[i]; ok {
			s.Months[i] += 1
		} else {
			s.Months[i] = 1
		}
	}
}

// AddDaysOfTheWeek add the days of the week listed to the schedule. Invalid values will be ignored.
func (s *Schedule) AddDaysOfTheWeek(daysOfTheWeek []int) {
	for _, i := range daysOfTheWeek {
		if i < FieldDayOfTheWeekMin || i > FieldDayOfTheWeekMax {
			continue
		}

		if _, ok := s.DaysOfTheWeek[i]; ok {
			s.DaysOfTheWeek[i] += 1
		} else {
			s.DaysOfTheWeek[i] = 1
		}
	}
}

// AddByIndex adds the values to the proper field based on the index.
func (s *Schedule) AddByIndex(values []int, index int) {
	switch index {
	case 0:
		s.AddMinutes(values)
	case 1:
		s.AddHours(values)
	case 2:
		s.AddDaysOfMonth(values)
	case 3:
		s.AddMonths(values)
	case 4:
		s.AddDaysOfTheWeek((values))
	}
}

// AddFieldStrByIndex adds the field Str value for the field at index.
func (s *Schedule) AddFieldStrByIndex(fieldStr string, index int) {
	switch index {
	case 0:
		s.MinutesStr = append(s.MinutesStr, fieldStr)
	case 1:
		s.HoursStr = append(s.HoursStr, fieldStr)
	case 2:
		s.DaysOfMonthStr = append(s.DaysOfMonthStr, fieldStr)
	case 3:
		s.MonthsStr = append(s.MonthsStr, fieldStr)
	case 4:
		s.DaysOfTheWeekStr = append(s.DaysOfTheWeekStr, fieldStr)
	}
}

// emptySchedule generates an empty schedule.
func EmptySchedule() Schedule {
	return Schedule{
		Minutes:          make(map[int]int),
		MinutesStr:       make([]string, 0, 0),
		MinutesSlice:     make([]int, 0, 0),
		Hours:            make(map[int]int),
		HoursStr:         make([]string, 0, 0),
		HoursSlice:       make([]int, 0, 0),
		DaysOfMonth:      make(map[int]int),
		DaysOfMonthStr:   make([]string, 0, 0),
		DaysOfMonthSlice: make([]int, 0, 0),
		Months:           make(map[int]int),
		MonthsStr:        make([]string, 0, 0),
		MonthsSlice:      make([]int, 0, 0),
		DaysOfTheWeek:    make(map[int]int),
		DaysOfTheWeekStr: make([]string, 0, 0),
		DaysOfWeekSlice:  make([]int, 0, 0),
		ScheduleStr:      "",
	}
}

// Parse will parse the cron schedule s and provide a Schedule ready to be used. If parsing fails an error will be
// provided.
// Parse only supports a full schedule so all 5 fields must be present.
func Parse(s string) (Schedule, error) {
	// Building the empty schedule that will be filled as parsing is completed.
	schedule := EmptySchedule()
	schedule.ScheduleStr = strings.TrimSpace(s)

	// Split the string by spaces to obtain each field. Expecting exactly 5 fields.
	fields := strings.Split(schedule.ScheduleStr, " ")
	if len(fields) != 5 {
		return schedule, fmt.Errorf("schedule should have 5 fields but found %d", len(fields))
	}

	// Process each field of the schedule working left to right so index 0 will be the minute while index 4 will be the
	// the day of the week.
	for i, field := range fields {

		// Checking for any empty values to prevent double spaces from being including in the entry.
		if field == "" {
			return schedule, fmt.Errorf("received empty value for field %s", FieldNameByIndex(i))
		}

		// Retrieving the min and max values for the current field which will be used to process the values
		// of the field.
		min, max, err := FieldMinMaxByIndex(i)
		if err != nil {
			return schedule, fmt.Errorf("failed to get min and max value for field %s: %s", FieldNameByIndex(i), err)
		}

		// Processing every value found in the field. This is specifically needed due to the multi value option
		// on fields.
		for _, value := range strings.Split(field, ",") {
			schedule.AddFieldStrByIndex(value, i)

			fieldValues, err := ParseFieldValue(value, min, max)
			if err != nil {
				return schedule, fmt.Errorf("failed to parse %s field with value of %s: %s", FieldNameByIndex(i), value, err)
			}

			schedule.AddByIndex(fieldValues, i)
		}
	}

	// Cleaning up day of week vs day of month wild card logic. By default the parser adds values for each as specified
	// the job description. Depending on the values of each the usable values in each list are changed.
	// |Day Of Month|Day Of Week|Result                                   |
	// |------------------------------------------------------------------|
	// |     *      |     *     |Both will be fully populated.            |
	// |     *      |     #     |Only Day Of Week will get populated.     |
	// |     #      |     *     |Only Day Of Month will get populated.    |
	//
	// NOTE: mutlivalue fields and interval fields containing * are undefined.
	if fields[2] == "*" && fields[4] == "*" {
		// TODO empty the day of week map. Update the processor to ignore building the time and checking day of week if
		// empty.
	}
	if fields[2] == "*" && fields[4] != "*" {
		schedule.DaysOfMonth = make(map[int]int)
	}
	if fields[2] != "*" && fields[4] == "*" {
		schedule.DaysOfTheWeek = make(map[int]int)
	}

	schedule.buildSlices()
	return schedule, nil
}

// buildSlices creates a sorted slice of the values for each field.
func (s *Schedule) buildSlices() {
	s.MinutesSlice = sortMapKeys(s.Minutes)
	s.HoursSlice = sortMapKeys(s.Hours)
	s.DaysOfMonthSlice = sortMapKeys(s.DaysOfMonth)
	s.MonthsSlice = sortMapKeys(s.Months)
	s.DaysOfWeekSlice = sortMapKeys(s.DaysOfTheWeek)
}

// ParseFieldValue parses a single value of a field and returns a slice of the values that are compassed by the field
// definition. If the field fails to parse an error is provided and the slice will be nil.
// The min and max values should be the min and max for the field being provided. The parser utilizes these values for
// validation and range generation.
func ParseFieldValue(value string, min int, max int) ([]int, error) {
	// Performing the regex match on the field. The match group determines the type of field provided and thus how to
	// parse it.
	match := re.FindAllStringSubmatch(value, -1)
	if match == nil {
		return nil, fmt.Errorf("[%s] is not in a supported field value format", value)
	}

	// Simplifying access to the match groups and doing some nil checking.
	if len(match) != 1 {
		panic("received a match with more than one match")
	}
	matchGroups := match[0]

	// Processing each format of the field as needed. The format is determined by the match group that is not nil
	// meaning it matched. Match group 0 is ignored as it's the just the full match.
	switch {
	case matchGroups[1] != "":
		// [*]
		values, err := GenerateValueSlice(min, max, 1, min, max)
		if err != nil {
			return nil, fmt.Errorf("failed to build values for [%s]: %s", matchGroups[1], err)
		}

		return values, nil

	case matchGroups[2] != "":
		// [*/#]
		params := strings.Split(matchGroups[2], "/")
		if len(params) != 2 {
			panic(fmt.Sprintf("regex matched [*/#] but failed to split [%s] on [/] into two strings, instead recieved %d", matchGroups[2], len(params)))
		}

		interval, err := strconv.Atoi(params[1])
		if err != nil {
			panic(fmt.Sprintf("regex matched [*/#] but failed to convert the # value of [%s] to integer: %s", params[1], err))
		}

		values, err := GenerateValueSlice(min, max, interval, min, max)
		if err != nil {
			return nil, fmt.Errorf("failed to build values for [%s]: %s", matchGroups[2], err)
		}

		return values, nil

	case matchGroups[3] != "":
		// [#-#]
		params := strings.Split(matchGroups[3], "-")
		if len(params) != 2 {
			panic(fmt.Sprintf("regex matched [#-#] but failed to split [%s] on [-] into two strings, instead recieved %d", matchGroups[3], len(params)))
		}

		startRange, err := strconv.Atoi(params[0])
		if err != nil {
			panic(fmt.Sprintf("regex matched [#-#] but failed to convert the first # value of [%s] to integer: %s", params[0], err))
		}

		endRange, err := strconv.Atoi(params[1])
		if err != nil {
			panic(fmt.Sprintf("regex matched [#-#] but failed to convert the second # value of [%s] to integer: %s", params[1], err))
		}

		values, err := GenerateValueSlice(startRange, endRange, 1, min, max)
		if err != nil {
			return nil, fmt.Errorf("failed to build values for [%s]: %s", matchGroups[3], err)
		}

		return values, nil

	case matchGroups[4] != "":
		// [#-#/#]
		components := strings.Split(matchGroups[4], "/")
		if len(components) != 2 {
			panic(fmt.Sprintf("regex matched [#-#/#] but failed to split [%s] on [/] into two strings, instead recieved %d", matchGroups[4], len(components)))
		}

		interval, err := strconv.Atoi(components[1])
		if err != nil {
			panic(fmt.Sprintf("regex matched [#-#/#] but failed to convert the interval valuee of [%s] to integer: %s", components[1], err))
		}

		params := strings.Split(components[0], "-")
		if len(params) != 2 {
			panic(fmt.Sprintf("regex matched [#-#/#] but failed to split [%s] on [-] into two strings, instead recieved %d", components[0], len(params)))
		}

		startRange, err := strconv.Atoi(params[0])
		if err != nil {
			panic(fmt.Sprintf("regex matched [#-#/#] but failed to convert the first range # value of [%s] to integer: %s", params[0], err))
		}

		endRange, err := strconv.Atoi(params[1])
		if err != nil {
			panic(fmt.Sprintf("regex matched [#-#/#] but failed to convert the second range # value of [%s] to integer: %s", params[1], err))
		}

		values, err := GenerateValueSlice(startRange, endRange, interval, min, max)
		if err != nil {
			return nil, fmt.Errorf("failed to build values for [%s]: %s", matchGroups[3], err)
		}

		return values, nil

	case matchGroups[5] != "":
		// [#/#]
		params := strings.Split(matchGroups[5], "/")
		if len(params) != 2 {
			panic(fmt.Sprintf("regex matched [#/#] but failed to split [%s] on [/] into two strings, instead recieved %d", matchGroups[5], len(params)))
		}

		startRange, err := strconv.Atoi(params[0])
		if err != nil {
			panic(fmt.Sprintf("regex matched [#/#] but failed to convert the start # value of [%s] to integer: %s", params[0], err))
		}

		interval, err := strconv.Atoi(params[1])
		if err != nil {
			panic(fmt.Sprintf("regex matched [#/#] but failed to convert the interval # value of [%s] to integer: %s", params[1], err))
		}

		values, err := GenerateValueSlice(startRange, max, interval, min, max)
		if err != nil {
			return nil, fmt.Errorf("failed to build values for [%s]: %s", matchGroups[3], err)
		}

		return values, nil

	case matchGroups[6] != "":
		// [#]
		singleValue, err := strconv.Atoi(matchGroups[6])
		if err != nil {
			panic(fmt.Sprintf("regex matched [#] but failed to convert the # value of [%s] to integer: %s", matchGroups[6], err))
		}

		values, err := GenerateValueSlice(singleValue, singleValue, 1, min, max)
		if err != nil {
			return nil, fmt.Errorf("failed to build values for [%s]: %s", matchGroups[3], err)
		}

		return values, nil

	default:
		panic("field matched without a match group found")
	}

}

// FieldNameByIndex returns the name of the filed based on the index i provided.
func FieldNameByIndex(i int) string {
	switch i {
	case 0:
		return "minute"
	case 1:
		return "hour"
	case 2:
		return "day of month"
	case 3:
		return "month"
	case 4:
		return "day of week"
	default:
		return "invalid"
	}
}

// FieldMinMaxByIndex returns the minimum and maximum value for the field specified by the index.
func FieldMinMaxByIndex(i int) (min int, max int, err error) {
	switch i {
	case 0:
		return FieldMinuteMin, FieldMinuteMax, nil
	case 1:
		return FieldHourMin, FieldHourMax, nil
	case 2:
		return FieldDayOfMonthMin, FieldDayOfMonthMax, nil
	case 3:
		return FieldMonthMin, FieldMonthMax, nil
	case 4:
		return FieldDayOfTheWeekMin, FieldDayOfTheWeekMax, nil
	default:
		return min, max, fmt.Errorf("unknown index %d", i)
	}
}

// GenerateValueSlice generates a slice of all values specified by the range, interval, and min/max.
func GenerateValueSlice(rangeStart int, rangeEnd int, interval int, fieldMin int, fieldMax int) ([]int, error) {

	// Rejecting any intervals that would result in the value not incrementing upwards.
	if interval <= 0 {
		return nil, fmt.Errorf("interval cannot be <= 0")
	}

	// Validating the range is specified from smaller to larger values.
	if rangeStart > rangeEnd {

		return nil, fmt.Errorf("range start value of [%d] is larger than range end value of [%d]", rangeStart, rangeEnd)
	}

	// Validating that the range start provided exists within the min provided.
	if rangeStart < fieldMin {
		return nil, fmt.Errorf("range start value of [%d] is below the field minimum value of [%d]", rangeStart, fieldMin)
	}

	// Validating that the range end provided exists within the max provided.
	if rangeEnd > fieldMax {
		return nil, fmt.Errorf("range end value of [%d] is larger than the field maximum value of [%d]", rangeEnd, fieldMax)
	}

	// NOTE: we do not need to check that the rangeStart is larger than fieldMax, nor do we need to check that rangeEnd
	// is smaller than fieldMin. The combination of the checks above would reject any situation were this check would
	// fail.

	// Build value list.
	values := make([]int, 0, 0)
	value := rangeStart
	for value <= rangeEnd {
		values = append(values, value)

		value = value + interval
	}

	return values, nil
}
