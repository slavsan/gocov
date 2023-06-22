package internal

func (cmd *Cmd) Report(tree *Tree, config *Config, fileMaxLen, stmtsMaxLen int) {
	tree.Render(config, fileMaxLen, stmtsMaxLen)
}
