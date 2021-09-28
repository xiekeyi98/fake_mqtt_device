package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/TylerBrock/colorjson"
	"github.com/pkg/errors"
)

func GetCurrentDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", errors.Cause(err)
	}
	return dir, nil
}

func GetMd5(s string) string {
	r := md5.Sum([]byte(s))
	return hex.EncodeToString(r[:])
}

var (
	colorjs = colorjson.NewFormatter()
)

func init() {
	colorjs.Indent = 2
}
func GetPrettyJSONStr(jsonStr string) string {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "{}"
	}
	return GetPrettyJSON(obj)
}

func GetPrettyJSON(jsonObj interface{}) string {
	s, err := colorjs.Marshal(jsonObj)
	if err != nil {
		return "{}"
	}
	return string(s)
}
