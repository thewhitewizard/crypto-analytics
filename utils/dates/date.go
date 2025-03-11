package dates

import "time"

const (
	DateFormat = "2006-01-02"
)

func GenerateDatesBetween2Dates(from string, to string, dateFormat string) []string {
	start, _ := time.Parse(dateFormat, from)
	end, err := time.Parse(dateFormat, to)

	if err != nil {
		end = time.Now()
	}

	var dates []string

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(dateFormat))
	}
	return dates
}

func GenerateDatesBetweenTwoDates(from time.Time, to time.Time) []time.Time {
	var dates []time.Time

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	return dates
}

func StringToDate(from string, dateFormat string) (time.Time, error) {
	return time.Parse(dateFormat, from)
}

func DateToString(from time.Time, dateFormat string) string {
	return from.Format(dateFormat)
}

// GetYesterdayTimestamps returns the start and end timestamps (Unix time in seconds) for yesterday's date
func GetYesterdayTimestamps() (int64, int64) {
	now := time.Now()

	// Get start of today (midnight)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Start of yesterday (midnight)
	startOfYesterday := startOfToday.AddDate(0, 0, -1)

	// End of yesterday (23:59:59)
	endOfYesterday := startOfToday.Add(-time.Second)

	startTimestamp := startOfYesterday.Unix()
	endTimestamp := endOfYesterday.Unix()

	return startTimestamp, endTimestamp
}
