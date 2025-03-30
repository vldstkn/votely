package db

import (
	"context"
	"github.com/tarantool/go-tarantool/v2"
	"time"
)

type TrntlDb struct {
	*tarantool.Connection
}

func InitTrntlDb(dialer tarantool.NetDialer, opts tarantool.Opts) (*TrntlDb, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		return nil, err
	}
	return &TrntlDb{
		conn,
	}, nil
}
