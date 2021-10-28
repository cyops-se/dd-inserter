package engine

import (
	"container/list"
	"log"
	"math"
	"sync"
	"time"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
)

var queueLock sync.Mutex
var dpLock sync.Mutex
var queue list.List
var datapoints map[string]*types.VolatileDataPoint = make(map[string]*types.VolatileDataPoint)
var NewMsg chan types.DataMessage = make(chan types.DataMessage)
var NewMeta chan []*types.DataPointMeta = make(chan []*types.DataPointMeta)
var NewEmitMsg chan types.DataPoint = make(chan types.DataPoint)
var NewEmitMetaMsg chan types.DataPointMeta = make(chan types.DataPointMeta)

// var totalReceived types.DataPoint
// var totalUpdated types.DataPoint
var totalReceived types.VolatileDataPoint
var totalUpdated types.VolatileDataPoint

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
	// totalReceived.Name = "Total Received"
	// totalReceived.Value = 0
	// vdp1 := &types.VolatileDataPoint{DataPoint: &totalReceived}
	// datapoints[totalReceived.Name] = vdp1

	// totalUpdated.Name = "Total Updated"
	// totalUpdated.Value = 0
	// vdp2 := &types.VolatileDataPoint{DataPoint: &totalUpdated}
	// datapoints[totalReceived.Name] = vdp2

	totalReceived.DataPoint = &types.DataPoint{Name: "Total Received", Value: 0}
	datapoints[totalReceived.DataPoint.Name] = &totalReceived

	totalUpdated.DataPoint = &types.DataPoint{Name: "Total Updated", Value: 0}
	datapoints[totalUpdated.DataPoint.Name] = &totalUpdated

	InitDataPointMap()
	go runDataDispatch()
	// go runMetaDispatch() // disable temporarily until fixed
}

func InitDataPointMap() {
	var items []types.DataPointMeta
	db.DB.Find(&items)
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

		// Update internal data point table
		dpLock.Lock()
		for _, dp := range msg.Points {
			totalReceived.DataPoint.Value = totalReceived.DataPoint.Value.(int) + 1
			totalReceived.DataPoint.Time = time.Now().UTC()
			NewEmitMsg <- *totalReceived.DataPoint

			entry, ok := datapoints[dp.Name]
			if !ok {
				entry = &types.VolatileDataPoint{}
				datapoints[dp.Name] = entry
			}

			entry.DataPoint = &dp

			switch updateType := entry.UpdateType; updateType {
			case types.UpdateTypePassthru:
				// log.Println("Emitting passthru", entry.DataPoint.Name)
				entry.LastEmitted = time.Now().UTC()
				NewEmitMsg <- dp
				totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(int) + 1
				totalUpdated.DataPoint.Time = entry.LastEmitted
				totalUpdated.LastEmitted = entry.LastEmitted
				NewEmitMsg <- *totalUpdated.DataPoint

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
						totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(int) + 1
						totalUpdated.DataPoint.Time = entry.LastEmitted
						totalUpdated.LastEmitted = entry.LastEmitted
						NewEmitMsg <- *totalUpdated.DataPoint
					}
				}

			case types.UpdateTypeInterval:
				if time.Since(entry.LastEmitted) > time.Duration(entry.Interval)*time.Second {
					// log.Println("Emitting interval", entry.DataPoint.Name)
					entry.LastEmitted = time.Now().UTC()
					NewEmitMsg <- dp
					totalUpdated.DataPoint.Value = totalUpdated.DataPoint.Value.(int) + 1
					totalUpdated.DataPoint.Time = entry.LastEmitted
					totalUpdated.LastEmitted = entry.LastEmitted
					NewEmitMsg <- *totalUpdated.DataPoint
				}
			}
		}

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
