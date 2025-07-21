package historymanager

import (
	"testing"
	"time"

	"github.com/massalabs/node-manager-plugin/int/db"
	"github.com/massalabs/node-manager-plugin/int/utils"
	"github.com/stretchr/testify/assert"
)

func TestSampleValueHistory(t *testing.T) {
	totValuePostInterval := int64(180) // 3 minutes
	delAfter := int64(3600)            // 1 hour (not relevant for test)

	type testCase struct {
		name           string
		sampleNum      int64
		sampleInterval time.Duration
		getExpected    func(since time.Time) []ValueHistorySample
		setupDbMock    func(mockDB *db.MockDB, since time.Time)
		hasError       bool
	}

	now := time.Now().Truncate(time.Second)
	interval := time.Duration(totValuePostInterval) * time.Second

	val1 := 100.0
	val2 := 200.0
	val3 := 300.0
	val4 := 400.0
	val5 := 500.0
	val6 := 600.0
	val7 := 700.0

	tests := []testCase{
		{
			name:           "4 samples and 4 dbEntries (one before since), sampleInterval = totValuePostInterval",
			sampleNum:      4,
			sampleInterval: interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				dbEntries := []db.ValueHistory{
					{Timestamp: since.Add(-interval), TotalValue: val1},
					{Timestamp: since, TotalValue: val2},
					{Timestamp: since.Add(interval), TotalValue: val3},
					{Timestamp: since.Add(2 * interval), TotalValue: val4},
				}
				retrieveSince := since.Add(-interval)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: &val2},
					{Timestamp: since.Add(interval), Value: &val3},
					{Timestamp: since.Add(2 * interval), Value: &val4},
					{Timestamp: since.Add(3 * interval), Value: nil},
				}
			},
		},
		{
			name:           "sampleInterval is twice totValuePostInterval, 3 sampleNum and 7 dbEntries (one before since)",
			sampleNum:      3,
			sampleInterval: 2 * interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				dbEntries := []db.ValueHistory{
					{Timestamp: since.Add(-interval), TotalValue: val1},
					{Timestamp: since, TotalValue: val2},
					{Timestamp: since.Add(interval), TotalValue: val3},
					{Timestamp: since.Add(2 * interval), TotalValue: val4},
					{Timestamp: since.Add(3 * interval), TotalValue: val5},
					{Timestamp: since.Add(4 * interval), TotalValue: val6},
					{Timestamp: since.Add(5 * interval), TotalValue: val7},
				}
				retrieveSince := since.Add(-interval)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: &val2},
					{Timestamp: since.Add(2 * interval), Value: &val4},
					{Timestamp: since.Add(4 * interval), Value: &val6},
				}
			},
		},
		{
			name:           "sampleInterval is 1.5 totValuePostInterval, 4 samples, 7 dbEntries",
			sampleNum:      4,
			sampleInterval: time.Duration(3*totValuePostInterval/2) * time.Second,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				intv := time.Duration(totValuePostInterval) * time.Second
				dbEntries := []db.ValueHistory{
					{Timestamp: since, TotalValue: val1},
					{Timestamp: since.Add(intv), TotalValue: val2},
					{Timestamp: since.Add(2 * intv), TotalValue: val3},
					{Timestamp: since.Add(3 * intv), TotalValue: val4},
					{Timestamp: since.Add(4 * intv), TotalValue: val5},
					{Timestamp: since.Add(5 * intv), TotalValue: val6},
					{Timestamp: since.Add(6 * intv), TotalValue: val7},
				}
				retrieveSince := since.Add(-intv)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: &val1},
					{Timestamp: since.Add(time.Duration(3*totValuePostInterval/2) * time.Second), Value: &val2},
					{Timestamp: since.Add(time.Duration(3*totValuePostInterval) * time.Second), Value: &val4},
					{Timestamp: since.Add(time.Duration(9*totValuePostInterval/2) * time.Second), Value: &val5},
				}
			},
		},
		{
			name:           "1 single sample with a sampleInterval of 3.5 totValuePostInterval, 3 dbEntries",
			sampleNum:      1,
			sampleInterval: time.Duration(7*totValuePostInterval/2) * time.Second,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				intv := time.Duration(totValuePostInterval) * time.Second
				dbEntries := []db.ValueHistory{
					{Timestamp: since, TotalValue: val1},
					{Timestamp: since.Add(intv), TotalValue: val2},
					{Timestamp: since.Add(2 * intv), TotalValue: val3},
				}
				retrieveSince := since.Add(-intv)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: &val1},
				}
			},
		},
		{
			name:           "4 sample but only 2 dbEntries with timestamp: since + totValuePostInterval and since + 3*totValuePostInterval",
			sampleNum:      4,
			sampleInterval: interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				dbEntries := []db.ValueHistory{
					{Timestamp: since.Add(interval), TotalValue: val1},
					{Timestamp: since.Add(3 * interval), TotalValue: val2},
				}
				retrieveSince := since.Add(-interval)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: nil},
					{Timestamp: since.Add(interval), Value: &val1},
					{Timestamp: since.Add(2 * interval), Value: nil},
					{Timestamp: since.Add(3 * interval), Value: &val2},
				}
			},
		},
		{
			name:           "4 samples but only 2 dbEntries with timestamps: since and since + 2 totValuePostInterval",
			sampleNum:      4,
			sampleInterval: interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				dbEntries := []db.ValueHistory{
					{Timestamp: since, TotalValue: val1},
					{Timestamp: since.Add(2 * interval), TotalValue: val2},
				}
				retrieveSince := since.Add(-interval)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: &val1},
					{Timestamp: since.Add(interval), Value: nil},
					{Timestamp: since.Add(2 * interval), Value: &val2},
					{Timestamp: since.Add(3 * interval), Value: nil},
				}
			},
		},
		{
			name:           "4 samples and 3 dbEntries but there is 1/2 totValuePostInterval of diff between a sample timespan and a dbEntry one.",
			sampleNum:      4,
			sampleInterval: interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				dbEntries := []db.ValueHistory{
					{Timestamp: since.Add(interval / 2), TotalValue: val1},
					{Timestamp: since.Add(interval + interval/2), TotalValue: val2},
					{Timestamp: since.Add(2*interval + interval/2), TotalValue: val3},
				}
				retrieveSince := since.Add(-interval)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: nil},
					{Timestamp: since.Add(interval), Value: &val1},
					{Timestamp: since.Add(2 * interval), Value: &val2},
					{Timestamp: since.Add(3 * interval), Value: &val3},
				}
			},
		},
		{
			name:           "1 sample of interval 3*totValuePostInterval and 2 dbEntries after since",
			sampleNum:      1,
			sampleInterval: 3 * interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				dbEntries := []db.ValueHistory{
					{Timestamp: since.Add(interval), TotalValue: val1},
					{Timestamp: since.Add(2 * interval), TotalValue: val2},
				}
				retrieveSince := since.Add(-interval)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: nil},
				}
			},
		},
		{
			name:           "2 samples of interval totValuePostInterval, 2 dbEntries after since",
			sampleNum:      2,
			sampleInterval: interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				dbEntries := []db.ValueHistory{
					{Timestamp: since.Add(interval), TotalValue: val1},
					{Timestamp: since.Add(2 * interval), TotalValue: val2},
				}
				retrieveSince := since.Add(-interval)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return []ValueHistorySample{
					{Timestamp: since, Value: nil},
					{Timestamp: since.Add(interval), Value: &val1},
				}
			},
		},
		{
			name:           "mockDb returns no dbEntries",
			sampleNum:      2,
			sampleInterval: interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				dbEntries := []db.ValueHistory{}
				retrieveSince := since.Add(-interval)
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(dbEntries, nil)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return nil
			},
		},
		{
			name:           "mockDb returns an error",
			sampleNum:      2,
			sampleInterval: interval,
			setupDbMock: func(mockDB *db.MockDB, since time.Time) {
				retrieveSince := since.Add(-interval)
				dbErr := assert.AnError
				mockDB.On("GetHistory", retrieveSince, utils.NetworkBuildnet).Return(nil, dbErr)
			},
			getExpected: func(since time.Time) []ValueHistorySample {
				return nil
			},
			hasError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(db.MockDB)
			since := now.Add(-tc.sampleInterval * time.Duration(tc.sampleNum))
			if tc.setupDbMock != nil {
				tc.setupDbMock(mockDB, since)
			}
			mgr := NewHistoryManager(mockDB, delAfter, totValuePostInterval)
			got, err := mgr.SampleValueHistory(since, tc.sampleNum, false, tc.sampleInterval)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			expect := tc.getExpected(since)
			assert.Equal(t, expect, got)
			mockDB.AssertExpectations(t)
		})
	}
}
