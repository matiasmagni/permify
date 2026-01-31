package coverage

import (
	"testing"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/token"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	info1 := SourceInfo{Line: 1, Column: 1}
	info2 := SourceInfo{Line: 2, Column: 5}

	r.Register("path1", info1, "OR")
	r.Register("path2", info2, "AND")

	r.Visit("path1")

	uncovered := r.Report()

	if len(uncovered) != 1 {
		t.Errorf("expected 1 uncovered node, got %d", len(uncovered))
	}

	if uncovered[0].Path != "path2" {
		t.Errorf("expected path2 to be uncovered, got %s", uncovered[0].Path)
	}

	r.Visit("path2")
	uncovered = r.Report()
	if len(uncovered) != 0 {
		t.Errorf("expected 0 uncovered nodes, got %d", len(uncovered))
	}
}

func TestDiscover(t *testing.T) {
	sch := &ast.Schema{
		Statements: []ast.Statement{
			&ast.EntityStatement{
				Name: token.Token{Literal: "repository"},
				PermissionStatements: []ast.Statement{
					&ast.PermissionStatement{
						Name: token.Token{Literal: "edit", PositionInfo: token.PositionInfo{LinePosition: 1, ColumnPosition: 12}},
						ExpressionStatement: &ast.ExpressionStatement{
							Expression: &ast.InfixExpression{
								Op:       token.Token{Literal: "or", PositionInfo: token.PositionInfo{LinePosition: 1, ColumnPosition: 20}},
								Operator: ast.OR,
								Left: &ast.Identifier{
									Idents: []token.Token{
										{Literal: "owner", PositionInfo: token.PositionInfo{LinePosition: 1, ColumnPosition: 15}},
									},
								},
								Right: &ast.Identifier{
									Idents: []token.Token{
										{Literal: "admin", PositionInfo: token.PositionInfo{LinePosition: 1, ColumnPosition: 25}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	r := NewRegistry()
	Discover(sch, r)

	report := r.Report()
	if len(report) != 4 {
		t.Errorf("expected 4 nodes (PERMISSION, OR, LEAF, LEAF), got %d", len(report))
	}

	// Verify paths: edit (PERMISSION), edit.op (OR), edit.op.0 (LEAF), edit.op.1 (LEAF)
	foundEdit := false
	foundEditOp := false
	foundEditOp0 := false
	foundEditOp1 := false

	for _, node := range report {
		switch node.Path {
		case "repository#edit":
			foundEdit = true
			if node.Type != "PERMISSION" {
				t.Errorf("expected PERMISSION type for repository#edit, got %s", node.Type)
			}
		case "repository#edit.op":
			foundEditOp = true
			if node.Type != "OR" && node.Type != "or" {
				t.Errorf("expected OR type for repository#edit.op, got %s", node.Type)
			}
		case "repository#edit.op.0":
			foundEditOp0 = true
			if node.Type != "LEAF" {
				t.Errorf("expected LEAF type for repository#edit.op.0, got %s", node.Type)
			}
		case "repository#edit.op.1":
			foundEditOp1 = true
			if node.Type != "LEAF" {
				t.Errorf("expected LEAF type for repository#edit.op.1, got %s", node.Type)
			}
		}
	}

	if !foundEdit || !foundEditOp || !foundEditOp0 || !foundEditOp1 {
		t.Errorf("missing paths: edit:%v, edit.op:%v, edit.op.0:%v, edit.op.1:%v", foundEdit, foundEditOp, foundEditOp0, foundEditOp1)
	}
}
