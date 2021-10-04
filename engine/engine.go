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

func GetProcessPoints() ([]*types.ProcessPoint, error) {
	var items []*types.ProcessPoint
	db.DB.Find(&items)
	return items, nil
}

func InitDispatchers() {
	InitDataPointMap()
	go runDataDispatch()
	go runMetaDispatch()
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
		// Update internal data point table
		dpLock.Lock()
		for _, dp := range msg.Points {
			if entry, ok := datapoints[dp.Name]; ok {
				entry.DataPoint = &dp

				switch updateType := entry.UpdateType; updateType {
				case types.UpdateTypePassthru:
					// log.Println("Emitting passthru", entry.DataPoint.Name)
					NewEmitMsg <- dp
					entry.LastEmitted = time.Now()
					break

				case types.UpdateTypeDeadband:
					// Check deadband
					if _, ok := dp.Value.(float64); ok {
						value := dp.Value.(float64)
						entry.Integrator += (entry.StoredValue - value)
						if math.Abs(entry.Integrator/value) > entry.IntegratingDeadband {
							entry.StoredValue = value
							entry.Integrator = 0.0
							entry.LastEmitted = time.Now()
							NewEmitMsg <- dp
						}
					}
					break

				case types.UpdateTypeInterval:
					if time.Now().Sub(entry.LastEmitted) > time.Duration(entry.Interval)*time.Second {
						// log.Println("Emitting interval", entry.DataPoint.Name)
						NewEmitMsg <- dp
						entry.LastEmitted = time.Now()
					}
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
