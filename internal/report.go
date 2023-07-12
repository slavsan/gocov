package internal

import (
	"fmt"
	"log"
	"strings"
)

func (cmd *Cmd) Report(tree *Tree, stats Stats, args []string, files map[string]*covFile, moduleDir string) {
	if cmd.config.HTMLOutput {
		cmd.ReportHTML(tree, stats, args, files, moduleDir)
		return
	}

	tree.Render(cmd.config, stats, args)
}

func (cmd *Cmd) ReportHTML(tree *Tree, stats Stats, args []string, files map[string]*covFile, moduleDir string) {
	var (
		filesSourceBuilder strings.Builder
		objectBuilder      strings.Builder
		sb                 strings.Builder
	)

	tree.RenderHTML(cmd, &filesSourceBuilder, &objectBuilder, cmd.config, stats, args, cmd.fsys, files, moduleDir)

	err := cmd.fw.Open("coverage.html")
	if err != nil {
		log.Fatal(err)
		return
	}
	f := cmd.fw

	sb.WriteString(`<script class="tree-data" type="application/json">`)
	sb.WriteString(objectBuilder.String())
	sb.WriteString("</script>")

	tmpl := strings.ReplaceAll(tmpl, "<!-- REPORT -->", sb.String())
	tmpl = strings.ReplaceAll(tmpl, "<!-- SCRIPT -->", Script)
	tmpl = strings.ReplaceAll(tmpl, "<!-- SOURCE -->", filesSourceBuilder.String())

	_, _ = fmt.Fprint(f, tmpl)
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
}

func escape(value string) string {
	value = strings.ReplaceAll(value, "&", "amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	value = strings.ReplaceAll(value, "&lt;/span&gt;", "</span>")
	value = strings.ReplaceAll(value, "&lt;span style=\"background: pink\"&gt;", "<span style=\"background: pink\">")
	return value
}
