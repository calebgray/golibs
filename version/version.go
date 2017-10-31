package version

import "time"

var _version = time.Now().Format("20060102150405")

func Get() string {
	return _version
}
