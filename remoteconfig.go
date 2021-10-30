// Package remoteconfig implements a library for reading JSON encoded
// configuration from a public location on the internet - e.g. a public Git
// repo, or a 'secret' GitHub Gist (which are unlisted but public).
// Because it only supports public configuration sources, it's obviously not
// suitable for secrets.
package remoteconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type RemoteConfig interface {
	Unmarshal(v interface{}) error
}

type remoteConfig struct {
	URL      string
	Duration time.Duration
	File     []byte
}

func New(url string) (RemoteConfig, error) {
	rc := &remoteConfig{
		URL:      url,
		Duration: time.Minute,
	}

	if err := rc.fetch(); err != nil {
		return nil, err
	}

	go rc.startRefresh()

	return rc, nil
}

func (rc *remoteConfig) Unmarshal(v interface{}) error {
	return json.Unmarshal(rc.File, v)
}

func (rc *remoteConfig) fetch() error {
	rsp, err := http.Get(rc.URL)
	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode > 299 {
		return fmt.Errorf("remoteconfig: fetch failed with status code %d", rsp.StatusCode)
	}

	file, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	rc.File = file

	return nil
}

func (rc *remoteConfig) startRefresh() {
	ticker := time.NewTicker(rc.Duration)
	for range ticker.C {
		err := rc.fetch()
		if err != nil {
			log.Printf("error fetching remoteconfig: %v", err)
			continue
		}
		log.Println("fetched remoteconfig")
	}
}
