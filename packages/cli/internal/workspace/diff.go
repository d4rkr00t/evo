package workspace

type WorkspacesDiff struct {
	Changed             bool
	FilesChanged        bool
	LocalDepsChanged    bool
	ExternalDepsChanged bool
	TargetsChanged      bool
}

func DiffWorkspaces(ws1 *Workspace, ws2 *Workspace) WorkspacesDiff {
	var diff = WorkspacesDiff{
		Changed:             false,
		FilesChanged:        false,
		LocalDepsChanged:    false,
		ExternalDepsChanged: false,
		TargetsChanged:      false,
	}

	if ws1.Hash == ws2.Hash {
		return diff
	}

	diff.Changed = true

	if ws1.FilesHash != ws2.FilesHash {
		diff.FilesChanged = true
	}

	if ws1.LocalDepsHash != ws2.LocalDepsHash {
		diff.LocalDepsChanged = true
	}

	if ws1.ExtDepsHash != ws2.ExtDepsHash {
		diff.ExternalDepsChanged = true
	}

	if ws1.TargetsHash != ws2.TargetsHash {
		diff.TargetsChanged = true
	}

	return diff
}
