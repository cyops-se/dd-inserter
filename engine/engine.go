package engine

import (
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
)

type groupSequenceInfo struct {
	groupName      string
	sequenceNumber types.VolatileDataPoint
}

var dpLock sync.Mutex
var datapoints map[string]*types.VolatileDataPoint // = make(map[string]*types.VolatileDataPoint)
var groupseq map[string]*groupSequenceInfo
var NewMsg chan types.DataMessage = make(chan types.DataMessage, 2000)
var NewMeta chan []*types.DataPointMeta = make(chan []*types.DataPointMeta)
var NewEmitMsg chan types.DataPoint = make(chan types.DataPoint, 2000)
var NewEmitMetaMsg chan types.DataPointMeta = make(chan types.DataPointMeta, 2000)

// var totalReceived types.DataPoint
// var totalUpdated types.DataPoint
var totalReceived types.VolatileDataPoint
var totalUpdated types.VolatileDataPoint

// var sequenceNumber types.VolatileDataPoint

func UpdateDataPointMeta(item *types.DataPointMeta) (err error) {
	return
}

func GetDataPoints() ([]types.VolatileDataPoint, error) {
	list := make([]types.VolatileDataPoint, len(datapoints))
	i := 0
	for _, item := range datapoints {
		list[i] = *item
		i++
	}
	return list, nil
}

func InitDispatchers() {
	InitDataPointMap()
	groupseq = make(map[string]*groupSequenceInfo)

	totalReceived.DataPoint = &types.DataPoint{Name: "Total Received", Value: uint64(0)}
	datapoints[totalReceived.DataPoint.Name] = &totalReceived

	totalUpdated.DataPoint = &types.DataPoint{Name: "Total Updated", Value: uint64(0)}
	datapoints[totalUpdated.DataPoint.Name] = &totalUpdated

	go runDataDispatch()
	// go runMetaDispatch() // disable temporarily until fixed
}

func InitDataPointMap() {
	var items []types.DataPointMeta
	db.DB.Find(&items)

	datapoints = make(map[string]*types.VolatileDataPoint)

	for _, item := range items {
		if _, ok := datapoints[item.Name]; !ok {
			vp := &types.VolatileDataPoint{DataPointMeta: item}
			datapoints[item.Name] = vp
		}

		datapoints[item.Name].DataPointMeta = item
	}
}

func runDataDispatch() {
	for {
		msg := <-NewMsg

		// Handle sequence number montioring and alerting
		gi, ok := groupseq[msg.Group]
		if !ok {
			gi = &groupSequenceInfo{groupName: msg.Group}
			gi.sequenceNumber.DataPoint = &types.DataPoint{Name: "Sequence Number: " + gi.groupName, Value: uint64(0)}
			datapoints[gi.sequenceNumber.DataPoint.Name] = &gi.sequenceNumber
			groupseq[msg.Group] = gi
		}

		diff := math.Abs(float64(msg.Sequence) - float64(gi.sequenceNumber.DataPoint.Value.(uint64)))
		if diff > 5.0 && gi.sequenceNumber.DataPoint.Value.(uint64) > 0 {
			db.Error("Engine data dispatch", "Sequence number out of sync for group %s, received %d, had %d, difference: %d", gi.groupName,
				msg.Sequence, gi.sequenceNumber.DataPoint.Value, uint(diff))
			SendAlerts()
		} else {
			// Update sequenceNumber
			gi.sequenceNumber.DataPoint.Value = msg.Sequence
			gi.sequenceNumber.DataPoint.Time = time.Now().UTC().Add(time.Nanosecond * time.Duration(rand.Uint64()))
			gi.sequenceNumber.LastEmitted = gi.sequenceNumber.DataPoint.Time
			// NewEmitMsg <- *gi.sequenceNumber.DataPoint
		}

		gi.sequenceNumber.DataPoint.Value = msg.Sequence

		// Update internal data point table
		dpLock.Lock()
		for _, dp := range msg.Points {

			// Filter out empty value with quality zero
			if dp.Quality == 0 {
				if v, ok := dp.Value.(float64); ok && v == 0 {
					continue
				}
			}

			entry, ok := datapoints[dp.Name]
			if !ok {
				entry = createDataPointEntry(&dp)
			}

			if strings.TrimSpace(entry.DataPointMeta.Alias) != "" {
				dp.Name = entry.DataPointMeta.Alias
			}

			entry.DataPoint = &dp
			switch updateType := entry.DataPointMeta.UpdateType; updateType {
			case types.UpdateTypePassthru:
				entry.LastEmitted = time.Now().UTC()
				NewEmitMsg <- dp
				totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(uint64) + 1

			case types.UpdateTypeDeadband:
				// Check deadband
				if _, ok := dp.Value.(float64); ok {
					value := dp.Value.(float64)
					difference := calculateDifference(entry, value)
					entry.Integrator += (value - entry.StoredValue)
					entry.Threshold = difference * entry.DataPointMeta.IntegratingDeadband
					if math.Abs(entry.Integrator) > entry.Threshold {
						entry.StoredValue = value
						entry.Integrator = 0.0
						entry.LastEmitted = time.Now().UTC()
						NewEmitMsg <- dp
						totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(uint64) + 1
					}
				}

			case types.UpdateTypeInterval:
				if time.Since(entry.LastEmitted) > time.Duration(entry.DataPointMeta.Interval)*time.Second {
					entry.LastEmitted = time.Now().UTC()
					NewEmitMsg <- dp
					totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(uint64) + 1
				}
			}

			NotifySubscribers("meta.entry", entry)
		}

		dpLock.Unlock()

		// Update totalRecevied and totalUpdated
		totalReceived.DataPoint.Value = totalReceived.DataPoint.Value.(uint64) + uint64(msg.Count)
		if time.Now().UTC().Sub(totalReceived.DataPoint.Time) > time.Second {
			totalReceived.DataPoint.Time = time.Now().UTC()
			totalReceived.LastEmitted = totalReceived.DataPoint.Time
			NewEmitMsg <- *totalReceived.DataPoint

			totalUpdated.DataPoint.Time = time.Now().UTC()
			totalUpdated.LastEmitted = totalUpdated.DataPoint.Time
			NewEmitMsg <- *totalUpdated.DataPoint
		}
	}
}

func calculateDifference(entry *types.VolatileDataPoint, value float64) float64 {
	if value > entry.DataPointMeta.MaxValue {
		db.Trace("Meta data adjusted", "Adjusting MAX value for item: %s, from: %f to %f", entry.DataPoint.Name, entry.DataPointMeta.MaxValue, value)
		entry.DataPointMeta.MaxValue = value
	}

	if value < entry.DataPointMeta.MinValue {
		db.Trace("Meta data adjusted", "Adjusting MIN value for item: %s, from: %f to %f", entry.DataPoint.Name, entry.DataPointMeta.MinValue, value)
		entry.DataPointMeta.MinValue = value
	}

	dividend := entry.DataPointMeta.MaxValue - entry.DataPointMeta.MinValue
	if dividend == 0.0 {
		dividend = 0.0001
	}

	return dividend
}

func runMetaDispatch() {
	for {
		msg := <-NewMeta // DataPointMeta

		NotifySubscribers("meta.message", msg)

		// Update database with new items
		for _, msgitem := range msg {
			item := &types.DataPointMeta{Name: msgitem.Name, ID: uint(msgitem.ID), IntegratingDeadband: 0.3}
			if err := db.DB.First(&item).Error; err != nil {
				// fmt.Println("Tag not found in database, creating:", msgitem.Name)
				db.DB.Create(&item)
			}

			if _, ok := datapoints[msgitem.Name]; !ok {
				datapoints[msgitem.Name] = &types.VolatileDataPoint{}
				// fmt.Println("MAP now has", len(datapoints), "entries")
			}

			// Update emitters with new meta
			NewEmitMetaMsg <- *msgitem
		}
	}
}

func createDataPointEntry(dp *types.DataPoint) *types.VolatileDataPoint {
	metaitem := &types.DataPointMeta{Name: dp.Name, IntegratingDeadband: 0.3, MaxValue: 100.0}
	db.DB.Create(&metaitem)
	entry := &types.VolatileDataPoint{DataPoint: dp, DataPointMeta: *metaitem}
	datapoints[dp.Name] = entry
	return entry
}
