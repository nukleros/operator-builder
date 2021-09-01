package parser

import "fmt"

func (p *Parser) emit() error {
	output, err := p.currentDefinition.InflateObject()
	if err != nil {
		return fmt.Errorf("unable to inflate object, %w", err)
	}

	result := &Result{
		Object:     output,
		MarkerText: p.scopeBuffer,
	}

	p.items <- result

	p.flush()

	return nil
}

type Result struct {
	Object     interface{}
	MarkerText string
}
