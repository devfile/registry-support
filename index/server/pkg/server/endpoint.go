package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/devfile/registry-support/index/server/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/segmentio/analytics-go.v3"
)

// serveRootEndpoint sets up the handler for the root (/) endpoint on the server
// If html is requested (i.e. from a web browser), the viewer is displayed, otherwise the devfile index is served.
func serveRootEndpoint(c *gin.Context) {
	// Determine if text/html was requested by the client
	acceptHeader := c.Request.Header.Values("Accept")
	if util.IsHtmlRequested(acceptHeader) {
		c.Redirect(http.StatusFound, "/viewer")
	} else {
		serveDevfileIndex(c)
	}
}

// serveDevfileIndex serves the index.json file located in the container at `serveDevfileIndex`
func serveDevfileIndex(c *gin.Context) {
	// Start the counter for the request
	var status string
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		getIndexLatency.WithLabelValues(status).Observe(v)
	}))
	defer func() {
		timer.ObserveDuration()
	}()

	// append the devfile type, for endpoint /index without type
	c.Params = append(c.Params, gin.Param{Key: "type", Value: string(indexSchema.StackDevfileType)})

	// Serve the index.json file
	buildIndexAPIResponse(c)
}

// serveDevfileIndexWithType returns the index file content with specific devfile type
func serveDevfileIndexWithType(c *gin.Context) {

	// Serve the index with type
	buildIndexAPIResponse(c)
}

// serveHealthCheck serves endpoint `/health` for registry health check
func serveHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "the server is up and running",
	})
}

// serveDevfile returns the devfile content
func serveDevfile(c *gin.Context) {
	name := c.Param("name")

	var index []indexSchema.Schema
	bytes, err := ioutil.ReadFile(indexPath)
	if err != nil {
		log.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"status": fmt.Sprintf("failed to pull the devfile of %s", name),
		})
		return
	}
	err = json.Unmarshal(bytes, &index)
	if err != nil {
		log.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"status": fmt.Sprintf("failed to pull the devfile of %s", name),
		})
		return
	}
	for _, devfileIndex := range index {
		if devfileIndex.Name == name {
			var bytes []byte
			if devfileIndex.Type == indexSchema.StackDevfileType {
				bytes, err = pullStackFromRegistry(devfileIndex)
			} else {
				// Retrieve the sample devfile stored under /registry/samples/<devfile>
				sampleDevfilePath := path.Join(samplesPath, devfileIndex.Name, devfileName)
				if _, err = os.Stat(sampleDevfilePath); err == nil {
					bytes, err = ioutil.ReadFile(sampleDevfilePath)
				}
			}
			if err != nil {
				log.Print(err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  err.Error(),
					"status": fmt.Sprintf("failed to pull the devfile of %s", name),
				})
				return
			}

			// Track event for telemetry
			if enableTelemetry {
				user := util.GetUser(c)
				client := util.GetClient(c)

				err := util.TrackEvent(analytics.Track{
					Event:   eventTrackMap["view"],
					UserId:  user,
					Context: util.SetContext(c),
					Properties: analytics.NewProperties().
						Set("name", name).
						Set("type", string(devfileIndex.Type)).
						Set("registry", registry).
						Set("client", client),
				})
				if err != nil {
					log.Println(err)
				}
			}
			c.Data(http.StatusOK, http.DetectContentType(bytes), bytes)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"status": fmt.Sprintf("the devfile of %s didn't exist", name),
	})
}

func serveUI(c *gin.Context) {
	remote, err := url.Parse(scheme + "://" + viewerService + "/viewer/")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Set up the request to the proxy
	// This is a good place to set up telemetry for requests to the OCI server (e.g. by parsing the path)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", remote.Host)
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// buildIndexAPIResponse builds the response of the REST API of getting the devfile index
func buildIndexAPIResponse(c *gin.Context) {

	indexType := c.Param("type")
	iconType := c.Query("icon")
	archs := c.QueryArray("arch")

	var bytes []byte
	var responseIndexPath, responseBase64IndexPath string
	isFiltered := false

	// Load the appropriate index file name based on the devfile type
	switch indexType {
	case string(indexSchema.StackDevfileType):
		responseIndexPath = stackIndexPath
		responseBase64IndexPath = stackBase64IndexPath
	case string(indexSchema.SampleDevfileType):
		responseIndexPath = sampleIndexPath
		responseBase64IndexPath = sampleBase64IndexPath
	case "all":
		responseIndexPath = indexPath
		responseBase64IndexPath = base64IndexPath
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"status": fmt.Sprintf("the devfile with %s type doesn't exist", indexType),
		})
		return
	}

	// cache index with the encoded icon if required and save the encoded index location
	if iconType != "" {
		if iconType == encodeFormat {
			if _, err := os.Stat(responseBase64IndexPath); os.IsNotExist(err) {
				_, err := util.EncodeIndexIconToBase64(responseIndexPath, responseBase64IndexPath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"status": fmt.Sprintf("failed to encode %s icons to base64 format: %v", indexType, err),
					})
					return
				}
			}

			responseIndexPath = responseBase64IndexPath
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": fmt.Sprintf("the icon type %s is not supported", iconType),
			})
			return
		}
	}

	// Filter the index if archs has been requested
	if len(archs) > 0 {
		isFiltered = true
		index, err := util.ReadIndexPath(responseIndexPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("failed to read the devfile index: %v", err),
			})
			return
		}
		index = util.FilterDevfileArchitectures(index, archs)

		bytes, err = json.MarshalIndent(&index, "", "  ")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("failed to serialize index data: %v", err),
			})
			return
		}
	}

	// serve either the filtered index or the unfiltered index
	if isFiltered {
		c.Data(http.StatusOK, http.DetectContentType(bytes), bytes)
	} else {
		c.File(responseIndexPath)
	}

	// Track event for telemetry
	if enableTelemetry {
		user := util.GetUser(c)
		client := util.GetClient(c)
		err := util.TrackEvent(analytics.Track{
			Event:   eventTrackMap["list"],
			UserId:  user,
			Context: util.SetContext(c),
			Properties: analytics.NewProperties().
				Set("type", indexType).
				Set("registry", registry).
				Set("client", client),
		})
		if err != nil {
			log.Println(err)
		}
	}
}
