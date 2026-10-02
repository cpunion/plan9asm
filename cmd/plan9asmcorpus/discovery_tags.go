package main

import (
	"bufio"
	"fmt"
	"go/build"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Only tags mentioned by the assembly and one Go file can affect whether
// that pair makes a Go package. A package may have thousands of unrelated
// tags across other files; none belongs in this pair's search space.
func discoveryPairTags(asmTags, goTags, customTags []string) []string {
	custom := make(map[string]bool, len(customTags))
	for _, tag := range customTags {
		custom[tag] = true
	}
	set := make(map[string]bool)
	for _, tag := range append(append([]string(nil), asmTags...), goTags...) {
		if custom[tag] {
			set[tag] = true
		}
	}
	tags := make([]string, 0, len(set))
	for tag := range set {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// MatchFile checks filenames and directives, but unlike ImportDir it does not
// exclude a Go file's implicit cgo requirement. Use original imports for that
// last check, including when OpenFile replays captured source offline.
func discoveryGoFileMatches(ctx build.Context, dir, goFile string) (bool, error) {
	matches, err := ctx.MatchFile(dir, goFile)
	if err != nil || !matches || ctx.CgoEnabled {
		return matches, err
	}
	file := filepath.Join(dir, goFile)
	open := func(name string) (io.ReadCloser, error) { return os.Open(name) }
	if ctx.OpenFile != nil {
		open = ctx.OpenFile
	}
	reader, err := open(file)
	if err != nil {
		return false, err
	}
	const maxGoSelectionBytes = 8 << 20
	data, readErr := io.ReadAll(io.LimitReader(reader, maxGoSelectionBytes+1))
	closeErr := reader.Close()
	if readErr != nil {
		return false, readErr
	}
	if closeErr != nil {
		return false, closeErr
	}
	if len(data) > maxGoSelectionBytes {
		return false, fmt.Errorf("Go import-selection input exceeds %d bytes: %s", maxGoSelectionBytes, file)
	}
	parsed, err := parser.ParseFile(token.NewFileSet(), file, data, parser.ImportsOnly)
	if err != nil {
		return false, fmt.Errorf("parse original Go imports %s: %w", file, err)
	}
	for _, spec := range parsed.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return false, fmt.Errorf("invalid original Go import %s: %w", file, err)
		}
		if importPath == "C" {
			return false, nil
		}
	}
	return true, nil
}

// Preserve the original prefix through its last complete import declaration.
// A package-only witness cannot replay ImportDir's implicit cgo exclusion.
func discoveryGoSelectionHeader(file string, data []byte) (string, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), file, data, parser.ImportsOnly)
	if err != nil {
		return "", fmt.Errorf("capture original Go import selection %s: %w", file, err)
	}
	end := parsed.Name.End()
	for _, declaration := range parsed.Decls {
		if declaration.End() > end {
			end = declaration.End()
		}
	}
	return string(data[:int(end)-1]) + "\n", nil
}

func discoveryTagSetMatches(ctx build.Context, dir, asmFile, goFile string, tags []string) (bool, error) {
	ctx.BuildTags = tags
	asmMatches, err := ctx.MatchFile(dir, asmFile)
	if err != nil {
		return false, fmt.Errorf("match assembly for %s/%s with tags %v: %w", ctx.GOOS, ctx.GOARCH, tags, err)
	}
	if !asmMatches {
		return false, nil
	}
	goMatches, err := discoveryGoFileMatches(ctx, dir, goFile)
	if err != nil {
		return false, fmt.Errorf("match Go file %s for %s/%s with tags %v: %w", goFile, ctx.GOOS, ctx.GOARCH, tags, err)
	}
	return goMatches, nil
}

func discoveryTagsLexicallyBefore(left, right []string) bool {
	for i := range left {
		if left[i] == right[i] {
			continue
		}
		return left[i] < right[i]
	}
	return false
}

func findDiscoveryBuildTags(
	ctx build.Context, dir, asmFile string, goFiles, customTags []string,
	constraintTagsByFile map[string][]string,
) ([]string, bool, error) {
	ctx.BuildTags = nil
	asmMatches, err := ctx.MatchFile(dir, asmFile)
	if err != nil {
		return nil, false, fmt.Errorf("match assembly for %s/%s without custom tags: %w", ctx.GOOS, ctx.GOARCH, err)
	}
	if asmMatches {
		for _, goFile := range goFiles {
			matches, err := discoveryGoFileMatches(ctx, dir, goFile)
			if err != nil {
				return nil, false, fmt.Errorf("match Go file %s for %s/%s without custom tags: %w", goFile, ctx.GOOS, ctx.GOARCH, err)
			}
			if matches {
				return nil, true, nil
			}
		}
	}

	asmTags := constraintTagsByFile[filepath.Join(dir, asmFile)]
	var best []string
	for _, goFile := range goFiles {
		goTags := constraintTagsByFile[filepath.Join(dir, goFile)]
		pairTags := discoveryPairTags(asmTags, goTags, customTags)
		if len(pairTags) == 0 {
			continue
		}
		maxCount := len(pairTags)
		if len(best) != 0 && len(best) < maxCount {
			maxCount = len(best)
		}
		candidate, ok, err := solveDiscoveryPairTags(
			ctx, dir, asmFile, goFile, pairTags, maxCount,
		)
		if err != nil {
			return nil, false, err
		}
		if ok && (len(best) == 0 || len(candidate) < len(best) ||
			len(candidate) == len(best) && discoveryTagsLexicallyBefore(candidate, best)) {
			best = candidate
		}
	}
	if len(best) == 0 {
		return nil, false, nil
	}
	return best, true, nil
}

func solveDiscoveryPairTags(
	ctx build.Context, dir, asmFile, goFile string,
	tags []string, maxCount int,
) ([]string, bool, error) {
	var asmExpr, goExpr constraint.Expr
	if len(tags) > 16 {
		var err error
		asmExpr, err = discoveryFileBuildExpression(filepath.Join(dir, asmFile), ctx.OpenFile)
		if err != nil {
			return nil, false, err
		}
		goExpr, err = discoveryFileBuildExpression(filepath.Join(dir, goFile), ctx.OpenFile)
		if err != nil {
			return nil, false, err
		}
	}

	// These are fixed by the corpus context, not variables in the search.
	// In particular, an ignored generator file must be pruned before a
	// wide disjunction of otherwise selectable custom tags is enumerated.
	assignments := map[string]bool{
		"ignore": false,
		"cgo":    ctx.CgoEnabled,
		"gc":     ctx.Compiler == "gc",
		"gccgo":  ctx.Compiler == "gccgo",
	}
	selected := make([]string, 0, len(tags))
	var search func(index, remaining int) (bool, error)
	search = func(index, remaining int) (bool, error) {
		if remaining < 0 || remaining > len(tags)-index {
			return false, nil
		}
		if !discoveryExpressionMayMatch(asmExpr, assignments) ||
			!discoveryExpressionMayMatch(goExpr, assignments) {
			return false, nil
		}
		if index == len(tags) {
			if remaining != 0 {
				return false, nil
			}
			return discoveryTagSetMatches(ctx, dir, asmFile, goFile, selected)
		}

		tag := tags[index]
		if remaining > 0 {
			assignments[tag] = true
			selected = append(selected, tag)
			found, err := search(index+1, remaining-1)
			if found || err != nil {
				return found, err
			}
			selected = selected[:len(selected)-1]
		}
		assignments[tag] = false
		found, err := search(index+1, remaining)
		delete(assignments, tag)
		return found, err
	}

	for wantedCount := 1; wantedCount <= maxCount; wantedCount++ {
		found, err := search(0, wantedCount)
		if err != nil {
			return nil, false, err
		}
		if found {
			return append([]string(nil), selected...), true, nil
		}
	}
	return nil, false, nil
}

// A partial constraint evaluation overapproximates both possible outcomes.
// Returning false only when no completion can satisfy the expression keeps
// the solver exact even for repeated or negated tags.
func discoveryExpressionMayMatch(expr constraint.Expr, assignments map[string]bool) bool {
	mayTrue, _ := discoveryExpressionOutcomes(expr, assignments)
	return mayTrue
}

func discoveryExpressionOutcomes(expr constraint.Expr, assignments map[string]bool) (bool, bool) {
	switch expr := expr.(type) {
	case nil:
		return true, false
	case *constraint.TagExpr:
		value, assigned := assignments[expr.Tag]
		if !assigned {
			return true, true
		}
		return value, !value
	case *constraint.NotExpr:
		mayTrue, mayFalse := discoveryExpressionOutcomes(expr.X, assignments)
		return mayFalse, mayTrue
	case *constraint.AndExpr:
		leftTrue, leftFalse := discoveryExpressionOutcomes(expr.X, assignments)
		rightTrue, rightFalse := discoveryExpressionOutcomes(expr.Y, assignments)
		return leftTrue && rightTrue, leftFalse || rightFalse
	case *constraint.OrExpr:
		leftTrue, leftFalse := discoveryExpressionOutcomes(expr.X, assignments)
		rightTrue, rightFalse := discoveryExpressionOutcomes(expr.Y, assignments)
		return leftTrue || rightTrue, leftFalse && rightFalse
	default:
		return true, true
	}
}

func discoveryFileBuildExpression(filePath string, openers ...func(string) (io.ReadCloser, error)) (constraint.Expr, error) {
	open := func(file string) (io.ReadCloser, error) { return os.Open(file) }
	if len(openers) != 0 && openers[0] != nil {
		open = openers[0]
	}
	file, err := open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open build constraints %s: %w", filePath, err)
	}
	defer file.Close()

	var goBuild, legacy constraint.Expr
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	inBlockComment := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
			}
			continue
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "/*") {
			inBlockComment = !strings.Contains(line, "*/")
			continue
		}
		if !strings.HasPrefix(line, "//") {
			break
		}
		goBuildLine := constraint.IsGoBuild(line)
		if !goBuildLine && !constraint.IsPlusBuild(line) {
			continue
		}
		parsed, err := constraint.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("parse build constraint %s: %w", filePath, err)
		}
		if goBuildLine {
			goBuild = parsed
		} else if legacy == nil {
			legacy = parsed
		} else {
			legacy = &constraint.AndExpr{X: legacy, Y: parsed}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read build constraints %s: %w", filePath, err)
	}
	if goBuild != nil {
		return goBuild, nil
	}
	return legacy, nil
}
