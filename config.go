package puremail

import (
	"context"
	"time"
)

// // // // // // // // // //

type ConfigMxObj struct {
	TllPos       time.Duration
	TllNeg       time.Duration
	RefreshAhead time.Duration

	TimeoutDns      time.Duration
	TimeoutDnsBurst time.Duration
	TimeoutRefresh  time.Duration

	ShardAbs     byte // must be 1..31
	ShardMaxSize uint32

	ConcurrencyLimitLookupMX uint32
}

type ConfigObj struct {
	NoCache bool
	MX      ConfigMxObj

	Ctx context.Context
}

// //

var DefaultConfig = &ConfigObj{
	NoCache: true,
	MX: ConfigMxObj{
		TllPos:       6 * time.Hour,
		TllNeg:       15 * time.Minute,
		RefreshAhead: 10 * time.Minute,

		TimeoutDns:      400 * time.Millisecond,
		TimeoutDnsBurst: 2 * time.Second,
		TimeoutRefresh:  90 * time.Second,

		ShardAbs:     4,
		ShardMaxSize: 10_000,

		ConcurrencyLimitLookupMX: 250,
	},

	Ctx: context.Background(),
}
