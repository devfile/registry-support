//
// Copyright 2022-2023 Red Hat, Inc.
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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/v2/pkg/util"
	libutil "github.com/devfile/registry-support/index/generator/library"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/devfile/registry-support/index/server/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/segmentio/analytics-go.v3"
)

type Server struct {
}

// ServeRootEndpoint sets up the handler for the root (/) endpoint on the server
// If html is requested (i.e. from a web browser), the viewer is displayed, otherwise the devfile index is served.
func (*Server) ServeRootEndpoint(c *gin.Context) {
	// Determine if text/html was requested by the client
	acceptHeader := c.Request.Header.Values("Accept")
	if util.IsHtmlRequested(acceptHeader) {
		c.Redirect(http.StatusFound, "/viewer")
	} else {
		c.Redirect(http.StatusFound, "/index")
	}
}

func (*Server) ServeDevfileIndexV1(c *gin.Context, params ServeDevfileIndexV1Params) {
	ServeDevfileIndex(c, true, IndexParams(params))
}

func (*Server) ServeDevfileIndexV2(c *gin.Context, params ServeDevfileIndexV2Params) {
	ServeDevfileIndex(c, false, IndexParams(params))
}

// ServeDevfileIndex serves the index.json file located in the container at `ServeDevfileIndex`
func ServeDevfileIndex(c *gin.Context, wantV1Index bool, params IndexParams) {
	// Start the counter for the request
	var status string
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		getIndexLatency.WithLabelValues(status).Observe(v)
	}))
	defer func() {
		timer.ObserveDuration()
	}()

	// Serve the index.json file
	buildIndexAPIResponse(c, string(indexSchema.StackDevfileType), wantV1Index, params)
}

func (*Server) ServeDevfileIndexV1WithType(c *gin.Context, indexType string, params ServeDevfileIndexV1WithTypeParams) {

	// Serve the index with type
	buildIndexAPIResponse(c, indexType, true, IndexParams(params))
}

func (*Server) ServeDevfileIndexV2WithType(c *gin.Context, indexType string, params ServeDevfileIndexV2WithTypeParams) {

	// Serve the index with type
	buildIndexAPIResponse(c, indexType, false, IndexParams(params))
}

// ServeHealthCheck serves endpoint `/health` for registry health check
func (*Server) ServeHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Message: "the server is up and running",
	})
}

func (*Server) ServeDevfileWithVersion(c *gin.Context, name string, version string) {
	bytes, devfileIndex := fetchDevfile(c, name, version)

	if len(bytes) != 0 {
		// Track event for telemetry.  Ignore events from the registry-viewer and DevConsole since those are tracked on the client side.  Ignore indirect calls from clients.
		if enableTelemetry && !util.IsWebClient(c) && !util.IsIndirectCall(c) {

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
	}
}

// ServeDevfile returns the devfile content
func (s *Server) ServeDevfile(c *gin.Context, name string) {
	// append the stack version, for endpoint /devfiles/name without version
	s.ServeDevfileWithVersion(c, name, "default")
}

// ServeDevfileStarterProject returns the starter project content for the devfile using default version
func (s *Server) ServeDevfileStarterProject(c *gin.Context, name string, starterProject string) {
	s.ServeDevfileStarterProjectWithVersion(c, name, "default", starterProject)
}

// ServeDevfileStarterProject returns the starter project content for the devfile using specified version
func (*Server) ServeDevfileStarterProjectWithVersion(c *gin.Context, name string, version string, starterProject string) {
	downloadTmpLoc := path.Join("/tmp", starterProject)
	stackLoc := path.Join(stacksPath, name)
	devfileBytes, devfileIndex := fetchDevfile(c, name, version)

	if len(devfileIndex.Versions) > 1 {
		versionMap, err := util.MakeVersionMap(devfileIndex)
		if err != nil {
			log.Print(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  err.Error(),
				"status": "failed to parse the stack version",
			})
			return
		}

		stackLoc = path.Join(stackLoc, versionMap[version].Version)
	}

	if len(devfileBytes) == 0 {
		// fetchDevfile was unsuccessful (error or not found)
		return
	} else {
		content, err := parser.ParseFromData(devfileBytes)
		filterOptions := common.DevfileOptions{
			FilterByName: starterProject,
		}
		var starterProjects []v1alpha2.StarterProject
		var downloadBytes []byte

		if err != nil {
			log.Print(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  err.Error(),
				"status": fmt.Sprintf("failed to parse the devfile of %s", name),
			})
			return
		}

		starterProjects, err = content.Data.GetStarterProjects(filterOptions)
		if err != nil {
			log.Print(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  err.Error(),
				"status": fmt.Sprintf("problem in reading starter project %s of devfile %s", starterProject, name),
			})
			return
		} else if len(starterProjects) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"status": fmt.Sprintf("the starter project named %s does not exist in the %s devfile", starterProject, name),
			})
			return
		}

		if selStarterProject := starterProjects[0]; selStarterProject.Git != nil {
			gitScheme := indexSchema.Git{
				Remotes:    selStarterProject.Git.Remotes,
				RemoteName: "origin",
				SubDir:     selStarterProject.SubDir,
			}

			if selStarterProject.Git.CheckoutFrom != nil {
				if selStarterProject.Git.CheckoutFrom.Remote != "" {
					gitScheme.RemoteName = selStarterProject.Git.CheckoutFrom.Remote
				}
				gitScheme.Revision = selStarterProject.Git.CheckoutFrom.Revision
			}

			gitScheme.Url = gitScheme.Remotes[gitScheme.RemoteName]

			if downloadBytes, err = libutil.DownloadStackFromGit(&gitScheme, downloadTmpLoc, false); err != nil {
				log.Print(err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
					"status": fmt.Sprintf("Problem with downloading starter project %s from location: %s",
						starterProject, gitScheme.Url),
				})
				return
			}
		} else if selStarterProject.Zip != nil {
			if _, err = url.ParseRequestURI(selStarterProject.Zip.Location); err != nil {
				localLoc := path.Join(stackLoc, selStarterProject.Zip.Location)
				log.Printf("zip location is not a valid http url: %v\nTrying local path %s..", err, localLoc)

				// If subdirectory is specified for starter project download then extract subdirectory
				// and create new archive for download.
				if selStarterProject.SubDir != "" {
					downloadFilePath := fmt.Sprintf("%s.zip", downloadTmpLoc)

					if _, err = os.Stat(downloadTmpLoc); os.IsExist(err) {
						err = os.Remove(downloadTmpLoc)
						if err != nil {
							log.Print(err.Error())
							c.JSON(http.StatusInternalServerError, gin.H{
								"error": err.Error(),
								"status": fmt.Sprintf("Problem removing existing temporary download directory '%s' for starter project %s",
									downloadTmpLoc,
									starterProject),
							})
							return
						}
					}

					_, err = dfutil.Unzip(localLoc, downloadTmpLoc, selStarterProject.SubDir)
					if err != nil {
						log.Print(err.Error())
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": err.Error(),
							"status": fmt.Sprintf("Problem with reading subDir '%s' of starter project %s at %s",
								selStarterProject.SubDir,
								starterProject,
								localLoc),
						})
						return
					}

					err = libutil.ZipDir(downloadTmpLoc, downloadFilePath)
					if err != nil {
						log.Print(err.Error())
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": err.Error(),
							"status": fmt.Sprintf("Problem with archiving subDir '%s' of starter project %s at %s",
								selStarterProject.SubDir,
								starterProject,
								downloadFilePath),
						})
						return
					}

					localLoc = downloadFilePath
				}

				downloadBytes, err = ioutil.ReadFile(filepath.Clean(localLoc))
				if err != nil {
					log.Print(err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
						"status": fmt.Sprintf("Problem with reading starter project %s at %s", starterProject,
							localLoc),
					})
					return
				}
			} else {
				downloadBytes, err = libutil.DownloadStackFromZipUrl(selStarterProject.Zip.Location, selStarterProject.SubDir, downloadTmpLoc)
				if err != nil {
					log.Print(err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":  err.Error(),
						"status": fmt.Sprintf("Problem with downloading starter project %s", starterProject),
					})
					return
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": fmt.Sprintf("Starter project %s has no source to download from", starterProject),
			})
			return
		}

		// Track event for telemetry. Ignore events from the registry-viewer and DevConsole since those are tracked on the client side. Ignore indirect calls from clients.
		if enableTelemetry && !util.IsWebClient(c) && !util.IsIndirectCall(c) {

			user := util.GetUser(c)
			client := util.GetClient(c)

			err := util.TrackEvent(analytics.Track{
				Event:   eventTrackMap["spdownload"],
				UserId:  user,
				Context: util.SetContext(c),
				Properties: analytics.NewProperties().
					Set("devfile", devfileName).
					Set("starterProject", starterProject).
					Set("client", client),
			})
			if err != nil {
				log.Println(err)
			}
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", starterProject))
		c.Data(http.StatusAccepted, starterProjectMediaType, downloadBytes)
	}
}

// ServeUIRoot handles registry viewer proxy request to root
func (s *Server) ServeUIRoot(c *gin.Context) {
	s.ServeUI(c, "")
}

// ServeUI handles registry viewer proxy requests
func (*Server) ServeUI(c *gin.Context, proxyPath string) {
	if headless {
		c.String(http.StatusBadRequest, "registry viewer is not available in headless mode")
		return
	}

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

	// Setup registry viewer proxy error response
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		buildProxyErrorResponse(w, r, err, "registry viewer")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// buildIndexAPIResponse builds the response of the REST API of getting the devfile index
func buildIndexAPIResponse(c *gin.Context, indexType string, wantV1Index bool, params IndexParams) {

	iconType := ""
	archs := []string{}

	if params.IconType != nil {
		iconType = *params.IconType
	}

	if params.Archs != nil {
		archs = append(archs, *params.Archs...)
	}

	var bytes []byte
	var responseIndexPath, responseBase64IndexPath string

	// Sets Access-Control-Allow-Origin response header to allow cross origin requests
	c.Header("Access-Control-Allow-Origin", "*")

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
	index, err := util.ReadIndexPath(responseIndexPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("failed to read the devfile index: %v", err),
		})
		return
	}
	if wantV1Index {
		index = util.ConvertToOldIndexFormat(index)
	} else {
		minSchemaVersion := c.Query("minSchemaVersion")
		maxSchemaVersion := c.Query("maxSchemaVersion")
		if maxSchemaVersion != "" || minSchemaVersion != "" {
			// check if schema version filters are in valid format.
			// should include major and minor versions as well as an optional bugfix version. e.g. 2.1 or 2.1.0
			if minSchemaVersion != "" {
				matched, err := regexp.MatchString(`^([2-9])\.([0-9]+)(\.[0-9]+)?$`, minSchemaVersion)
				if !matched || err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"status": fmt.Sprintf("minSchemaVersion %s is not valid, version format should be '+2.x' or '+2.x.x'. %v", minSchemaVersion, err),
					})
					return
				}
			}
			if maxSchemaVersion != "" {
				matched, err := regexp.MatchString(`^([2-9])\.([0-9]+)(\.[0-9]+)?$`, maxSchemaVersion)
				if !matched || err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"status": fmt.Sprintf("maxSchemaVersion %s is not valid, version format should be '+2.x' or '+2.x.x'. %v", maxSchemaVersion, err),
					})
					return
				}
			}

			index, err = util.FilterDevfileSchemaVersion(index, minSchemaVersion, maxSchemaVersion)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status": fmt.Sprintf("failed to apply schema version filter: %v", err),
				})
				return
			}
		}
	}
	// Filter the index if archs has been requested
	if len(archs) > 0 {
		index = util.FilterDevfileArchitectures(index, archs, wantV1Index)
	}
	bytes, err = json.MarshalIndent(&index, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("failed to serialize index data: %v", err),
		})
		return
	}
	c.Data(http.StatusOK, http.DetectContentType(bytes), bytes)

	// Track event for telemetry.  Ignore events from the registry-viewer and DevConsole since those are tracked on the client side
	if enableTelemetry && !util.IsWebClient(c) && !util.IsIndirectCall(c) {
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

// buildProxyErrorResponse builds an error response for proxy routes
func buildProxyErrorResponse(w http.ResponseWriter, r *http.Request, err error, name string) {
	var writeErr error

	log.Print(err.Error())

	if strings.Contains(err.Error(), "connection refused") {
		w.WriteHeader(http.StatusBadGateway)
		_, writeErr = w.Write([]byte(fmt.Sprintf("%s is not accessible", name)))

	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr = w.Write([]byte("internal server error"))
	}

	if writeErr != nil {
		log.Print(writeErr.Error())
	}
}

// fetchDevfile retrieves a specified devfile by fetching stacks from the OCI
// registry and samples from the `samplesPath` given by server. Also retrieves index
// schema from `indexPath` given by server.
func fetchDevfile(c *gin.Context, name string, version string) ([]byte, indexSchema.Schema) {
	var index []indexSchema.Schema
	bytes, err := ioutil.ReadFile(indexPath)
	if err != nil {
		log.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"status": fmt.Sprintf("failed to pull the devfile of %s", name),
		})
		return []byte{}, indexSchema.Schema{}
	}
	err = json.Unmarshal(bytes, &index)
	if err != nil {
		log.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"status": fmt.Sprintf("failed to pull the devfile of %s", name),
		})
		return []byte{}, indexSchema.Schema{}
	}

	// minSchemaVersion and maxSchemaVersion will only be applied if looking for latest stack version
	if version == "latest" {
		minSchemaVersion := c.Query("minSchemaVersion")
		maxSchemaVersion := c.Query("maxSchemaVersion")
		// check if schema version filters are in valid format.
		// should include major and minor versions as well as an optional bugfix version. e.g. 2.1 or 2.1.0
		if minSchemaVersion != "" {
			matched, err := regexp.MatchString(`^([2-9])\.([0-9]+)(\.[0-9]+)?$`, minSchemaVersion)
			if !matched || err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": fmt.Sprintf("minSchemaVersion %s is not valid, version format should be '+2.x' or '+2.x.x'. %v", minSchemaVersion, err),
				})
				return []byte{}, indexSchema.Schema{}
			}
		}
		if maxSchemaVersion != "" {
			matched, err := regexp.MatchString(`^([2-9])\.([0-9]+)(\.[0-9]+)?$`, maxSchemaVersion)
			if !matched || err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status": fmt.Sprintf("maxSchemaVersion %s is not valid, version format should be '+2.x' or '+2.x.x'. %v", maxSchemaVersion, err),
				})
				return []byte{}, indexSchema.Schema{}
			}
		}

		index, err = util.FilterDevfileSchemaVersion(index, minSchemaVersion, maxSchemaVersion)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": fmt.Sprintf("failed to apply schema version filter: %v", err),
			})
			return []byte{}, indexSchema.Schema{}
		}
	}

	for _, devfileIndex := range index {
		if devfileIndex.Name == name {
			var sampleDevfilePath string
			var bytes []byte
			if devfileIndex.Versions == nil || len(devfileIndex.Versions) == 0 {
				if devfileIndex.Type == indexSchema.SampleDevfileType {
					sampleDevfilePath = path.Join(samplesPath, devfileIndex.Name, name)
				}
			} else {
				versionMap, err := util.MakeVersionMap(devfileIndex)
				if err != nil {
					log.Print(err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":  err.Error(),
						"status": "failed to parse the stack version",
					})
					return []byte{}, indexSchema.Schema{}
				}
				if foundVersion, ok := versionMap[version]; ok {
					if devfileIndex.Type == indexSchema.StackDevfileType {
						bytes, err = pullStackFromRegistry(foundVersion)
						if err != nil {
							log.Print(err.Error())
							c.JSON(http.StatusInternalServerError, gin.H{
								"error":  err.Error(),
								"status": fmt.Sprintf("Problem pulling version %s from OCI Registry", foundVersion.Version),
							})
							return []byte{}, indexSchema.Schema{}
						}
					} else {
						// Retrieve the sample devfile stored under /registry/samples/<devfile>
						sampleDevfilePath = path.Join(samplesPath, devfileIndex.Name, foundVersion.Version, name)
					}
				} else {
					c.JSON(http.StatusNotFound, gin.H{
						"status": fmt.Sprintf("version: %s not found in stack %s", version, name),
					})
					return []byte{}, indexSchema.Schema{}
				}
			}
			if sampleDevfilePath != "" {
				if _, err = os.Stat(sampleDevfilePath); err == nil {
					/* #nosec G304 -- sampleDevfilePath is constructed from path.Join which cleans the input paths */
					bytes, err = ioutil.ReadFile(sampleDevfilePath)
				}
				if err != nil {
					log.Print(err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":  err.Error(),
						"status": fmt.Sprintf("failed to pull the devfile of %s", name),
					})
					return []byte{}, indexSchema.Schema{}
				}
			}

			return bytes, devfileIndex
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"status": fmt.Sprintf("the devfile of %s didn't exist", name),
	})
	return []byte{}, indexSchema.Schema{}
}

func (*Server) ServeOciServerProxy(c *gin.Context, proxyPath string) {
	remote, err := url.Parse(scheme + "://" + registryService + "/v2")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Set up the request to the proxy
	// Track event for telemetry for GET requests only
	if enableTelemetry && c.Request.Method == http.MethodGet && proxyPath != "" {
		var name string
		var resource string
		parts := strings.Split(proxyPath, "/")
		// Check proxyPath (e.g. /devfile-catalog/java-quarkus/blobs/sha256:d913cab108c3bc1bd06ce61f1e0cdb6eea2222a7884378f7e656fa26249990b9)
		if len(parts) == 5 {
			name = parts[2]
			resource = parts[3]
		}

		//Ignore events from the registry-viewer and DevConsole since those are tracked on the client side.  Ignore indirect calls from clients.
		if resource == "blobs" && !util.IsWebClient(c) && !util.IsIndirectCall(c) {
			user := util.GetUser(c)
			client := util.GetClient(c)

			err := util.TrackEvent(analytics.Track{
				Event:   eventTrackMap["download"],
				UserId:  user,
				Context: util.SetContext(c),
				Properties: analytics.NewProperties().
					Set("name", name).
					Set("registry", registry).
					Set("client", client),
			})
			if err != nil {
				log.Println(err.Error())
			}
		}
	}

	proxy.Director = func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", remote.Host)
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (s *Server) GetOciServerProxy(c *gin.Context, proxyPath string) {
	s.ServeOciServerProxy(c, proxyPath)
}

func (s *Server) HeadOciServerProxy(c *gin.Context, proxyPath string) {
	s.ServeOciServerProxy(c, proxyPath)
}
