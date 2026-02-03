package gofiberswagger

import (
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
)

type RouteInfo = openapi3.Operation

var (
	acquiredRoutesInfo map[string]*RouteInfo
	mutex              = &sync.Mutex{}
)

func RegisterRoute(method string, path string, info *RouteInfo) {
	mutex.Lock()
	defer mutex.Unlock()

	if acquiredRoutesInfo == nil {
		acquiredRoutesInfo = make(map[string]*RouteInfo)
	}
	if info == nil {
		info = &RouteInfo{}
	}
	acquiredRoutesInfo[getAcquiredRoutesInfoId(method, path)] = info
}

func getAcquiredRoutesInfo(method string, path string) *RouteInfo {
	mutex.Lock()
	defer mutex.Unlock()

	if acquiredRoutesInfo == nil {
		return nil
	}
	return acquiredRoutesInfo[getAcquiredRoutesInfoId(method, path)]
}

func getAcquiredRoutesInfoId(method string, path string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(method)+path, " ", ""), "//", "/")
}
