package util

import (
	"fmt"
)

func GetRid(ver, app_id, hash string) string {
	magic := "Y0o1OlO0*XT0811"
	rid := fmt.Sprintf("%s%x", ver, MMHash([]byte(app_id+"-"+hash)))
	hashStr := []byte(magic + rid + magic)
	suffix := Sha1hex(hashStr)[0:3]
	return rid + "." + suffix
}

func HashToRid(app_id, hash string) string {
	return GetRid("R0", app_id, hash)
}

func IdToRid(app_id, id string) string {
	return GetRid("R1", app_id, id)
}
