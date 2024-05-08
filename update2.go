package traefik_headers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

func (g *GlobalHeaders) update(b []byte) error {
	tmph := make(http.Header)
	if err := json.Unmarshal(b, &tmph); err != nil {
		return err
	}

	newh := make(http.Header, len(tmph))
	for k, vv := range tmph {
		vv2 := make([]string, 0, len(vv))
		for _, v := range vv {
			v = strings.TrimSpace(v)
			if len(v) != 0 {
				vv2 = append(vv2, v)
			}
		}
		if len(vv2) > 0 {
			newh[http.CanonicalHeaderKey(k)] = vv2
		}
	}

	curheader := int(atomic.LoadInt32(g.curheader))
	oldh := g.headers[curheader].headers
	locallog(fmt.Sprintf("use %d headers", len(newh)))
	if compHeader(newh, oldh) {
		return nil
	}
	newheaders := &headers{
		headers: newh,
	}
	curheader = (curheader + 1) % HEADERS
	g.headers[curheader] = newheaders
	atomic.StoreInt32(g.curheader, int32(curheader))
	return nil
}

func compHeader(h1, h2 http.Header) bool {
	if len(h1) != len(h2) {
		return false
	}
	for k1, v1 := range h1 {
		if v2, ok := h2[k1]; ok {
			if !compareSliceString(v1, v2) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func compareSliceString(v1, v2 []string) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := 0; i < len(v1); i++ {
		if v1[i] != v2[i] {
			return false
		}
	}
	return true
}
