package consensus

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

// RenderDAGToDot is a helper function for debugging tests.
// It requires graphviz installed.
func (tc *testConsensus) RenderDAGToDot(filename string) error {
	dotScript, _ := tc.convertToDot()
	return renderDotScript(dotScript, filename)
}

func (tc *testConsensus) convertToDot() (string, error) {
	var dotScriptBuilder strings.Builder
	dotScriptBuilder.WriteString("digraph {\n\trankdir = TB; \n")

	edges := []string{}

	blocksIterator, err := tc.blockStore.AllBlockHashesIterator(tc.databaseContext)
	if err != nil {
		return "", err
	}
	defer blocksIterator.Close()

	for ok := blocksIterator.First(); ok; ok = blocksIterator.Next() {
		hash, err := blocksIterator.Get()
		if err != nil {
			return "", err
		}
		dotScriptBuilder.WriteString(fmt.Sprintf("\t\"%s\";\n", hash))

		parents, err := tc.dagTopologyManager.Parents(hash)
		if err != nil {
			return "", err
		}

		for _, parentHash := range parents {
			edges = append(edges, fmt.Sprintf("\t\"%s\" -> \"%s\";", hash, parentHash))
		}
	}

	dotScriptBuilder.WriteString("\n")

	dotScriptBuilder.WriteString(strings.Join(edges, "\n"))

	dotScriptBuilder.WriteString("\n}")

	return dotScriptBuilder.String(), nil
}

func renderDotScript(dotScript string, filename string) error {
	command := exec.Command("dot", "-Tsvg")
	stdin, err := command.StdinPipe()
	if err != nil {
		return fmt.Errorf("Error creating stdin pipe: %s", err)
	}
	spawn("renderDotScript", func() {
		defer stdin.Close()

		_, err = io.WriteString(stdin, dotScript)
		if err != nil {
			panic(fmt.Errorf("Error writing dotScript into stdin pipe: %s", err))
		}
	})

	var stderr bytes.Buffer
	command.Stderr = &stderr
	svg, err := command.Output()
	if err != nil {
		return fmt.Errorf("Error getting output of dot: %s\nstderr:\n%s", err, stderr.String())
	}

	return ioutil.WriteFile(filename, svg, 0600)
}
