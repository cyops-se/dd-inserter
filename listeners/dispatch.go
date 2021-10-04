package listeners

import (
	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
)

type IListener interface {
	InitListener()
}

var listeners []IListener

func RunDispatch() {
}

func Init() {
	var listeners []types.Listener

	db.DB.Find(&listeners)
	udpdata := &UDPDataListener{}
	udpdata.InitListener()

	udpmeta := &UDPMetaListener{}
	udpmeta.InitListener()
}
