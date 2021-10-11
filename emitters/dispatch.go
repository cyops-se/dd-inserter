package emitters

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
)

type EmitterType struct {
	Name string `json:"name"`
	Type reflect.Type
}

var typeRegistry = make(map[string]*EmitterType)
var Emitters []types.Emitter

func RunDispatch() {
	go messageDispatch()
	go metaDispatch()
}

func CreateEmitter(typename string) (*types.Emitter, error) {
	if et := typeRegistry[typename]; et != nil {
		instance := reflect.New(et.Type).Interface().(types.IEmitter)
		emitter := &types.Emitter{Instance: instance}
		return emitter, nil
	}

	return nil, fmt.Errorf("Typename not in registry: %s", typename)
}

func Load(emitter *types.Emitter) error {
	if et := typeRegistry[emitter.Type]; et != nil {
		emitter.Instance = reflect.New(et.Type).Interface().(types.IEmitter)
		return json.Unmarshal([]byte(emitter.Settings), &emitter.Instance)
	}

	return fmt.Errorf("Typename not in registry: %s", emitter.Type)
}

func LoadEmitter(emitter *types.Emitter) error {
	// Emitters = append(Emitters, *emitter)
	if err := Load(emitter); err != nil {
		db.Log("error", "Failed to load emitter settings", err.Error())
		return err
	}

	if err := emitter.Instance.InitEmitter(); err != nil {
		db.Log("error", "Failed to load initialize emitter", err.Error())
		return err
	}

	return nil
}

func LoadEmitters() {
	// Load dispatcher from database
	if err := db.DB.Find(&Emitters).Error; err != nil {
		db.Log("error", "Failed to find emitters in database", err.Error())
		return
	}

	for i, _ := range Emitters {
		LoadEmitter(&Emitters[i])
	}
}

func RegisterType(name string, i interface{}) {
	typename := reflect.TypeOf(i).String()
	log.Println("Registering emitter type:", name, typename)
	typeRegistry[name] = &EmitterType{Name: name, Type: reflect.TypeOf(i)}
}

func GetTypeNames() []string {
	var names []string
	for _, item := range typeRegistry {
		names = append(names, item.Name)
	}
	return names
}

func messageDispatch() {
	for {
		msg := <-engine.NewEmitMsg
		for _, emitter := range Emitters {
			if emitter.Instance != nil {
				emitter.Instance.ProcessMessage(&msg)
			}
		}

		// Always emit to websocket subscribers
		engine.NotifySubscribers("data.message", msg)
	}
}

func metaDispatch() {
	for {
		msg := <-engine.NewEmitMetaMsg
		for _, emitter := range Emitters {
			if emitter.Instance != nil {
				emitter.Instance.ProcessMeta(&msg)
			}
		}

		// Always emit to websocket subscribers
		engine.NotifySubscribers("data.meta", msg)
	}
}
