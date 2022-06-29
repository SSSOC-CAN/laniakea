package plugins

import (
	"strings"

	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/hashicorp/go-plugin"
)

const (
	DATASOURCE_STR         = "datasource"
	CONTROLLER_STR         = "controller"
	ErrInvalidPluginString = bg.Error("invalid plugin string")
	ErrInvalidPluginType   = bg.Error("invalid plugin type")
	ErrDuplicatePluginName = bg.Error("identical plugin names")
)

var (
	HandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "LANIAKEA_PLUGIN_MAGIC_COOKIE",
		MagicCookieValue: "a56e5daaa516e17d3d4b3d4685df9f8ca59c62c2d818cd5a7df13c039f134e16",
	}
)

type PluginManager struct {
	pluginMap   map[string]plugin.Plugin
	pluginExecs map[string]string
}

func NewPluginManager(listOfPlugins []string) (*PluginManager, error) {
	plugins := make(map[string]plugin.Plugin)
	execs := make(map[string]string)
	var err error
loop:
	for _, pluginStr := range listOfPlugins {
		split := strings.Split(pluginStr, ":")
		if len(split) == 0 {
			err = ErrInvalidPluginString
			break loop
		}
		if _, ok := plugins[split[0]]; ok {
			err = ErrDuplicatePluginName
			break loop
		}
		var plugType plugin.Plugin
		switch split[1] {
		case DATASOURCE_STR:
			plugType = &DatasourcePlugin{}
		case CONTROLLER_STR:
			plugType = &ControllerPlugin{}
		default:
			err = ErrInvalidPluginType
			break loop
		}
		plugins[split[0]] = plugType
		execs[split[0]] = split[2]
	}
	if err != nil {
		return nil, err
	}
	return &PluginManager{
		pluginMap:   plugins,
		pluginExecs: execs,
	}, nil
}
