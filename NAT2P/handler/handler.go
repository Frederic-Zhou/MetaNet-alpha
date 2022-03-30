package handler

import (
	"fmt"
	"time"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/network"
)

type HandleFunc func(*network.Event) error

var txt_handlers []HandleFunc
var file_handlers []HandleFunc
var flow_handlers []HandleFunc
var errLog []error

func EventHandleLoop() {

	for {
		e := network.GetEvent()
		fmt.Println(e)
		if e != nil {
			var hs []HandleFunc

			switch e.GetDataType() {
			case network.DataType_Text:
				hs = txt_handlers
			case network.DataType_File:
				hs = file_handlers
			case network.DataType_Flow:
				hs = flow_handlers
			}

			go func() {
				for _, h := range hs {
					err := h(e)
					if err != nil {
						errLog = append(errLog, err)
					}
				}
			}()

			continue
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func RegisterTxtHandler(h HandleFunc) {
	txt_handlers = append(txt_handlers, h)
}

func RegisterFileHandler(h HandleFunc) {
	file_handlers = append(file_handlers, h)
}
