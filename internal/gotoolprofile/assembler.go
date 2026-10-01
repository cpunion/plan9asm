package gotoolprofile

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xgo-dev/plan9asm"
)

type AssemblerMacros struct {
	Protocol           string            `json:"protocol"`
	FeatureID          string            `json:"feature_id"`
	Target             string            `json:"target"`
	GoVersion          string            `json:"go_version"`
	PackagePath        string            `json:"package_path"`
	PackageRole        string            `json:"package_role"`
	Defines            []string          `json:"defines"`
	ToolSourceSHA256   map[string]string `json:"tool_source_sha256"`
	RegistrationSHA256 map[string]string `json:"registration_ast_sha256"`
}

// This derives package-role macros from actual registered Go sources. It is
// not a package source-selection observation: production must separately bind
// packagePath to the actual go list package being checked under this profile.
func CaptureAssemblerMacros(root string, observed *Observation, packagePath string) (*AssemblerMacros, error) {
	if observed == nil || observed.Protocol != "go_driver_builtin_features_v2" || observed.GoVersion != observed.Environment["GOVERSION"] {
		return nil, fmt.Errorf("missing or inconsistent observed assembler environment")
	}
	if !filepath.IsAbs(root) || packagePath == "" || path.Clean(packagePath) != packagePath || strings.HasPrefix(packagePath, "/") || strings.ContainsAny(packagePath, "@ \\\t\r\n\"'") {
		return nil, fmt.Errorf("invalid actual assembler source root or package path")
	}
	minor, err := goMinor(observed.GoVersion)
	if err != nil || minor < 20 || minor > 27 {
		return nil, fmt.Errorf("assembler registration is not audited for %s", observed.GoVersion)
	}
	defines, err := plan9asm.GoAssemblerDefinesForEnvironment(observed.Environment["GOOS"], observed.Environment["GOARCH"], observed.Environment)
	if err != nil {
		return nil, err
	}
	proof := &AssemblerMacros{Protocol: "go_source_assembler_macros_v1", FeatureID: ProfileID(observed), Target: observed.Target, GoVersion: observed.GoVersion, PackagePath: packagePath, PackageRole: "ordinary_path", Defines: defines, ToolSourceSHA256: make(map[string]string), RegistrationSHA256: make(map[string]string)}
	version, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil || strings.TrimSpace(strings.Split(string(version), "\n")[0]) != observed.GoVersion {
		return nil, fmt.Errorf("assembler source VERSION differs from actual driver")
	}
	proof.ToolSourceSHA256["VERSION"] = bytesSHA256(version)
	readSource := func(name, pkg string) (*ast.File, *token.FileSet, error) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return nil, nil, err
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, name, data, 0)
		if err != nil || file.Name.Name != pkg {
			return nil, nil, fmt.Errorf("invalid assembler registration source %s", name)
		}
		proof.ToolSourceSHA256[name] = bytesSHA256(data)
		return file, fset, nil
	}
	gc, gcSet, err := readSource("src/cmd/go/internal/work/gc.go", "work")
	if err != nil {
		return nil, err
	}
	args, err := discoveryAssemblerSourceFunction(gc, "asmArgs")
	if err != nil {
		return nil, err
	}
	fingerprint, err := ASTSHA(gcSet, args)
	if err != nil || fingerprint != discoveryAssemblerArgsFingerprints[minor] {
		return nil, fmt.Errorf("unrecognized Go 1.%d cmd/go asmArgs registration", minor)
	}
	proof.RegistrationSHA256["asmArgs"] = fingerprint
	asm, asmSet, err := readSource("src/cmd/asm/main.go", "main")
	if err != nil {
		return nil, err
	}
	main, err := discoveryAssemblerSourceFunction(asm, "main")
	if err != nil {
		return nil, err
	}
	// Find every registration literal, not merely a guessed function name.
	// A new registration elsewhere or a changed condition must fail closed.
	literalCount := 0
	ast.Inspect(asm, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if ok && literal.Kind == token.STRING {
			value, _ := strconv.Unquote(literal.Value)
			if strings.HasPrefix(value, "GOEXPERIMENT_") {
				literalCount++
			}
		}
		return true
	})
	var blocks []ast.Stmt
	for _, statement := range main.Body.List {
		var formatted bytes.Buffer
		if err := format.Node(&formatted, asmSet, statement); err != nil {
			return nil, err
		}
		if strings.Contains(formatted.String(), "GOEXPERIMENT_") {
			blocks = append(blocks, statement)
		}
	}
	if minor < 22 {
		if literalCount != 0 || len(blocks) != 0 {
			return nil, fmt.Errorf("unexpected pre-Go1.22 experiment macro registration")
		}
		proof.RegistrationSHA256["experiment_registration_absent"] = bytesSHA256(nil)
		proof.PackageRole = "legacy_no_experiment_macros"
		return proof, nil
	}
	if literalCount != 1 || len(blocks) != 1 {
		return nil, fmt.Errorf("missing or additional experiment macro registrations")
	}
	fingerprint, err = ASTSHA(asmSet, blocks[0])
	if err != nil || fingerprint != "32995e3cdb0c0ab1499abc494906e85ec47b639f2ebeac6b4c049d6d676a8b0f" {
		return nil, fmt.Errorf("unrecognized package-role experiment macro registration")
	}
	proof.RegistrationSHA256["experiment_registration"] = fingerprint
	special, specialSet, err := readSource("src/cmd/internal/objabi/pkgspecial.go", "objabi")
	if err != nil {
		return nil, err
	}
	fingerprint, err = ASTSHA(specialSet, special)
	if err != nil || fingerprint != discoveryAssemblerSpecialFingerprints[minor] {
		return nil, fmt.Errorf("unrecognized Go 1.%d package-special registration", minor)
	}
	proof.RegistrationSHA256["package_special_registration"] = fingerprint
	paths, err := discoveryAssemblerAllowedPackagePaths(special)
	if err != nil {
		return nil, err
	}
	if paths[packagePath] {
		proof.PackageRole = "allow_asm_abi_path"
		for _, tag := range observed.ToolTags {
			if strings.HasPrefix(tag, "goexperiment.") {
				if !observed.MarkerSelection[tag] {
					return nil, fmt.Errorf("unobserved enabled experiment %s", tag)
				}
				proof.Defines = append(proof.Defines, "GOEXPERIMENT_"+strings.TrimPrefix(tag, "goexperiment."))
			}
		}
	}
	return proof, nil
}

func discoveryAssemblerSourceFunction(file *ast.File, name string) (*ast.FuncDecl, error) {
	var found *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == name {
			if found != nil || function.Body == nil || function.Recv != nil {
				return nil, fmt.Errorf("invalid assembler registration function %s", name)
			}
			found = function
		}
	}
	if found == nil {
		return nil, fmt.Errorf("missing assembler registration function %s", name)
	}
	return found, nil
}

func ASTSHA(fset *token.FileSet, node ast.Node) (string, error) {
	var formatted bytes.Buffer
	if err := format.Node(&formatted, fset, node); err != nil {
		return "", err
	}
	var scan scanner.Scanner
	scan.Init(token.NewFileSet().AddFile("registration", -1, formatted.Len()), formatted.Bytes(), nil, 0)
	var canonical bytes.Buffer
	for {
		_, tok, literal := scan.Scan()
		if tok == token.EOF {
			break
		}
		fmt.Fprintf(&canonical, "%s\x00%s\x00", tok, literal)
	}
	return bytesSHA256(canonical.Bytes()), nil
}

func discoveryAssemblerAllowedPackagePaths(file *ast.File) (map[string]bool, error) {
	paths := make(map[string]bool)
	registrations := 0
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			value := spec.(*ast.ValueSpec)
			if len(value.Names) == 0 || value.Names[0].Name != "allowAsmABIPkgs" {
				continue
			}
			registrations++
			if len(value.Names) != 1 || len(value.Values) != 1 {
				return nil, fmt.Errorf("unrecognized package-special path registration")
			}
			list, ok := value.Values[0].(*ast.CompositeLit)
			if !ok {
				return nil, fmt.Errorf("nonliteral package-special path registration")
			}
			for _, element := range list.Elts {
				literal, ok := element.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					return nil, fmt.Errorf("nonliteral package-special path")
				}
				pkg, err := strconv.Unquote(literal.Value)
				if err != nil || pkg == "" || paths[pkg] {
					return nil, fmt.Errorf("invalid or duplicate package-special path")
				}
				paths[pkg] = true
			}
		}
	}
	if registrations != 1 || len(paths) == 0 {
		return nil, fmt.Errorf("incomplete package-special path inventory")
	}
	return paths, nil
}

// Complete normalized registration ASTs from the eight already audited actual
// Go sources. These are source grammar guards, not lowerer coverage baselines.
var discoveryAssemblerArgsFingerprints = map[int]string{
	20: "bf67bdbe3e3c20c355bf387525aa895bee95ba89937c0d2c3ff97aa3c39676b9",
	21: "bf67bdbe3e3c20c355bf387525aa895bee95ba89937c0d2c3ff97aa3c39676b9",
	22: "bfb3ca150175f463b517ae6b7d0b6a5a3399844258429f9e4a70e5d3544313e5",
	23: "2a986d99ca5b98d9e6d2b1bc904b03c8f5776192010cfe5a0422875c959851c2",
	24: "2a986d99ca5b98d9e6d2b1bc904b03c8f5776192010cfe5a0422875c959851c2",
	25: "2a986d99ca5b98d9e6d2b1bc904b03c8f5776192010cfe5a0422875c959851c2",
	26: "2a986d99ca5b98d9e6d2b1bc904b03c8f5776192010cfe5a0422875c959851c2",
	27: "4586887594337565ace2e18002673943804b443b3f70391842e933f3af557c8d",
}

var discoveryAssemblerSpecialFingerprints = map[int]string{
	22: "c49dc93a7426aa8657d56ea5e6aebd86d4798b3d27c7b27ee04422784f709198",
	23: "fb45940df217866cbba98f749855c2ea93fd95e7f4b8c5c0e284841e762b3078",
	24: "1aa641469c37d6ad633e20cf6750eac1f4809aad749a5e60e6ec0304a0f131b0",
	25: "e2bb3c22dee594fc9278db7fe3c4eadd28c7f005bacba42bdb81cd52eb5babcc",
	26: "4a67f46838d103efc652cea4670309c68eff5f07332fbb17df5f6fe3aa4161a9",
	27: "488cab6435fd67db5c35939190ad5e3be6ea7a4cdd2b3e3a24c6993cea466c0f",
}
