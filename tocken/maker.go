package tocken

import "time"

type Maker interface {
	CreateTocken(username string, duration time.Duration) (string, error)
	VeifyTocken(token string) (*Payload, error)
}
