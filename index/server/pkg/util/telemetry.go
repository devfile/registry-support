package util

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"net"
)

const (
	telemetryKey = "6HBMiy5UxBtsbxXx7O4n0t0u4dt8IAR3"
	defaultUser  = "anonymous"
)

//TrackEvent tracks event for telemetry
func TrackEvent(event analytics.Message) error {
	// Initialize client for telemetry
	client := analytics.New(telemetryKey)
	defer client.Close()

	err := client.Enqueue(event)
	if err != nil {
		return err
	}
	return nil
}

//GetUser gets the user
func GetUser(c *gin.Context) string {
	user := defaultUser
	if len(c.Request.Header["User"]) != 0 {
		user = c.Request.Header["User"][0]
	}
	return user
}

//SetContext suppresses the collection of IP addresses in Segment but infers the country code from the HTTP `Accept-Language` header
func SetContext(c *gin.Context) *analytics.Context {
	aContext := analytics.Context{}
	aContext.IP = net.IPv4(0, 0, 0, 0)
	aContext.Location = analytics.LocationInfo{
		Country: getRegion(c),
	}
	return &aContext
}

//GetRegion returns the region that's set in Accept-Language header.  If the header is unset or can't be determined, an empty string will be returned.
func getRegion(c *gin.Context) string {
	userPrefs := c.Request.Header["Accept-Language"]
	defaultRegion := ""

	if len(userPrefs) != 0 {
		tags, _, err := language.ParseAcceptLanguage(userPrefs[0])
		if err != nil {
			log.Println(err.Error() + ", returning an empty string for region")
			return defaultRegion
		}

		if len(tags) > 0 {
			//tags are returned in order of precedence, so we can assume the first one is preferred
			region, _ := tags[0].Region()

			//if region is undetermined, return the default region
			if region.String() == "ZZ" {
				log.Println("Region is undetermined, returning an empty string for region")
				return defaultRegion
			}

			return region.String()
		}
	}

	//Accept-Language is not set, empty region
	log.Println("Accept-Language header is empty, returning an empty string for region")
	return defaultRegion

}
