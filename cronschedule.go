// Package cronschedule provides the functionality to parse and request execution times for a schedule provided in the
// cron format.
package cronschedule

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	Minutes    map[int]int
	MinutesStr []string

	Hours    map[int]int
	HoursStr []string

	DaysOfMonth    map[int]int
	DaysOfMonthStr []string

	Months    map[int]int
	MonthsStr []string

	DaysOfTheWeek    map[int]int
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

// sortMapKeys sorts the keys of an int keyed map and returns a slice of the sorted keys.
func sortMapKeys(m map[int]int) []int {
	list := make([]int, 0, len(m))
	for k, _ := range m {
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
		Hours:            make(map[int]int),
		HoursStr:         make([]string, 0, 0),
		DaysOfMonth:      make(map[int]int),
		DaysOfMonthStr:   make([]string, 0, 0),
		Months:           make(map[int]int),
		MonthsStr:        make([]string, 0, 0),
		DaysOfTheWeek:    make(map[int]int),
		DaysOfTheWeekStr: make([]string, 0, 0),
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

	return schedule, nil
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
