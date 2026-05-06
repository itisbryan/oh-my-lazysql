package components

import (
	"strings"
	"testing"
)

func TestTreeNodeLabelAddsNerdFontPrefix(t *testing.T) {
	label := treeNodeLabel(NodeTypeTable, "addresses")

	if !strings.Contains(label, "addresses") {
		t.Fatalf("expected label to keep raw name, got %q", label)
	}

	if label == "addresses" {
		t.Fatal("expected label to include an icon prefix")
	}
}

func TestIsSystemSchema(t *testing.T) {
	for _, schema := range []string{"information_schema", "pg_catalog", "mysql", "sys"} {
		if !isSystemSchema(schema) {
			t.Fatalf("expected %q to be treated as a system schema", schema)
		}
	}

	if isSystemSchema("public") {
		t.Fatal("expected public to remain a normal schema")
	}
}
