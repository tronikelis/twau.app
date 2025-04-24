package about

import "twau.app/pkgs/server/req"

func getIndex(ctx req.ReqContext) error {
	return pageIndex().Render(ctx.Context(), ctx.Writer())
}
