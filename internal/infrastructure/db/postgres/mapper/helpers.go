package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/exp/constraints"
)

func unixFromTime(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UTC().Unix()
}

func timeFromUnix[T constraints.Integer](sec T) time.Time {
	if sec == 0 {
		return time.Time{}
	}
	return time.Unix(int64(sec), 0).UTC()
}

func textFromString(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func stringFromText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func int8FromTime(t time.Time) pgtype.Int8 {
	if t.IsZero() {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: unixFromTime(t), Valid: true}
}

func timeFromInt8(v pgtype.Int8) time.Time {
	if !v.Valid {
		return time.Time{}
	}
	return timeFromUnix(v.Int64)
}
