package ware

import (
	"context"
	"go-rocket/metadata"
)

type HandlerUnit func(ctx context.Context, data *metadata.MetaData) (response any, err error)

func (e HandlerUnit) WareBuild() Middleware {
	return func(next HandlerUnit) HandlerUnit {
		return func(ctx context.Context, data *metadata.MetaData) (response any, err error) {
			if _, err := e(ctx, data); err != nil {
				return nil, err
			} else {
				return next(ctx, data)
			}
		}
	}
}

type Middleware func(HandlerUnit) HandlerUnit

type AfterUnit func(data *metadata.AfterMetaData) (response any, err error)
