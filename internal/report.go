package internal

func (cmd *Cmd) Report(tree *Tree, config *Config, foo Stats, args []string) {
	tree.Render(config, foo, args)
}
