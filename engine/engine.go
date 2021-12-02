package engine

import (
	"log"
	"math"
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

func GetGroups() ([]*interface{}, error) {
	return nil, nil
}

func UpdateDataPointMeta(item *types.DataPointMeta) (err error) {
	return
}

func GetDataPoints() ([]types.VolatileDataPoint, error) {
	list := make([]types.VolatileDataPoint, len(datapoints))
	i := 0
	for _, item := range datapoints {
		list[i] = *item
		i++
		log.Println("item:", item)
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
			vp := &types.VolatileDataPoint{
				UpdateType:          item.UpdateType,
				Interval:            item.Interval,
				IntegratingDeadband: item.IntegratingDeadband}

			datapoints[item.Name] = vp
		} else {
			datapoints[item.Name].UpdateType = item.UpdateType
			datapoints[item.Name].Interval = item.Interval
			datapoints[item.Name].IntegratingDeadband = item.IntegratingDeadband
		}
	}
}

func runDataDispatch() {
	for {
		msg := <-NewMsg

		// Update totalRecevied
		totalReceived.DataPoint.Value = totalReceived.DataPoint.Value.(uint64) + uint64(msg.Count)
		if time.Now().UTC().Sub(totalReceived.DataPoint.Time) > time.Second {
			totalReceived.DataPoint.Time = time.Now().UTC()
			totalReceived.LastEmitted = totalReceived.DataPoint.Time
			NewEmitMsg <- *totalReceived.DataPoint
		}

		// Handle sequence number montioring and alerting
		gi, ok := groupseq[msg.Group]
		if !ok {
			gi = &groupSequenceInfo{groupName: msg.Group}
			gi.sequenceNumber.DataPoint = &types.DataPoint{Name: "Sequence Number: " + gi.groupName, Value: uint64(0)}
			datapoints[gi.sequenceNumber.DataPoint.Name] = &gi.sequenceNumber
			groupseq[msg.Group] = gi
		}

		diff := math.Abs(float64(msg.Sequence) - float64(gi.sequenceNumber.DataPoint.Value.(uint64)))
		if diff > 1.0 && gi.sequenceNumber.DataPoint.Value.(uint64) > 0 {
			db.Error("Engine data dispatch", "Sequence number out of sync, received %d, had %d, difference: %f",
				msg.Sequence, gi.sequenceNumber.DataPoint.Value, diff)
			SendAlerts()
		}

		// Update sequenceNumber
		gi.sequenceNumber.DataPoint.Value = msg.Sequence
		gi.sequenceNumber.DataPoint.Time = time.Now().UTC()
		gi.sequenceNumber.LastEmitted = gi.sequenceNumber.DataPoint.Time
		NewEmitMsg <- *gi.sequenceNumber.DataPoint

		// Update internal data point table
		dpLock.Lock()
		for _, dp := range msg.Points {

			entry, ok := datapoints[dp.Name]
			if !ok {
				entry = createDataPointEntry(&dp)
			}

			entry.DataPoint = &dp

			switch updateType := entry.UpdateType; updateType {
			case types.UpdateTypePassthru:
				entry.LastEmitted = time.Now().UTC()
				NewEmitMsg <- dp
				totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(uint64) + 1

			case types.UpdateTypeDeadband:
				// Check deadband
				if _, ok := dp.Value.(float64); ok {
					value := dp.Value.(float64)
					entry.Integrator += (entry.StoredValue - value)
					if math.Abs(entry.Integrator/value) > entry.IntegratingDeadband {
						entry.StoredValue = value
						entry.Integrator = 0.0
						entry.LastEmitted = time.Now().UTC()
						NewEmitMsg <- dp
						totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(uint64) + 1
					}
				}

			case types.UpdateTypeInterval:
				if time.Since(entry.LastEmitted) > time.Duration(entry.Interval)*time.Second {
					entry.LastEmitted = time.Now().UTC()
					NewEmitMsg <- dp
					totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(uint64) + 1
				}
			}
		}

		totalUpdated.DataPoint.Time = time.Now().UTC()
		totalUpdated.LastEmitted = totalUpdated.DataPoint.Time
		NewEmitMsg <- *totalUpdated.DataPoint

		dpLock.Unlock()
	}
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
	var lastitem types.DataPointMeta
	db.DB.Order("id desc").Last(&lastitem)
	item := &types.DataPointMeta{Name: dp.Name, IntegratingDeadband: 0.3, ID: lastitem.ID + 1}
	entry := &types.VolatileDataPoint{DataPoint: dp}
	datapoints[dp.Name] = entry
	db.DB.Create(item)
	return entry
}
