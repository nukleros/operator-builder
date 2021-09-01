package inspect

import (
	"fmt"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/parser"
	"gopkg.in/yaml.v3"
)

type YAMLResult struct {
	*parser.Result
	Nodes []*yaml.Node
}

func (s *Inspector) InspectYAML(data []byte, transforms ...YAMLTransformer) (*yaml.Node, []*YAMLResult, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return &node, nil, fmt.Errorf("error unmarshaling yaml, %w", err)
	}

	results := s.inspectYAML(&node)
	for _, result := range results {
		if v, ok := result.Result.Object.(error); ok {
			return &node, results, v
		}
	}

	for _, transform := range transforms {
		if err := transform(results...); err != nil {
			return &node, nil, err
		}
	}

	return &node, results, nil
}

func (s *Inspector) inspectYAML(nodes ...*yaml.Node) (results []*YAMLResult) {
	for _, node := range nodes {
		results = append(results, s.inspectYAMLComments(node)...)

		if node.Kind == yaml.MappingNode {
			results = append(results, s.inspectYAMLMap(node.Content...)...)
		} else if node.Content != nil {
			results = append(results, s.inspectYAML(node.Content...)...)
		}
	}

	return results
}

func (s *Inspector) inspectYAMLMap(nodes ...*yaml.Node) (results []*YAMLResult) {
	for i := 0; i < len(nodes); i += 2 {
		results = append(results, s.inspectYAMLComments(nodes[i], nodes[i+1])...)

		if nodes[i+1].Kind == yaml.MappingNode {
			results = append(results, s.inspectYAMLMap(nodes[i+1].Content...)...)
		} else {
			results = append(results, s.inspectYAML(nodes[i+1].Content...)...)
		}
	}

	return results
}

func (s *Inspector) inspectYAMLComments(nodes ...*yaml.Node) (results []*YAMLResult) {
	var markers []*parser.Result

	for _, node := range nodes {
		markers = append(markers, s.parse(fmt.Sprintf("%s\n%s\n%s", node.HeadComment, node.LineComment, node.FootComment))...)
	}

	for _, marker := range markers {
		result := &YAMLResult{
			Result: marker,
			Nodes:  nodes,
		}

		results = append(results, result)
	}

	return results
}
