package main

import (
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	S_MOTION = 1
)

type MySensMsg struct {
	NodeID  int
	ChildID int
	Type    int
	Ack     int
	SubType int
	Data    string
}

func main() {
	logwriter, err := syslog.New(syslog.LOG_NOTICE, "risense")
	if err == nil {
		log.SetOutput(logwriter)
	} else {
		log.Print("Unable to connect to syslog", err)
	}

	file, err := os.Open("/dev/ttyUSB20")
	if err != nil {
		log.Fatal(err)
	}

	var mutexOff = &sync.Mutex{}
	offsig := 0

	for {
		data := make([]byte, 100)
		count, err := file.Read(data)

		if err != nil {
			log.Fatal(err)
		}

		// We blindly assume that we got ONE ENTIRE package here

		sData := strings.SplitN(string(data[:count]), ";", 6)

		var msg MySensMsg
		msg.NodeID, _ = strconv.Atoi(sData[0])
		msg.ChildID, _ = strconv.Atoi(sData[1])
		msg.Type, _ = strconv.Atoi(sData[2])
		msg.Ack, _ = strconv.Atoi(sData[3])
		msg.SubType, _ = strconv.Atoi(sData[4])
		msg.Data = sData[5][:len(sData[5])-1]
		log.Printf("%s", msg)

		if msg.NodeID == 1 && msg.Data == "1" {
			cmd := exec.Command("xset", "dpms", "force", "on")
			err := cmd.Start()
			if err != nil {
				log.Fatal(err)
			}
			cmd.Wait()
		} else if msg.NodeID == 1 && msg.Data == "0" {
			mutexOff.Lock()
			offsig++
			mutexOff.Unlock()

			go func() {
				time.Sleep(5 * time.Minute)

				turnoff := false
				mutexOff.Lock()
				offsig--
				if offsig == 0 {
					turnoff = true
				}
				mutexOff.Unlock()

				if turnoff {

					cmd := exec.Command("xset", "dpms", "force", "off")
					err := cmd.Start()
					if err != nil {
						log.Fatal(err)
					}
					cmd.Wait()
				}
			}()
		}

	}
}
