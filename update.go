package traefik_headers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync/atomic"
)

func (g *GlobalHeaders) logWorkingHeaders() {
	buf := new(bytes.Buffer)
	if err := json.Compact(buf, g.rawheaders); err != nil {
		locallog(fmt.Sprintf("working headers: %s", g.rawheaders))
	} else {
		locallog(fmt.Sprintf("working headers: %s", buf.String()))
	}
	for k, vv := range g.headers[int(atomic.LoadInt32(ghs.curheader))].headers {
		locallog(fmt.Sprintf("working header key %d,%d: %q: %+q", g.version.Version, g.version.ModRevision, k, vv))
	}
}

func (g *GlobalHeaders) setFromData() error {
	defer g.logWorkingHeaders()
	if g.config == nil {
		return fmt.Errorf("config not specified")
	}
	b := []byte(g.config.HeadersData)
	err := g.update(b)
	if err == nil {
		g.rawheaders = b
		g.version.Version = 0
		g.version.ModRevision = 0
	}
	return err
}

func (g *GlobalHeaders) setFromSettings() error {
	if g.config == nil {
		g.logWorkingHeaders()
		return fmt.Errorf("config not specified")
	}
	result, err := g.settings.Get(g.config.KeeperHeadersKey)
	if err != nil {
		g.logWorkingHeaders()
		return err
	}
	if result == nil || len(result.Value) == 0 {
		g.logWorkingHeaders()
		return fmt.Errorf("settings not found in keeper")
	}

	if !g.version.Equal(result) {
		defer g.logWorkingHeaders()
		if g.version != nil {
			locallog(fmt.Sprintf("old configuration: Version: %d, ModRevision: %d", g.version.Version, g.version.ModRevision))
		}
		err = g.update([]byte(result.Value))
		if err != nil {
			return err
		}
		g.rawheaders = []byte(result.Value)
		g.version = result
		locallog(fmt.Sprintf("new configuration loaded: Version: %d, ModRevision: %d", g.version.Version, g.version.ModRevision))
	}
	return nil
}

func (h *Headers) Update(b []byte) error {
	return ghs.update(b)
}
