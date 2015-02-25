package timer

// reference: https://github.com/robfig/cron
import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Field name   | Mandatory? | Allowed values | Allowed special characters
// ----------   | ---------- | -------------- | --------------------------
// Seconds      | No         | 0-59           | * / , -
// Minutes      | Yes        | 0-59           | * / , -
// Hours        | Yes        | 0-23           | * / , -
// Day of month | Yes        | 1-31           | * / , -
// Month        | Yes        | 1-12           | * / , -
// Day of week  | Yes        | 0-6            | * / , -
type CronExpr struct {
	sec   uint64
	min   uint64
	hour  uint64
	dom   uint64
	month uint64
	dow   uint64
}

func NewCronExpr(expr string) (cronExpr *CronExpr, err error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 && len(fields) != 6 {
		err = fmt.Errorf("invalid expr %v: expected 5 or 6 fields, got %v", expr, len(fields))
		return
	}

	if len(fields) == 5 {
		fields = append([]string{"0"}, fields...)
	}

	cronExpr = new(CronExpr)
	// Seconds
	cronExpr.sec, err = parseCronField(fields[0], 0, 59)
	if err != nil {
		goto onError
	}
	// Minutes
	cronExpr.min, err = parseCronField(fields[1], 0, 59)
	if err != nil {
		goto onError
	}
	// Hours
	cronExpr.hour, err = parseCronField(fields[2], 0, 23)
	if err != nil {
		goto onError
	}
	// Day of month
	cronExpr.dom, err = parseCronField(fields[3], 1, 31)
	if err != nil {
		goto onError
	}
	// Month
	cronExpr.month, err = parseCronField(fields[4], 1, 12)
	if err != nil {
		goto onError
	}
	// Day of week
	cronExpr.dow, err = parseCronField(fields[5], 0, 6)
	if err != nil {
		goto onError
	}
	return

onError:
	err = fmt.Errorf("invalid expr %v: %v", expr, err)
	return
}

// 1. *
// 2. num
// 3. num-num
// 4. */num
// 5. num/num (means num-max/num)
// 6. num-num/num
func parseCronField(field string, min int, max int) (cronField uint64, err error) {
	fields := strings.Split(field, ",")
	for _, field := range fields {
		rangeAndIncr := strings.Split(field, "/")
		if len(rangeAndIncr) > 2 {
			err = fmt.Errorf("too many slashes: %v", field)
			return
		}

		// range
		startAndEnd := strings.Split(rangeAndIncr[0], "-")
		if len(startAndEnd) > 2 {
			err = fmt.Errorf("too many hyphens: %v", rangeAndIncr[0])
			return
		}

		var start, end int
		if startAndEnd[0] == "*" {
			if len(startAndEnd) != 1 {
				err = fmt.Errorf("invalid range: %v", rangeAndIncr[0])
				return
			}
			start = min
			end = max
		} else {
			// start
			start, err = strconv.Atoi(startAndEnd[0])
			if err != nil {
				err = fmt.Errorf("invalid range: %v", rangeAndIncr[0])
				return
			}
			// end
			if len(startAndEnd) == 1 {
				if len(rangeAndIncr) == 2 {
					end = max
				} else {
					end = start
				}
			} else {
				end, err = strconv.Atoi(startAndEnd[1])
				if err != nil {
					err = fmt.Errorf("invalid range: %v", rangeAndIncr[0])
					return
				}
			}
		}

		if start > end {
			err = fmt.Errorf("invalid range: %v", rangeAndIncr[0])
			return
		}
		if start < min {
			err = fmt.Errorf("out of range [%v, %v]: %v", min, max, rangeAndIncr[0])
			return
		}
		if end > max {
			err = fmt.Errorf("out of range [%v, %v]: %v", min, max, rangeAndIncr[0])
			return
		}

		// increment
		var incr int
		if len(rangeAndIncr) == 1 {
			incr = 1
		} else {
			incr, err = strconv.Atoi(rangeAndIncr[1])
			if err != nil {
				err = fmt.Errorf("invalid increment: %v", rangeAndIncr[1])
				return
			}
			if incr <= 0 {
				err = fmt.Errorf("invalid increment: %v", rangeAndIncr[1])
				return
			}
		}

		// cronField
		if incr == 1 {
			cronField |= ^(math.MaxUint64 << uint(end+1)) & (math.MaxUint64 << uint(start))
		} else {
			for i := start; i <= end; i += incr {
				cronField |= 1 << uint(i)
			}
		}
	}

	return
}

func (e *CronExpr) Next(t time.Time) time.Time {
	return time.Time{}
}
