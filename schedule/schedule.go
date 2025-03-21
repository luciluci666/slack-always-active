package schedule

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Schedule struct {
	workDays  []time.Weekday
	startTime time.Time
	endTime   time.Time
	offset    int // GMT offset in hours
}

func parseTime(timeStr string) (time.Time, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour := 0
	minute := 0
	if _, err := fmt.Sscanf(parts[0], "%d", &hour); err != nil {
		return time.Time{}, fmt.Errorf("invalid hour: %s", parts[0])
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &minute); err != nil {
		return time.Time{}, fmt.Errorf("invalid minute: %s", parts[1])
	}

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("time out of range: %s", timeStr)
	}

	return time.Date(2000, 1, 1, hour, minute, 0, 0, time.UTC), nil
}

func parseWorkDays(daysStr string) ([]time.Weekday, error) {
	days := strings.Split(daysStr, ",")
	workDays := make([]time.Weekday, 0, len(days))

	for _, day := range days {
		day = strings.TrimSpace(day)
		switch strings.ToLower(day) {
		case "monday":
			workDays = append(workDays, time.Monday)
		case "tuesday":
			workDays = append(workDays, time.Tuesday)
		case "wednesday":
			workDays = append(workDays, time.Wednesday)
		case "thursday":
			workDays = append(workDays, time.Thursday)
		case "friday":
			workDays = append(workDays, time.Friday)
		case "saturday":
			workDays = append(workDays, time.Saturday)
		case "sunday":
			workDays = append(workDays, time.Sunday)
		default:
			return nil, fmt.Errorf("invalid day: %s", day)
		}
	}

	if len(workDays) == 0 {
		return nil, fmt.Errorf("no valid work days specified")
	}

	return workDays, nil
}

func parseGMTOffset(offsetStr string) (int, error) {
	if offsetStr == "" {
		return 0, nil // Default to UTC
	}

	// Remove any spaces and convert to lowercase
	offsetStr = strings.TrimSpace(strings.ToLower(offsetStr))

	// Handle GMT+/-HH format
	var sign int
	var hours int
	if strings.HasPrefix(offsetStr, "gmt+") {
		sign = 1
		offsetStr = offsetStr[4:]
	} else if strings.HasPrefix(offsetStr, "gmt-") {
		sign = -1
		offsetStr = offsetStr[4:]
	} else {
		return 0, fmt.Errorf("invalid GMT offset format: %s", offsetStr)
	}

	if _, err := fmt.Sscanf(offsetStr, "%d", &hours); err != nil {
		return 0, fmt.Errorf("invalid hours in GMT offset: %s", offsetStr)
	}

	if hours < 0 || hours > 23 {
		return 0, fmt.Errorf("GMT offset hours must be between 0 and 23: %d", hours)
	}

	return sign * hours, nil
}

func NewSchedule() (*Schedule, error) {
	workDaysStr := os.Getenv("WORK_DAYS")
	startTimeStr := os.Getenv("WORK_START")
	endTimeStr := os.Getenv("WORK_END")
	gmtOffsetStr := os.Getenv("GMT_OFFSET")

	if workDaysStr == "" {
		workDaysStr = "monday,tuesday,wednesday,thursday,friday"
	}
	if startTimeStr == "" {
		startTimeStr = "09:00"
	}
	if endTimeStr == "" {
		endTimeStr = "17:00"
	}

	workDays, err := parseWorkDays(workDaysStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse work days: %v", err)
	}

	startTime, err := parseTime(startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %v", err)
	}

	endTime, err := parseTime(endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %v", err)
	}

	offset, err := parseGMTOffset(gmtOffsetStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GMT offset: %v", err)
	}

	return &Schedule{
		workDays:  workDays,
		startTime: startTime,
		endTime:   endTime,
		offset:    offset,
	}, nil
}

func (s *Schedule) IsWorkingTime() bool {
	now := time.Now().UTC()
	weekday := now.Weekday()

	// Adjust current time by GMT offset
	adjustedHour := (now.Hour() + s.offset + 24) % 24
	currentTime := time.Date(2000, 1, 1, adjustedHour, now.Minute(), 0, 0, time.UTC)

	isWorkDay := false
	for _, day := range s.workDays {
		if weekday == day {
			isWorkDay = true
			break
		}
	}
	if !isWorkDay {
		return false
	}

	return currentTime.After(s.startTime) && currentTime.Before(s.endTime)
}

func (s *Schedule) GetNextWorkingTime() time.Time {
	now := time.Now().UTC()
	weekday := now.Weekday()

	// Adjust current time by GMT offset
	adjustedHour := (now.Hour() + s.offset + 24) % 24
	currentTime := time.Date(2000, 1, 1, adjustedHour, now.Minute(), 0, 0, time.UTC)

	if currentTime.After(s.startTime) && currentTime.Before(s.endTime) {
		// Calculate end time in UTC
		endHour := (s.endTime.Hour() - s.offset + 24) % 24
		return time.Date(now.Year(), now.Month(), now.Day(), endHour, s.endTime.Minute(), 0, 0, time.UTC)
	}

	nextWorkDay := weekday
	daysToAdd := 0
	for {
		nextWorkDay = (nextWorkDay + 1) % 7
		daysToAdd++
		for _, day := range s.workDays {
			if nextWorkDay == day {
				// Calculate start time in UTC
				startHour := (s.startTime.Hour() - s.offset + 24) % 24
				return now.AddDate(0, 0, daysToAdd).UTC().Truncate(24 * time.Hour).Add(time.Duration(startHour)*time.Hour + time.Duration(s.startTime.Minute())*time.Minute)
			}
		}
	}
}

// GetOffset returns the GMT offset in hours
func (s *Schedule) GetOffset() int {
	return s.offset
}
