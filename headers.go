package traefik_headers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wbpaygate/traefik-headers/internal/keeper"
)

const DEBUG = false

func CreateConfig() *Config {
	return &Config{
		KeeperReloadInterval: "30s",
		HeadersData:          `{}`,
	}
}

type Config struct {
	KeeperHeadersKey     string `json:"keeperHeadersKey,omitempty"`
	KeeperURL            string `json:"keeperURL,omitempty"`
	KeeperReqTimeout     string `json:"keeperReqTimeout,omitempty"`
	KeeperAdminPassword  string `json:"keeperAdminPassword,omitempty"`
	HeadersData          string `json:"headersData,omitempty"`
	KeeperReloadInterval string `json:"keeperReloadInterval,omitempty"`
}

type headers struct {
	headers http.Header
}

type Headers struct {
	name string
	next http.Handler
	cnt  *int32
	l    *log.Logger
}

type GlobalHeaders struct {
	config     *Config
	version    *keeper.Resp
	settings   keeper.Settings
	umtx       sync.Mutex
	curheader  *int32
	headers    []*headers
	rawheaders []byte
	ticker     *time.Ticker
	tickerto   time.Duration
	icnt       *int32
}

var ghs *GlobalHeaders

const HEADERS = 5

func init() {
	ghs = &GlobalHeaders{
		curheader:  new(int32),
		headers:    make([]*headers, HEADERS),
		version:    &keeper.Resp{},
		rawheaders: []byte(""),
		icnt:       new(int32),
	}
	ghs.headers[0] = &headers{
		headers: make(http.Header),
	}
	config := CreateConfig()
	to := 30 * time.Second
	if du, err := time.ParseDuration(string(config.KeeperReloadInterval)); err == nil {
		to = du
	}
	ghs.ticker = time.NewTicker(to)
	ghs.tickerto = to
	ghs.configure(nil, config)
	go func() {
		for {
			select {
			case <-ghs.ticker.C:
				ghs.sync()
			}
		}
	}()
	locallog("init")
}

// New created a new plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	var l *log.Logger
	if DEBUG {
		f, err := os.CreateTemp("/tmp", "log")
		if err == nil {
			l = log.New(f, "", 0)
			l.Println("start")
		}
	}
	locallog(fmt.Sprintf("config keeper key: %q, url: %q", config.KeeperHeadersKey, config.KeeperURL))
	if len(config.KeeperHeadersKey) == 0 {
		locallog("config: config: keeperHeadersKey is empty")
	}
	if len(config.KeeperURL) == 0 {
		locallog("config: keeperURL is empty")
	}
	if len(config.KeeperAdminPassword) == 0 {
		locallog("config: keeperAdminPassword is empty")
	}
	r := newHeaders(ctx, next, config, name)
	r.l = l
	return r, nil
}

func (g *GlobalHeaders) sync() {
	g.umtx.Lock()
	defer g.umtx.Unlock()
	locallog("sync")
	err := ghs.setFromSettings()
	if err != nil {
		locallog("cant get headers from keeper: ", err)
	}
}

func (g *GlobalHeaders) configure(ctx context.Context, config *Config) {
	to := 300 * time.Second
	if du, err := time.ParseDuration(string(config.KeeperReqTimeout)); err == nil {
		to = du
	}
	if ctx != nil {
		i := atomic.AddInt32(g.icnt, 1)
		locallog("run instance. cnt: ", i)
	}
	g.umtx.Lock()
	defer g.umtx.Unlock()

	if to, err := time.ParseDuration(string(config.KeeperReloadInterval)); err == nil && ghs.tickerto != to {
		g.ticker.Reset(to)
		ghs.tickerto = to
	}
	g.settings = keeper.New(config.KeeperURL, to, config.KeeperAdminPassword)
	g.config = config
	err := ghs.setFromSettings()
	if err != nil {
		if ctx == nil {
			locallog(fmt.Sprintf("init0: keeper: %v. try init from middleware HeadersData configuration", err))
		} else {
			locallog(fmt.Sprintf("init: keeper: %v. try init from middleware HeadersData configuration", err))
		}
		err = ghs.setFromData()
		if err != nil {
			if ctx == nil {
				locallog(fmt.Sprintf("init0: data: %v", err))
			} else {
				locallog(fmt.Sprintf("init: data: %v", err))
			}
		}
	}
}

func NewHeaders(next http.Handler, config *Config, name string) *Headers {
	return newHeaders(nil, next, config, name)
}

func newHeaders(ctx context.Context, next http.Handler, config *Config, name string) *Headers {
	r := &Headers{
		name: name,
		next: next,
		cnt:  new(int32),
	}
	ghs.configure(ctx, config)
	return r
}

func (h *Headers) log(v ...any) {
	if h.l != nil {
		h.l.Println(v...)
	}
}

func locallog(v ...any) {
	_, _ = os.Stderr.WriteString(fmt.Sprintf("time=%q traefikPlugin=\"headers\" msg=%q\n", time.Now().UTC().Format("2006-01-02 15:04:05Z"), fmt.Sprint(v...)))
}
