package gofunc

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

func init() {

	appPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	appPath = appPath + string(os.PathSeparator)
	appPath = filepath.Join(appPath + "..")

	dataConf = &gameDataConf{data: make(map[string]interface{})}
}

var appPath string
var dataConf *gameDataConf

const (
	CONFIGS = "configs"
	DATAS   = "datas"
)

// gameDataConf 游戏策划配置
type gameDataConf struct {
	mux  sync.RWMutex
	data map[string]interface{}
}

//LoadJsonConf 加载json静态配置数据
//model：configs or data
//name: 文件名
func LoadJsonConf(model, name string, v interface{}) {

	fileName := filepath.Join(appPath, model, name+".json")

	err := loadJsonFile(fileName, v)
	if err != nil {

		panic(err)
	}
}

//GetAppPath 获取应用路径
func GetAppPath() string {

	return appPath
}

func loadJsonFile(file string, v interface{}) error {

	buf, err := ioutil.ReadFile(file)
	if err != nil {

		return err
	}

	err = json.Unmarshal(buf, v)

	return err
}

// GetdataConf 数据
func GetDataConf(key string) interface{} {

	dataConf.mux.RLock()
	defer dataConf.mux.RUnlock()

	return dataConf.data[key]
}

// Set 数据
func SetDataConf(key string, v interface{}) {

	dataConf.mux.Lock()
	defer dataConf.mux.Unlock()

	dataConf.data[key] = v
}
