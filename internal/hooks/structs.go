package hooks

import "github.com/nats-io/nats.go"

type ArtifactsSvcHooks struct{
	nc *nats.Conn
}