package internal

func (cmd *Cmd) Report(tree *Tree, config *Config, fileMaxLen, stmtsMaxLen, fullPathMaxLen int, args []string) {
	tree.Render(config, fileMaxLen, stmtsMaxLen, fullPathMaxLen, args)
}
