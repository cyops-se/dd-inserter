package listeners

import (
	"log"
	"reflect"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
)

var typeRegistry = make(map[string]reflect.Type)
var listeners []types.IListener

func RegisterType(i types.IListener) {
	typename := reflect.TypeOf(i).String()
	log.Println("Registering emitter type:", typename, i)
	typeRegistry[typename] = reflect.TypeOf(i)
}

func RunDispatch() {
}

func Init() {
	var listeners []types.Listener

	db.DB.Find(&listeners)
	udpdata := &UDPDataListener{}
	udpdata.InitListener()

	udpmeta := &UDPMetaListener{}
	udpmeta.InitListener()

	cache := &CacherListener{}
	cache.InitListener()
}
