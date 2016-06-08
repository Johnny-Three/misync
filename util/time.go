package util

import (
	//"fmt"
	"time"
)

//传入timestamp，返回是否在同一小时，在 true ，不在 false
func JudgeInSameHour(begin, end int64) bool {

	t1 := time.Unix(begin, 0)
	t2 := time.Unix(end, 0)
	if t1.Hour() == t2.Hour() {
		return true
	}
	return false
}

func lastDayOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location())
}

func firstDayOfNextYear(t time.Time) time.Time {
	return time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, t.Location())
}

//传入timestamp，如果跨天，则返回值大于1,否则为0
func DaysDiff(end, begin int64) (days int) {

	a := time.Unix(end, 0)
	b := time.Unix(begin, 0)

	cur := b
	for cur.Year() < a.Year() {
		// add 1 to count the last day of the year too.
		days += lastDayOfYear(cur).YearDay() - cur.YearDay() + 1
		cur = firstDayOfNextYear(cur)
	}
	days += a.YearDay() - cur.YearDay()
	if b.AddDate(0, 0, days).After(a) {
		//days -= 1
	}
	return days
}
