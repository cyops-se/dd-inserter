package listeners

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
)

type CacherListener struct {
	Port int `json:"port"`
}

func (listener *CacherListener) InitListener() {
	listeners = append(listeners, listener)
	go listener.run()
}

func (listener *CacherListener) run() {

	ticker := time.NewTicker(500 * time.Millisecond)

	for {
		<-ticker.C
		processDirectory()
	}
}

func processDirectory() {
	dir := path.Join("incoming", "done", "cache")
	if err := filepath.Walk(dir, indexer); err != nil {
		log.Println("FILEWALK ERROR:", err.Error())
	}
}

func indexer(filename string, info os.FileInfo, err error) error {
	if info == nil {
		return fmt.Errorf("file information nil, %s", filename)
	}

	if !info.IsDir() && filepath.Ext(filename) == ".gz" {
		file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
		if err != nil {
			os.Remove(filename)
			return fmt.Errorf("failed to open file, %s, error: %s", filename, err.Error())
		}

		gzr, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			file.Sync()
			os.Remove(filename)
			return fmt.Errorf("failed to open gzip stream, %s, error: %s", filename, err.Error())
		}

		fr := bufio.NewReader(gzr)

		data, _ := ioutil.ReadAll(fr)
		var msgs []types.DataMessage
		if err := json.Unmarshal(data, &msgs); err == nil {
			for _, msg := range msgs {
				engine.NewMsg <- msg
			}
		}

		gzr.Close()
		file.Close()

		os.Remove(filename)
	}

	return nil
}
