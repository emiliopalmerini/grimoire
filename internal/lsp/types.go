package lsp

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type CodeAction struct {
	Title string         `json:"title"`
	Kind  string         `json:"kind"`
	Edit  *WorkspaceEdit `json:"edit,omitempty"`
}

type WorkspaceEdit struct {
	Changes         map[string][]TextEdit `json:"changes,omitempty"`
	DocumentChanges []TextDocumentEdit    `json:"documentChanges,omitempty"`
}

type TextDocumentEdit struct {
	TextDocument map[string]any `json:"textDocument"`
	Edits        []TextEdit     `json:"edits"`
}

type DocumentSymbol struct {
	Name    string
	Kind    string
	Line    int
	EndLine int
}

var symbolKindNames = map[int]string{
	1: "File", 2: "Module", 3: "Namespace", 4: "Package", 5: "Class",
	6: "Method", 7: "Property", 8: "Field", 9: "Constructor", 10: "Enum",
	11: "Interface", 12: "Function", 13: "Variable", 14: "Constant", 15: "String",
	16: "Number", 17: "Boolean", 18: "Array", 19: "Object", 20: "Key",
	21: "Null", 22: "EnumMember", 23: "Struct", 24: "Event", 25: "Operator",
	26: "TypeParameter",
}

type rawDocumentSymbol struct {
	Name     string              `json:"name"`
	Kind     int                 `json:"kind"`
	Range    Range               `json:"range"`
	Children []rawDocumentSymbol `json:"children"`
}

type rawSymbolInformation struct {
	Name     string `json:"name"`
	Kind     int    `json:"kind"`
	Location struct {
		Range Range `json:"range"`
	} `json:"location"`
}
