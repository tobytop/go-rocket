package ware

import (
	"context"
	"go-rocket/metadata"
)

type HandlerUnit func(ctx context.Context, data *metadata.MetaData) (any, error)

func (e HandlerUnit) WareBuild() Middleware {
	return func(next HandlerUnit) HandlerUnit {
		return func(ctx context.Context, data *metadata.MetaData) (any, error) {
			if result, err := e(ctx, data); err != nil {
				return result, err
			} else {
				return next(ctx, data)
			}
		}
	}
}

type Middleware func(HandlerUnit) HandlerUnit

type AfterUnit func(data *metadata.AfterMetaData) (any, error)
