package mapper

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestUnixTimeRoundTrip(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)
	require.Equal(t, ts, timeFromUnix(unixFromTime(ts)))
	require.Equal(t, int64(0), unixFromTime(time.Time{}))
	require.True(t, timeFromUnix(0).IsZero())
}

func TestTextFromString(t *testing.T) {
	t.Parallel()

	empty := textFromString("")
	require.False(t, empty.Valid)

	got := textFromString("hello")
	require.True(t, got.Valid)
	require.Equal(t, "hello", got.String)
	require.Equal(t, "hello", stringFromText(got))
	require.Empty(t, stringFromText(pgtype.Text{}))
}

func TestInt8FromTime(t *testing.T) {
	t.Parallel()

	require.False(t, int8FromTime(time.Time{}).Valid)

	ts := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)
	got := int8FromTime(ts)
	require.True(t, got.Valid)
	require.Equal(t, ts.Unix(), got.Int64)
	require.Equal(t, ts, timeFromInt8(got))
	require.True(t, timeFromInt8(pgtype.Int8{}).IsZero())
}
