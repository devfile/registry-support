//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"net"
)

const (
	defaultUser = "devfile-registry"
	viewerId    = "registry-viewer"
	consoleId   = "openshift-console"
)

var telemetryKey = GetOptionalEnv("TELEMETRY_KEY", "").(string)

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
	user := GetClient(c)
	if len(c.Request.Header["User"]) != 0 {
		user = c.Request.Header["User"][0]
	}
	return user
}

func GetClient(c *gin.Context) string {
	client := defaultUser

	cHeader := c.Request.Header["Client"]
	if len(cHeader) != 0 {
		client = cHeader[0]
	}

	return client
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

// getRegion returns the region that's set in Accept-Language header if request is coming from a browser or the Locale header if the request is from a client.
// If the header is unset or can't be determined, an empty string will be returned.
func getRegion(c *gin.Context) string {
	userPrefs := c.Request.Header["Accept-Language"]
	defaultRegion := ""

	if len(userPrefs) == 0 {
		userPrefs = c.Request.Header["Locale"]
		if len(userPrefs) == 0 {
			log.Println("The Accept-Language or Locale headers are unset , returning an empty string for region")
			return defaultRegion
		}
	}

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
	} else {
		log.Println("Locale is unset, returning an empty string for region")
		return defaultRegion
	}

}

//IsWebClient determines if the event is coming from the registry viewer or DevConsole client.
func IsWebClient(c *gin.Context) bool {
	client := GetClient(c)
	userId := GetUser(c)
	if client == viewerId || userId == consoleId {
		return true
	}

	return false
}
