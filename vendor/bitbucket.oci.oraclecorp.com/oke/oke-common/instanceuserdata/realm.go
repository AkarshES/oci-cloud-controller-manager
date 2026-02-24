package instanceuserdata

import (
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

var realm string = "oc1"
var realmOnce sync.Once

// GetRealm will return realm for the
// current running environment.
// If for some reason the instancemetadata is failing we have bigger problems
// but the default will be oc1
func GetRealm() string {
	realmOnce.Do(func() {
		data, err := GetObject()
		if err != nil {
			log.WithError(err).Error("could not get instance metadata on startup")
			return
		}
		realm = getRealmString(data.CompartmentID)
	})

	return realm
}

func getRealmString(compartmentID string) string {
	if compartmentID == "" {
		return compartmentID
	}

	parts := strings.Split(strings.ToLower(compartmentID), ".")
	return parts[2]
}
