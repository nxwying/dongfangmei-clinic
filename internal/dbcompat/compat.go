// Package dbcompat provides cross-database SQL helpers.
package dbcompat

import "gorm.io/gorm"

var driver string

// Init stores the active database driver for runtime SQL selection.
func Init(db *gorm.DB) {
	dialector := db.Dialector.Name()
	switch dialector {
	case "sqlite":
		driver = "sqlite"
	default:
		driver = "postgres"
	}
}

// IsSQLite returns true when the backend is SQLite.
func IsSQLite() bool { return driver == "sqlite" }

// MonthExpr returns "YYYY-MM" from a timestamp column.
func MonthExpr(col string) string {
	if IsSQLite() {
		return "strftime('%Y-%m', " + col + ")"
	}
	return "to_char(date_trunc('month', " + col + "), 'YYYY-MM')"
}

// DateExpr extracts the date portion from a timestamp.
func DateExpr(col string) string {
	return "DATE(" + col + ")"
}

// NowExpr returns the current timestamp.
func NowExpr() string {
	if IsSQLite() {
		return "datetime('now')"
	}
	return "NOW()"
}

// TodayExpr returns today's date.
func TodayExpr() string {
	if IsSQLite() {
		return "date('now')"
	}
	return "CURRENT_DATE"
}

// DaysAgoExpr returns a date N days ago.
func DaysAgoExpr(days int) string {
	if IsSQLite() {
		return "date('now', '-" + intToStr(days) + " days')"
	}
	return "CURRENT_DATE - INTERVAL '" + intToStr(days) + " days'"
}

// MonthsAgoExpr returns a timestamp N months ago.
func MonthsAgoExpr(months int) string {
	if IsSQLite() {
		return "datetime('now', '-" + intToStr(months) + " months')"
	}
	return "NOW() - INTERVAL '" + intToStr(months) + " months'"
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
