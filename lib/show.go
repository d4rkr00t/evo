package lib

func ShowHash(ctx Context, ws_name string) {
	ctx.stats.StartMeasure("show-hash", MEASURE_KIND_STAGE)
	ctx.logger.Log()
	ctx.logger.LogWithBadge("cwd", "   "+ctx.cwd)

	var workspaces, _ = GetWorkspaces(ctx.root, &ctx.config, &ctx.cache)

	var ws, ok = workspaces[ws_name]
	if !ok {
		ctx.logger.Log("Package", ws_name, "not found!")
		return
	}

	var lg = ctx.logger.CreateGroup()
	lg.Start("Package hash consists of:")

	lg.Log("Files:")
	var files = ws.get_files()
	for _, file_name := range files {
		lg.Log("–", file_name)
	}

	lg.Log()
	lg.Log("Deps:")
	var deps = ws.Deps
	for dep_name, dep_ver := range deps {
		lg.Log("–", dep_name, ":", dep_ver)
	}

	lg.Log()
	lg.Log("Rules:")
	var rules = ws.get_rules()
	for _, rule := range rules {
		lg.Log("–", rule)
	}

	lg.Log()
	lg.Log("Hash:")
	lg.Log("–", ws.Hash(&workspaces))

	lg.End(ctx.stats.StopMeasure("show-hash"))
}
