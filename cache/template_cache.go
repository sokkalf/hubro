package cache

import (
	"html/template"
	"log/slog"
	"sync"

	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/utils/broker"
)

type CachedTemplate struct {
	template *template.Template
}

var templateCache map[string]CachedTemplate
var templateCacheMutex = sync.RWMutex{}
var MsgBroker *broker.Broker[index.Message]

func InitTemplateCache() {
	templateCache = make(map[string]CachedTemplate)
	MsgBroker = broker.NewBroker[index.Message]()
	go MsgBroker.Start()
	go func() {
		msgChan := MsgBroker.Subscribe()
		for {
			switch <-msgChan {
			case index.Reset:
				slog.Debug("Resetting template cache")
				Clear()
			default:
				slog.Error("Unknown message received")
			}
		}
	}()
}

func Get(key string) *template.Template {
	templateCacheMutex.RLock()
	defer templateCacheMutex.RUnlock()
	if val, ok := templateCache[key]; ok {
		return val.template
	}
	return nil
}

func Put(key string, template *template.Template) {
	templateCacheMutex.Lock()
	defer templateCacheMutex.Unlock()
	templateCache[key] = CachedTemplate{template: template}
}

func Clear() {
	templateCacheMutex.Lock()
	defer templateCacheMutex.Unlock()
	templateCache = make(map[string]CachedTemplate)
}
