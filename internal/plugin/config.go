package plugin

import (
	"github.com/buger/jsonparser"
)

// Configuration is the resolved configuration of the plugin, which has no
// options since `tofu fmt` has none.
type Configuration struct{}

type configDiagnostic struct {
	PropertyName string
	Message      string
}

// resolveConfig resolves the plugin's configuration from its JSON object,
// where every property is unknown.
func resolveConfig(pluginConfig []byte) (Configuration, []configDiagnostic) {
	var diagnostics []configDiagnostic
	for _, key := range objectKeys(pluginConfig) {
		diagnostics = append(diagnostics, configDiagnostic{
			PropertyName: key,
			Message:      "Unknown property in configuration",
		})
	}
	return Configuration{}, diagnostics
}

// objectKeys gets the keys of a JSON object in order.
func objectKeys(data []byte) []string {
	var keys []string
	if len(data) == 0 {
		return keys
	}
	_ = jsonparser.ObjectEach(data, func(key, _ []byte, _ jsonparser.ValueType, _ int) error {
		keys = append(keys, string(key))
		return nil
	})
	return keys
}

func marshalConfig(Configuration) []byte {
	return []byte("{}")
}

func marshalConfigDiagnostics(diagnostics []configDiagnostic) []byte {
	data := []byte{'['}
	for i, diagnostic := range diagnostics {
		if i != 0 {
			data = append(data, ',')
		}
		data = append(data, `{"propertyName":`...)
		data = appendJSONString(data, diagnostic.PropertyName)
		data = append(data, `,"message":`...)
		data = appendJSONString(data, diagnostic.Message)
		data = append(data, '}')
	}
	return append(data, ']')
}
