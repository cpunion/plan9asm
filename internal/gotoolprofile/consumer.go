package gotoolprofile

const ConsumerProtocol = "actual_go_cpu_profile_consumer_v1"

// ConsumerInput is a private invocation artifact, not a portable report. The
// source root binds an actual replacement/module directory; all selected bytes
// must match the producer's pre-load hashes. Only SelectionProof is published.
type ConsumerInput struct {
	Protocol     string              `json:"protocol"`
	ID           string              `json:"id"`
	Observed     *Observation        `json:"observed"`
	Module       string              `json:"module"`
	Version      string              `json:"version"`
	SourceModule string              `json:"source_module,omitempty"`
	SourceRoot   string              `json:"source_root"`
	Sources      map[string]string   `json:"sources"`
	Headers      map[string]string   `json:"selection_headers"`
	ToolSources  map[string]string   `json:"tool_sources,omitempty"`
	Directories  map[string][]string `json:"directories"`
	AsmFiles     []string            `json:"asm_files"`
}

type SelectionProof struct {
	Protocol   string         `json:"protocol"`
	ProfileID  string         `json:"profile_id"`
	CustomTags []string       `json:"custom_tags,omitempty"`
	Packages   []PackageProof `json:"packages"`
	CPP        []CPPProof     `json:"cpp"`
	Outputs    []OutputProof  `json:"outputs"`
}

type PackageProof struct {
	PackagePath     string            `json:"package_path"`
	ModulePath      string            `json:"module_path"`
	ModuleVersion   string            `json:"module_version"`
	SourceModule    string            `json:"source_module"`
	SourceVersion   string            `json:"source_version"`
	SourceRole      string            `json:"source_role"`
	GoFiles         []string          `json:"go_files"`
	CompiledGoFiles []string          `json:"compiled_go_files"`
	SFiles          []string          `json:"s_files"`
	SourceSHA256    map[string]string `json:"source_sha256"`
	Macros          *AssemblerMacros  `json:"macros"`
}

type CPPProof struct {
	File                string            `json:"file"`
	ExpandedSHA256      string            `json:"expanded_sha256"`
	TypedExpandedSHA256 string            `json:"typed_expanded_sha256"`
	Inputs              map[string]string `json:"inputs"`
}

type OutputProof struct {
	File   string `json:"file"`
	Part   string `json:"part"`
	IR     string `json:"ir_sha256"`
	Object string `json:"object_sha256,omitempty"`
}
