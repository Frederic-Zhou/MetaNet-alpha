package handler

import (
	"fmt"
	"time"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/network"
)

type HandleFunc func(*network.Event) error

var handlers []HandleFunc
var errLog []error

func EventHandleLoop() {

	for {
		e := network.GetEvent()
		fmt.Println(e)
		if e != nil {
			go func() {
				for _, h := range handlers {
					err := h(e)
					if err != nil {
						errLog = append(errLog, err)
					}
				}
			}()
		} else {
			time.Sleep(200 * time.Millisecond)
		}

	}

}

func AddHandlers(hf HandleFunc) {
	handlers = append(handlers, hf)
}
