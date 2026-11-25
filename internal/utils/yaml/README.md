# YAML Utilities

This package provides YAML marshaling and unmarshaling utilities using `gopkg.in/yaml.v3`, which properly respects `yaml:` struct tags.

## Why this package?

The `sigs.k8s.io/yaml` package used in some parts of the codebase only respects `json:` struct tags, not `yaml:` tags. This is because it marshals to JSON first, then converts to YAML.

For OpenTelemetry Collector configuration structs that only have `yaml:` tags, you must use this package instead of `sigs.k8s.io/yaml` or `test.YAMLString()`.

## Usage

### Basic Marshaling

```go
import utilyaml "github.com/openshift/cluster-logging-operator/internal/utils/yaml"

type Config struct {
    Name    string   `yaml:"name"`
    Enabled bool     `yaml:"enabled,omitempty"`
    Tags    []string `yaml:"tags,omitempty"`
}

config := Config{
    Name: "my-config",
    Tags: []string{"prod", "us-east"},
}

// Get YAML string (returns empty string on error)
yamlStr := utilyaml.MustMarshal(config)
fmt.Println(yamlStr)
// Output:
// name: my-config
// tags:
//   - prod
//   - us-east
```

### Error Handling

```go
// For proper error handling, use Marshal instead of MustMarshal
bytes, err := utilyaml.Marshal(config)
if err != nil {
    log.Error(err, "failed to marshal config")
    return
}
```

### Unmarshaling

```go
yamlData := `
name: my-config
enabled: true
tags:
  - prod
  - us-east
`

var config Config
err := utilyaml.Unmarshal([]byte(yamlData), &config)
if err != nil {
    log.Error(err, "failed to unmarshal YAML")
    return
}
```

### Converting to Map

```go
// Convert struct to map[string]interface{} while preserving YAML tag behavior
configMap, err := utilyaml.MarshalToMap(config)
if err != nil {
    log.Error(err, "failed to convert to map")
    return
}
```

## API Functions

- **`MustMarshal(v interface{}) string`**: Returns YAML string or empty string on error. Recovers from panics.
- **`Marshal(v interface{}) ([]byte, error)`**: Returns YAML bytes or error.
- **`Unmarshal(data []byte, v interface{}) error`**: Unmarshals YAML data into a value.
- **`UnmarshalStrict(data []byte, v interface{}) error`**: Unmarshals with strict mode (fails on unknown fields).
- **`MarshalToMap(v interface{}) (map[string]interface{}, error)`**: Converts struct to map preserving YAML behavior.

## Comparison with sigs.k8s.io/yaml

| Package | Struct Tag | Use Case |
|---------|-----------|----------|
| `sigs.k8s.io/yaml` | `json:` | Kubernetes resources, test utilities |
| `gopkg.in/yaml.v3` (this package) | `yaml:` | OpenTelemetry Collector configs, pure YAML |

If your struct has both `json:` and `yaml:` tags, either package will work. If it only has `yaml:` tags, you **must** use this package.
