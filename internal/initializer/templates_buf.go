package initializer

func bufYamlTemplate() string {
	return `version: v2
modules:
  - path: api/proto
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
`
}

func bufGenYamlTemplate() string {
	return `version: v2
managed:
  enabled: true
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen/proto
    opt: paths=source_relative
  - remote: buf.build/grpc/go
    out: gen/proto
    opt: paths=source_relative
`
}
