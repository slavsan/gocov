package internal

func (cmd *Cmd) Report(tree *Tree, config *Config, stats Stats, args []string) {
	tree.Render(config, stats, args)
}
