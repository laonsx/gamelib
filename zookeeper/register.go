package zookeeper

import (
	"log"

	"github.com/samuel/go-zookeeper/zk"
)

const schema = "gamelibzk"

var paths []string

func Register(target, server, value string) error {

	_, err := InitConn(target)
	if err != nil {

		return err
	}

	path := "/" + schema + "/" + server + "/" + value

	_, err = zkc.Create(path, nil, 0, zk.WorldACL(zk.PermAll))
	switch {
	case err == zk.ErrNoNode:

		_, _ = zkc.Create("/"+schema, nil, 0, zk.WorldACL(zk.PermAll))
		_, _ = zkc.Create("/"+schema+"/"+server, nil, 0, zk.WorldACL(zk.PermAll))
		_, err := zkc.Create(path, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {

			return err
		}

	case err == zk.ErrNodeExists:

	case err == nil:

	default:

		return err
	}

	log.Println("register =>", path)

	paths = append(paths, path)

	return nil
}

func UnRegister() {

	if zkc == nil {

		return
	}

	for _, path := range paths {

		err := zkc.Delete(path, -1)
		if err == nil {

			log.Println("unregister =>", path)
		}
	}
}
