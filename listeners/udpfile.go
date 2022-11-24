package listeners

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cyops-se/dd-inserter/types"
)

type header struct {
	directory string
	filename  string
	size      int
	hashvalue []byte
}

type context struct {
	basedir       string
	processingdir string
	donedir       string
	faildir       string
}

type UDPFileListener struct {
	Port int `json:"port"`
}

func (listener *UDPFileListener) InitListener(gctx *types.Context) (err error) {
	ctx := initContext(gctx)
	go listenForIncoming(ctx)
	return err
}

func initContext(gctx *types.Context) *context {
	ctx := &context{basedir: "incoming", processingdir: "processing", donedir: "done", faildir: "failed"}
	os.MkdirAll(path.Join(gctx.Wdir, ctx.basedir, ctx.processingdir), 0755)
	os.MkdirAll(path.Join(gctx.Wdir, ctx.basedir, ctx.donedir), 0755)
	os.MkdirAll(path.Join(gctx.Wdir, ctx.basedir, ctx.faildir), 0755)
	return ctx
}

func listenForIncoming(ctx *context) {
	addr := net.UDPAddr{
		Port: 4358,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Failed to listen %v\n", err)
		return
	}

	var h header
	p := make([]byte, 1200)

	log.Println("UDP listening for FILE data header ...")

	for {
		// Resume blocking read
		var zero time.Time
		ser.SetReadDeadline(zero)
		n, _, err := ser.ReadFromUDP(p)
		if err != nil {
			log.Printf("Failed to read HEADER: %v", err)
			continue
		}

		data := string(p[:n])
		count, err := fmt.Sscanf(data, "DD-FILETRANSFER BEGIN v2 %s %s %d %x", &h.filename, &h.directory, &h.size, &h.hashvalue)
		if count != 4 || err != nil {
			if strings.HasPrefix(data, "DD-FILETRANSFER END v2") {
				continue
			}

			log.Printf("Failed to read 4 items from header, got %d, error: %s", count, err.Error())
			// log.Println("-----------\n" + data + "\n-----------\n")
			continue
		}

		// Got header, now receive the file
		log.Printf("BEGIN transfer file: %s, dir: %s, size: %d, hash: %x", h.filename, h.directory, h.size, h.hashvalue)
		receiveFile(ctx, h, ser)
	}
}

func receiveFile(ctx *context, h header, ser *net.UDPConn) {
	filename := path.Join(ctx.basedir, ctx.processingdir, h.filename)
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create file %s, error: %s", filename, err.Error())
		return
	}

	totalsize := 0
	count := 0
	previousCounter := uint32(0)
	totalmissing := uint32(0)

	p := make([]byte, 1200)

	start := time.Now()
	for totalsize < h.size {
		ser.SetReadDeadline(time.Now().Add(time.Millisecond * 1000))
		if n, _, err := ser.ReadFromUDP(p); err == nil {
			counter := binary.LittleEndian.Uint32(p)
			chunksize := binary.LittleEndian.Uint32(p[4:])
			// log.Printf("Chunk counter: %d, size: %d, bytes read: %d, content: %s", counter, chunksize, n, string(p[8:n-8]))

			missing := counter - (previousCounter + 1)
			totalmissing += missing
			if missing > 0 && !(counter == 0 && previousCounter == 0) {
				log.Printf("Missing packets (%d), got sequence %d, wanted %d", missing, counter, previousCounter+1)
				break
			}

			previousCounter = counter

			file.Write(p[8 : chunksize+8])
			// totalsize += (n - 8)
			totalsize += int(chunksize)
			count++

			if count%1000 == 0 {
				percent := (float64(totalsize) / float64(h.size)) * 100.0
				log.Printf("Progress: %.2f (%d), %d (%d), total packets: %d (%d)\n", percent, n, totalsize, h.size, count, counter)
			}
		} else {
			log.Printf("Error while reading from UDP stream for file %s, error: %s", filename, err.Error())
			break
		}
	}

	// There is a bug in dd-opcda in production that sends the last content package
	// twice. This piece will empty the UDP pipe after each file

	for {
		if trailingcount, _, _ := ser.ReadFromUDP(p); trailingcount > 0 {
			if strings.HasPrefix(string(p), "DD-FILETRANSFER END") {
				// log.Println("END OF FILE TRANSFER DETECTED")
				break
			}
		}
	}

	end := time.Now()

	percent := (float64(totalsize) / float64(h.size)) * 100.0
	log.Printf("END file transfer: %.2f, %d (%d), packets: %d, time: %d\n", percent, totalsize, h.size, count, end.Sub(start)/time.Second)

	file.Close()
	hash := calcHash(filename)
	hashvalue := hash.Sum(nil)
	result := bytes.Compare(hashvalue, h.hashvalue)
	if result == 0 {
		log.Printf("File received SUCCESSFULLY")
		todir := path.Join(ctx.basedir, ctx.donedir, h.directory)
		os.MkdirAll(todir, 0755)
		toname := path.Join(todir, h.filename)
		if err = os.Rename(filename, toname); err != nil {
			log.Printf("Failed to move file to done directory: %s", err.Error())
		}
	} else {
		log.Printf("Failed to receive file completely (%d packets missing), hash mismatch: %x != %x", totalmissing, hashvalue, h.hashvalue)
		todir := path.Join(ctx.basedir, ctx.faildir, h.directory)
		os.MkdirAll(todir, 0755)
		toname := path.Join(todir, h.filename)
		os.Rename(filename, toname)
	}
}

func calcHash(filename string) hash.Hash {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return h
}
