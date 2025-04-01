package home

import "github.com/tronikelis/maruchi"

func getIndex(ctx maruchi.ReqContext) {
	if err := pageIndex().Render(ctx.Context(), ctx.Writer()); err != nil {
		panic(err)
	}
}
