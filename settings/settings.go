package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {
	js, _ := getJson()
	if js == nil {
		js = map[string]interface{}{}
		js["TGBotApi"] = ""
		js["HostTG"] = "http://127.0.0.1:8081"
		js["HostTS"] = "http://127.0.0.1:8090"
		js["ContentDir"] = "/tmp"

		dir := filepath.Dir(os.Args[0])
		buf, err := json.MarshalIndent(js, "", " ")
		if err == nil {
			ioutil.WriteFile(filepath.Join(dir, "gv.cfg"), buf, 0666)
		}
	}
}

func GetTGBotApi() string {
	return get("TGBotApi", "")
}

func GetTGHost() string {
	return get("HostTG", "http://127.0.0.1:8081")
}

func GetTSHost() string {
	return get("HostTS", "http://127.0.0.1:8090")
}

func GetDownloadDir() string {
	return get("ContentDir", "/tmp")
}

func get[T any](name string, def T) T {
	js, err := getJson()
	if err != nil {
		return def
	}
	if v, ok := js[name]; !ok {
		return def
	} else {
		return v.(T)
	}
}

func getJson() (map[string]interface{}, error) {
	dir := filepath.Dir(os.Args[0])
	buf, err := os.ReadFile(filepath.Join(dir, "gv.cfg"))
	if err != nil {
		return nil, err
	}
	var js map[string]interface{}
	err = json.Unmarshal(buf, &js)
	return js, err
}
