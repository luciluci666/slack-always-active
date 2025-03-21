package schedule

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Schedule struct {
	workDays []time.Weekday
	start    time.Time
	end      time.Time
	offset   int
}

func parseTime(timeStr string) (time.Time, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid minute: %s", parts[1])
	}

	// Create time for today with the specified hour and minute
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC), nil
}

func parseWorkDays(daysStr string) ([]time.Weekday, error) {
	if daysStr == "" {
		// Default to Monday-Friday
		return []time.Weekday{
			time.Monday,
			time.Tuesday,
			time.Wednesday,
			time.Thursday,
			time.Friday,
		}, nil
	}

	days := strings.Split(daysStr, ",")
	workDays := make([]time.Weekday, 0, len(days))

	for _, day := range days {
		day = strings.TrimSpace(strings.ToLower(day))
		switch day {
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

	return workDays, nil
}

func parseOffset(offsetStr string) (int, error) {
	if offsetStr == "" {
		return 0, nil // Default to UTC
	}

	// Remove "GMT" prefix if present
	offsetStr = strings.TrimPrefix(offsetStr, "GMT")
	offsetStr = strings.TrimPrefix(offsetStr, "gmt")

	// Parse the offset
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0, fmt.Errorf("invalid GMT offset: %s", offsetStr)
	}

	return offset, nil
}

func NewSchedule() (*Schedule, error) {
	// Get work days from environment
	workDays, err := parseWorkDays(os.Getenv("WORK_DAYS"))
	if err != nil {
		return nil, fmt.Errorf("error parsing work days: %v", err)
	}

	// Get work hours from environment
	startTime, err := parseTime(os.Getenv("WORK_START"))
	if err != nil {
		return nil, fmt.Errorf("error parsing start time: %v", err)
	}

	endTime, err := parseTime(os.Getenv("WORK_END"))
	if err != nil {
		return nil, fmt.Errorf("error parsing end time: %v", err)
	}

	// Get GMT offset from environment
	offset, err := parseOffset(os.Getenv("GMT_OFFSET"))
	if err != nil {
		return nil, fmt.Errorf("error parsing GMT offset: %v", err)
	}

	return &Schedule{
		workDays: workDays,
		start:    startTime,
		end:      endTime,
		offset:   offset,
	}, nil
}

func (s *Schedule) IsWorkingTime() bool {
	// Get current time in UTC
	now := time.Now().UTC()

	// Adjust current time by the GMT offset
	now = now.Add(time.Duration(s.offset) * time.Hour)

	// Check if current day is a work day
	isWorkDay := false
	for _, day := range s.workDays {
		if now.Weekday() == day {
			isWorkDay = true
			break
		}
	}

	if !isWorkDay {
		return false
	}

	// Create current time with only hour and minute for comparison
	currentTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
	startTime := time.Date(now.Year(), now.Month(), now.Day(), s.start.Hour(), s.start.Minute(), 0, 0, time.UTC)
	endTime := time.Date(now.Year(), now.Month(), now.Day(), s.end.Hour(), s.end.Minute(), 0, 0, time.UTC)

	// Check if current time is within work hours
	return currentTime.After(startTime) && currentTime.Before(endTime)
}

func (s *Schedule) GetNextWorkingTime() time.Time {
	now := time.Now().UTC()
	now = now.Add(time.Duration(s.offset) * time.Hour)

	// Create current time with only hour and minute for comparison
	currentTime := time.Date(2000, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)
	startTime := time.Date(2000, 1, 1, s.start.Hour(), s.start.Minute(), 0, 0, time.UTC)
	endTime := time.Date(2000, 1, 1, s.end.Hour(), s.end.Minute(), 0, 0, time.UTC)

	// If we're before start time today, return start time today
	if currentTime.Before(startTime) {
		nextTime := time.Date(now.Year(), now.Month(), now.Day(), s.start.Hour(), s.start.Minute(), 0, 0, time.UTC)
		return nextTime.Add(-time.Duration(s.offset) * time.Hour)
	}

	// If we're after end time today, find next work day
	if currentTime.After(endTime) {
		// Start from tomorrow
		nextDay := now.Add(24 * time.Hour)
		for {
			// Check if this is a work day
			isWorkDay := false
			for _, day := range s.workDays {
				if nextDay.Weekday() == day {
					isWorkDay = true
					break
				}
			}

			if isWorkDay {
				// Return start time on this day
				nextTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), s.start.Hour(), s.start.Minute(), 0, 0, time.UTC)
				return nextTime.Add(-time.Duration(s.offset) * time.Hour)
			}

			nextDay = nextDay.Add(24 * time.Hour)
		}
	}

	// If we're within working hours, return current time
	return now.Add(-time.Duration(s.offset) * time.Hour)
}

func (s *Schedule) GetOffset() int {
	return s.offset
}
