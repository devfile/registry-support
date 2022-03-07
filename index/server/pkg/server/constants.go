package server

import (
	"os"

	"github.com/devfile/registry-support/index/server/pkg/util"
)

const (
	// Constants for resource names and media types
	archiveMediaType       = "application/x-tar"
	archiveName            = "archive.tar"
	devfileName            = "devfile.yaml"
	devfileNameHidden      = ".devfile.yaml"
	devfileConfigMediaType = "application/vnd.devfileio.devfile.config.v2+json"
	devfileMediaType       = "application/vnd.devfileio.devfile.layer.v1"
	pngLogoMediaType       = "image/png"
	pngLogoName            = "logo.png"
	svgLogoMediaType       = "image/svg+xml"
	svgLogoName            = "logo.svg"
	vsxMediaType           = "application/vnd.devfileio.vsx.layer.v1.tar"
	vsxName                = "vsx"

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
