package zookeeper

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var once sync.Once
var zkc *zk.Conn

func InitConn(target string) (*zk.Conn, error) {

	once.Do(func() {

		addrs := strings.Split(target, ",")
		zkConn, events, err := zk.Connect(addrs, time.Second*5)
		if err != nil {

			log.Println("zk connect err " + err.Error())

			return
		}

		for {

			isConnected := false
			select {

			case connEvent := <-events:

				if connEvent.State == zk.StateConnected {

					isConnected = true
					log.Println("connect to zookeeper server success!")
				}

			case <-time.After(time.Second * 5):

				log.Println("connect to zookeeper server timeout!")
			}

			if isConnected {

				break
			}
		}

		zkc = zkConn
	})

	if zkc == nil {

		return nil, InitZkcErr
	}

	return zkc, nil
}

var InitZkcErr = errors.New("zkc init err")
