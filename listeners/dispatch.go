package listeners

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
)

type ListenerType struct {
	Name string `json:"name"`
	Type reflect.Type
}

var typeRegistry = make(map[string]*ListenerType)
var Listeners []types.Listener

func RegisterType(name string, i interface{}) {
	typename := reflect.TypeOf(i).String()
	log.Println("Registering listener type:", name, typename)
	typeRegistry[name] = &ListenerType{Name: name, Type: reflect.TypeOf(i)}
}

func RunDispatch() {
}

// func Init(ctx types.Context) {
// 	var listeners []types.Listener

// 	db.DB.Find(&listeners)
// 	udpdata := &UDPDataListener{}
// 	udpdata.InitListener(&ctx)

// 	udpmeta := &UDPMetaListener{}
// 	udpmeta.InitListener(&ctx)

// 	udpfile := &UDPFileListener{}
// 	udpfile.InitListener(&ctx)

// 	cache := &CacherListener{}
// 	cache.InitListener(&ctx)
// }

func CreateListener(typename string) (*types.Listener, error) {
	if et := typeRegistry[typename]; et != nil {
		instance := reflect.New(et.Type).Interface().(types.IListener)
		listener := &types.Listener{Instance: instance}
		return listener, nil
	}

	return nil, fmt.Errorf("Typename not in registry: %s", typename)
}

func Load(listener *types.Listener) error {
	if et := typeRegistry[listener.Type]; et != nil {
		listener.Instance = reflect.New(et.Type).Interface().(types.IListener)
		return json.Unmarshal([]byte(listener.Settings), &listener.Instance)
	}

	return fmt.Errorf("Typename not in registry: %s", listener.Type)
}

func LoadListener(listener *types.Listener, ctx *types.Context) error {
	// Emitters = append(Emitters, *emitter)
	if err := Load(listener); err != nil {
		db.Log("error", "Failed to load emitter settings", err.Error())
		return err
	}

	if err := listener.Instance.InitListener(ctx); err != nil {
		db.Log("error", "Failed to load initialize emitter", err.Error())
		return err
	}

	return nil
}

func LoadListeners(ctx types.Context) {
	// Load dispatcher from database
	if err := db.DB.Find(&Listeners).Error; err != nil {
		db.Log("error", "Failed to find listeners in database", err.Error())
		return
	}

	for i, _ := range Listeners {
		LoadListener(&Listeners[i], &ctx)
	}
}

func GetTypeNames() []string {
	var names []string
	for _, item := range typeRegistry {
		names = append(names, item.Name)
	}
	return names
}
