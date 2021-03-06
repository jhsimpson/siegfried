// Copyright 2014 Richard Lehane. All rights reserved.
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

package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var pronom = struct {
	droid            string   // name of droid file e.g. DROID_SignatureFile_V78.xml
	container        string   // e.g. container-signature-19770502.xml
	reports          string   // directory where PRONOM reports are stored
	noreports        bool     // build signature directly from DROID file rather than PRONOM reports
	doubleup         bool     // include byte signatures for formats that also have container signatures
	inspect          bool     // setting for inspecting PRONOM signatures
	limit            []string // limit signature to a set of included PRONOM reports
	exclude          []string // exclude a set of PRONOM reports from the signature
	extensions       string   // directory where custom signature extensions are stored
	extend           []string
	extendc          []string //container extensions
	harvestURL       string
	harvestTimeout   time.Duration
	harvestTransport *http.Transport
	// archive puids
	zip    string
	tar    string
	gzip   string
	arc    string
	arc1_1 string
	warc   string
	// text puid
	text string
}{
	reports:          "pronom",
	extensions:       "custom",
	harvestURL:       "http://apps.nationalarchives.gov.uk/pronom/",
	harvestTimeout:   120 * time.Second,
	harvestTransport: &http.Transport{Proxy: http.ProxyFromEnvironment},
	zip:              "x-fmt/263",
	tar:              "x-fmt/265",
	gzip:             "x-fmt/266",
	arc:              "x-fmt/219",
	arc1_1:           "fmt/410",
	warc:             "fmt/289",
	text:             "x-fmt/111",
}

// GETTERS

// DROID returns the location of the DROID signature file.
// If not set, infers the latest file.
func Droid() string {
	if pronom.droid == "" {
		droid, err := latest("DROID_SignatureFile_V", ".xml")
		if err != nil {
			return ""
		}
		return filepath.Join(siegfried.home, droid)
	}
	if filepath.Dir(pronom.droid) == "." {
		return filepath.Join(siegfried.home, pronom.droid)
	}
	return pronom.droid
}

// DROID base returns the base filename of the DROID signature file.
// If not set, infers the latest file.
func DroidBase() string {
	if pronom.droid == "" {
		droid, err := latest("DROID_SignatureFile_V", ".xml")
		if err != nil {
			return ""
		}
		return droid
	}
	return pronom.droid
}

// Container returns the location of the DROID container signature file.
// If not set, infers the latest file.
func Container() string {
	if pronom.container == "" {
		container, err := latest("container-signature-", ".xml")
		if err != nil {
			return ""
		}
		return filepath.Join(siegfried.home, container)
	}
	if filepath.Dir(pronom.container) == "." {
		return filepath.Join(siegfried.home, pronom.container)
	}
	return pronom.container
}

// ContainerBase returns the base filename of the DROID container signature file.
// If not set, infers the latest file.
func ContainerBase() string {
	if pronom.container == "" {
		container, err := latest("container-signature-", ".xml")
		if err != nil {
			return ""
		}
		return container
	}
	return pronom.container
}

func latest(prefix, suffix string) (string, error) {
	var hits []string
	var ids []int
	files, err := ioutil.ReadDir(siegfried.home)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		nm := f.Name()
		if strings.HasPrefix(nm, prefix) && strings.HasSuffix(nm, suffix) {
			hits = append(hits, nm)
			id, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(nm, prefix), suffix))
			if err != nil {
				return "", err
			}
			ids = append(ids, id)
		}
	}
	if len(hits) == 0 {
		return "", fmt.Errorf("Config: no file in %s with prefix %s", siegfried.home, prefix)
	}
	if len(hits) == 1 {
		return hits[0], nil
	}
	max, idx := ids[0], 0
	for i, v := range ids[1:] {
		if v > max {
			max = v
			idx = i + 1
		}
	}
	return hits[idx], nil
}

// Reports returns the location of the PRONOM reports directory.
func Reports() string {
	if pronom.noreports || pronom.reports == "" {
		return ""
	}
	if filepath.Dir(pronom.reports) == "." {
		return filepath.Join(siegfried.home, pronom.reports)
	}
	return pronom.reports
}

// Inspect reports whether roy is being run in inspect mode.
func Inspect() bool {
	return pronom.inspect
}

// HasLimit reports whether a limited set of signatures has been selected.
func HasLimit() bool {
	return len(pronom.limit) > 0
}

// Limit takes a slice of puids and returns a new slice containing only those puids in the limit set.
func Limit(puids []string) []string {
	ret := make([]string, 0, len(pronom.limit))
	for _, v := range pronom.limit {
		for _, w := range puids {
			if v == w {
				ret = append(ret, v)
			}
		}
	}
	return ret
}

// HasExclude reports whether an exlusion set of signatures has been provided.
func HasExclude() bool {
	return len(pronom.exclude) > 0
}

func exclude(puids, ex []string) []string {
	ret := make([]string, 0, len(puids))
	for _, v := range puids {
		excluded := false
		for _, w := range ex {
			if v == w {
				excluded = true
				break
			}
		}
		if !excluded {
			ret = append(ret, v)
		}
	}
	return ret
}

// Exclude takes a slice of puids and omits those that are also in the pronom.exclude slice.
func Exclude(puids []string) []string {
	return exclude(puids, pronom.exclude)
}

// DoubleUp reports whether the doubleup flag has been set. This will cause byte signatures to be built for formats where container signatures are also provided.
func DoubleUp() bool {
	return pronom.doubleup
}

// ExcludeDoubles takes a slice of puids and a slice of container puids and exludes those that are in the container slice, if nodoubles is set.
func ExcludeDoubles(puids, cont []string) []string {
	return exclude(puids, cont)
}

func extensionPaths(e []string) []string {
	ret := make([]string, len(e))
	for i, v := range e {
		if filepath.Dir(v) == "." {
			ret[i] = filepath.Join(siegfried.home, pronom.extensions, v)
		} else {
			ret[i] = v
		}
	}
	return ret
}

// Extend reports whether a set of signature extensions has been provided.
func Extend() []string {
	return extensionPaths(pronom.extend)
}

// Extend reports whether a set of container signature extensions has been provided.
func ExtendC() []string {
	return extensionPaths(pronom.extendc)
}

// HarvestOptions reports the PRONOM url, timeout and transport.
func HarvestOptions() (string, time.Duration, *http.Transport) {
	return pronom.harvestURL, pronom.harvestTimeout, pronom.harvestTransport
}

// ZipPuid reports the puid for a zip archive.
func ZipPuid() string {
	return pronom.zip
}

// TextPuid reports the puid for a text file.
func TextPuid() string {
	return pronom.text
}

// IsArchive returns an Archive that corresponds to the provided puid (or none if no match).
func IsArchive(p string) Archive {
	switch p {
	case pronom.zip:
		return Zip
	case pronom.gzip:
		return Gzip
	case pronom.tar:
		return Tar
	case pronom.arc, pronom.arc1_1:
		return ARC
	case pronom.warc:
		return WARC
	}
	return None
}

// SETTERS

// SetDroid sets the name and/or location of the DROID signature file.
// I.e. can provide a full path or a filename relative to the HOME directory.
func SetDroid(d string) func() private {
	return func() private {
		pronom.droid = d
		return private{}
	}
}

// SetContainer sets the name and/or location of the DROID container signature file.
// I.e. can provide a full path or a filename relative to the HOME directory.
func SetContainer(c string) func() private {
	return func() private {
		pronom.container = c
		return private{}
	}
}

// SetReports sets the location of the PRONOM reports directory.
func SetReports(r string) func() private {
	return func() private {
		pronom.reports = r
		return private{}
	}
}

// SetNoReports instructs roy to build from the DROID signature file alone (and not from the PRONOM reports).
func SetNoReports() func() private {
	return func() private {
		pronom.noreports = true
		return private{}
	}
}

// SetDoubleUp causes byte signatures to be built for formats where container signatures are also provided.
func SetDoubleUp() func() private {
	return func() private {
		pronom.doubleup = true
		return private{}
	}
}

// SetInspect causes roy to run in inspect mode.
func SetInspect() func() private {
	return func() private {
		pronom.inspect = true
		return private{}
	}
}

// SetLimit limits the set of signatures built to the list provide.
func SetLimit(l []string) func() private {
	return func() private {
		pronom.limit = l
		return private{}
	}
}

// SetExclude excludes the provided signatures from those built.
func SetExclude(l []string) func() private {
	return func() private {
		pronom.exclude = l
		return private{}
	}
}

// SetExtend adds extension signatures to the build.
func SetExtend(l []string) func() private {
	return func() private {
		pronom.extend = l
		return private{}
	}
}

// SetExtendC adds container extension signatures to the build.
func SetExtendC(l []string) func() private {
	return func() private {
		pronom.extendc = l
		return private{}
	}
}

// unlike other setters, these are only relevant in the roy tool so can't be converted to the Option type

// SetHarvestTimeout sets a time limit on PRONOM harvesting.
func SetHarvestTimeout(d time.Duration) {
	pronom.harvestTimeout = d
}

// SetHarvestTransport sets the PRONOM harvesting transport.
func SetHarvestTransport(t *http.Transport) {
	pronom.harvestTransport = t
}
