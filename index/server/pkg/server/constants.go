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

package server

import (
	"os"

	"github.com/devfile/registry-support/index/server/pkg/util"
)

const (
	// Constants for resource names and media types
	archiveMediaType        = "application/x-tar"
	archiveName             = "archive.tar"
	starterProjectMediaType = "application/zip"
	devfileName             = "devfile.yaml"
	devfileNameHidden       = ".devfile.yaml"
	devfileConfigMediaType  = "application/vnd.devfileio.devfile.config.v2+json"
	devfileMediaType        = "application/vnd.devfileio.devfile.layer.v1"
	pngLogoMediaType        = "image/png"
	pngLogoName             = "logo.png"
	svgLogoMediaType        = "image/svg+xml"
	svgLogoName             = "logo.svg"
	vsxMediaType            = "application/vnd.devfileio.vsx.layer.v1.tar"
	vsxName                 = "vsx"

	scheme          = "http"
	registryService = "localhost:5000"
	viewerService   = "localhost:3000"
	encodeFormat    = "base64"
)

var (
	stacksPath            = os.Getenv("DEVFILE_STACKS")
	samplesPath           = os.Getenv("DEVFILE_SAMPLES")
	indexPath             = os.Getenv("DEVFILE_INDEX")
	base64IndexPath       = os.Getenv("DEVFILE_BASE64_INDEX")
	sampleIndexPath       = os.Getenv("DEVFILE_SAMPLE_INDEX")
	sampleBase64IndexPath = os.Getenv("DEVFILE_SAMPLE_BASE64_INDEX")
	stackIndexPath        = os.Getenv("DEVFILE_STACK_INDEX")
	stackBase64IndexPath  = os.Getenv("DEVFILE_STACK_BASE64_INDEX")
	enableTelemetry       = util.IsTelemetryEnabled()
	registry              = util.GetOptionalEnv("REGISTRY_NAME", "devfile-registry")
)
