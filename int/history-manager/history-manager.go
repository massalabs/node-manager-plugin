package historymanager

import (
	"time"

	"github.com/massalabs/node-manager-plugin/int/db"
	"github.com/massalabs/node-manager-plugin/int/utils"
)

type ValueHistorySample struct {
	Timestamp time.Time `json:"timestamp"`
	Value     *float64  `json:"value"`
}

type HistoryManager struct {
	db                   db.DB
	delAfter             int64 // seconds
	totValuePostInterval int64 // seconds
}

var globalManager *HistoryManager

func NewHistoryManager(database db.DB, delAfter int64, totValuePostInterval int64) *HistoryManager {
	mgr := &HistoryManager{db: database, delAfter: delAfter, totValuePostInterval: totValuePostInterval}
	globalManager = mgr
	return mgr
}

func init() {
	if globalManager != nil {
		cutoff := time.Now().Add(-time.Duration(globalManager.delAfter) * time.Second)
		_ = globalManager.db.DeleteOldValueHistory(cutoff)
	}
}

// SampleValueHistory returns sampleNum samples between since and now, using the closest value <= sample timestamp, not reused.
func (mgr *HistoryManager) SampleValueHistory(since time.Time, sampleNum int64, isMainnet bool, interval time.Duration) ([]ValueHistorySample, error) {
	net := utils.NetworkBuildnet
	if isMainnet {
		net = utils.NetworkMainnet
	}

	// Retrieve values from since - totValuePostInterval to ensure we have enough data
	retrieveSince := since.Add(-time.Duration(mgr.totValuePostInterval) * time.Second)
	dbEntries, err := mgr.db.GetHistory(retrieveSince, net)
	if err != nil {
		return nil, err
	}

	lenDbEntries := len(dbEntries)
	if lenDbEntries == 0 {
		return nil, nil
	}

	//now := time.Now()

	// interval := now.Sub(since) / time.Duration(sampleNum)
	// used := make(map[int]bool)
	result := make([]ValueHistorySample, sampleNum)
	dbEntryIndex := 0
	for i := int64(0); i < sampleNum; i++ {
		ts := since.Add(time.Duration(i) * interval)
		var val *float64
		chosen := -1

		for dbEntryIndex < lenDbEntries {
			if dbEntries[dbEntryIndex].Timestamp.After(ts) {
				break
			}
			chosen = dbEntryIndex
			dbEntryIndex++
		}

		// for j, entry := range dbEntries {
		// 	if used[j] {
		// 		continue
		// 	}

		// 	if chosen == -1 || entry.Timestamp.After(dbEntries[chosen].Timestamp) {
		// 		chosen = j
		// 	}
		// }
		if chosen != -1 {
			v := dbEntries[chosen].TotalValue
			val = &v
			// used[chosen] = true
		}
		result[i] = ValueHistorySample{Timestamp: ts, Value: val}
	}
	return result, nil
}
