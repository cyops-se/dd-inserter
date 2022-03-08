package listeners

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/nats-io/nats.go"
)

type NATSFileListener struct {
	URL      string `json:"url"`
	nc       *nats.Conn
	ctx      *context
	n        int
	err      error
	h        header
	filename string
	file     *os.File
}

func (listener *NATSFileListener) InitListener(gctx *types.Context) (err error) {
	listener.nc, err = nats.Connect(nats.DefaultURL)
	if err == nil {
		listener.nc.Subscribe("file", listener.callbackHandler)
		log.Printf("NATS server connected")
	}

	listener.ctx = &context{basedir: path.Join(gctx.Wdir, "incoming"), processingdir: "processing", donedir: "done", faildir: "failed"}
	os.MkdirAll(path.Join(listener.ctx.basedir, listener.ctx.processingdir), 0755)
	os.MkdirAll(path.Join(listener.ctx.basedir, listener.ctx.donedir), 0755)
	os.MkdirAll(path.Join(listener.ctx.basedir, listener.ctx.faildir), 0755)

	return err
}

func (listener *NATSFileListener) callbackHandler(msg *nats.Msg) {
	text := string(msg.Data)
	if strings.HasPrefix(text, "DD-FILETRANSFER BEGIN") {
		listener.n, listener.err = fmt.Sscanf(text, "DD-FILETRANSFER BEGIN v2 %s %s %d %x",
			&listener.h.filename, &listener.h.directory, &listener.h.size, &listener.h.hashvalue)

		if listener.n != 4 || listener.err != nil {
			db.Trace("file transfer error", "Failed to read 4 items from header, got %d, error: %s", listener.n, listener.err.Error())
			log.Printf(text)
			return
		}

		listener.close()
		listener.filename = path.Join(listener.ctx.basedir, listener.ctx.processingdir, listener.h.filename)
		listener.file, listener.err = os.Create(listener.filename)
		if listener.err != nil {
			db.Trace("file transfer error", "Failed to create file %s, error: %s", listener.filename, listener.err.Error())
			return
		}

		db.Trace("file transfer start", listener.filename)
	} else if strings.HasPrefix(text, "DD-FILETRANSFER END") {
		db.Trace("file transfer end", listener.filename)
		listener.close()
	} else {
		if listener.file != nil && len(msg.Data) > 8 {
			listener.file.Write(msg.Data[8:])
		}
	}
}

func (listener *NATSFileListener) close() {
	if listener.file != nil {
		listener.file.Close()
		listener.file = nil
	}
}
