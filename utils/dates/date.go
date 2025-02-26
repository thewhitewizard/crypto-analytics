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
