package plugins

import (
	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/hashicorp/go-plugin"
)

type PluginType int32

const (
	DATASOURCE_PLUGIN PluginType = iota
	CONTROLLER_PLUGIN
)

const (
	ErrInvalidPluginType = bg.Error("invalid plugin type")
)

var (
	HandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "LANIAKEA_PLUGIN_MAGIC_COOKIE",
		MagicCookieValue: "a56e5daaa516e17d3d4b3d4685df9f8ca59c62c2d818cd5a7df13c039f134e16",
	}
)

type PluginManager struct {
	pluginMap map[string]plugin.Plugin
}

func NewPluginManager(listOfPlugins map[string]PluginType) (*PluginManager, error) {
	plugins := make(map[string]plugin.Plugin)
	var err error
loop:
	for name, pluginType := range listOfPlugins {
		var plug plugin.Plugin
		switch pluginType {
		case DATASOURCE_PLUGIN:
			plug = &DatasourcePlugin{}
		case CONTROLLER_PLUGIN:
			plug = &ControllerPlugin{}
		default:
			err = ErrInvalidPluginType
			break loop
		}
		plugins[name] = plug
	}
	if err != nil {
		return nil, err
	}
	return &PluginManager{
		pluginMap: plugins,
	}, nil
}
