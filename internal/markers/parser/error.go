package parser

import "fmt"

func (p *Parser) error(err error) stateFn {
	var markerName string

	if p.currentDefinition != nil {
		markerName = p.currentDefinition.GetName()
	} else {
		markerName = "Unknown Marker"
	}
	p.items <- &Result{
		Object:     fmt.Errorf("%w, on marker %s at %+v", err, markerName, p.currentLexeme.Pos),
		MarkerText: p.scopeBuffer,
	}

	return nil
}
