//go:build go1.22

package plan9asm

import (
	"errors"
	"go/token"
	"go/types"
	"reflect"
	"testing"
)

func amd64ArgumentAlias(typ types.Type, count int) types.Type {
	for index := 0; index < count; index++ {
		typ = types.NewAlias(types.NewTypeName(token.NoPos, nil, "Alias", nil), typ)
	}
	return typ
}

func amd64ArgumentAliasSignature(typ types.Type) *types.Signature {
	return types.NewSignatureType(nil, nil, nil,
		types.NewTuple(types.NewVar(token.NoPos, nil, "value", typ)), nil, false)
}

func TestAMD64GoArgumentPlanAliasesPreserveConcreteIdentity(t *testing.T) {
	for _, count := range []int{1, 2, 32} {
		typ := amd64ArgumentAlias(types.Typ[types.Uint32], count)
		plan, err := planAMD64GoArguments(amd64ArgumentAliasSignature(typ))
		if err != nil {
			t.Fatal(err)
		}
		argument := plan.Arguments[0]
		if argument.Type != typ || argument.Stack || len(argument.Carriers) != 1 ||
			argument.Carriers[0].Type != I32 || argument.Carriers[0].Width != 4 || argument.Carriers[0].Register != AX {
			t.Fatalf("concrete alias/group identity was lost at depth %d: %#v", count, argument)
		}
	}
	boolean := amd64ArgumentAlias(types.Typ[types.Bool], 2)
	floating := amd64ArgumentAlias(types.Typ[types.Float32], 1)
	aggregate := amd64ArgumentAlias(types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "B", boolean, false),
		types.NewField(token.NoPos, nil, "F", floating, false),
	}, nil), 2)
	plan, err := planAMD64GoArguments(amd64ArgumentAliasSignature(aggregate))
	if err != nil {
		t.Fatal(err)
	}
	want := []amd64GoArgumentCarrier{
		{Path: []int{0}, Offset: 0, Width: 1, Type: I1, Kind: amd64GoArgumentBoolean, Register: AX},
		{Path: []int{1}, Offset: 4, Width: 4, Type: "float", Kind: amd64GoArgumentFloating, Register: "X0"},
	}
	if plan.Arguments[0].Type != aggregate || !reflect.DeepEqual(plan.Arguments[0].Carriers, want) {
		t.Fatalf("alias descendants changed their concrete byte kinds: %#v", plan)
	}
}

func TestAMD64GoArgumentPlanAliasesRejectUnresolvedAndUnboundedRHS(t *testing.T) {
	parameter := types.NewTypeParam(types.NewTypeName(token.NoPos, nil, "T", nil), types.NewInterfaceType(nil, nil).Complete())
	generic := types.NewNamed(types.NewTypeName(token.NoPos, nil, "Generic", nil), types.Typ[types.Uint64], nil)
	generic.SetTypeParams([]*types.TypeParam{types.NewTypeParam(types.NewTypeName(token.NoPos, nil, "U", nil), types.NewInterfaceType(nil, nil).Complete())})
	for _, test := range []struct {
		name string
		typ  types.Type
	}{
		{"type-parameter", amd64ArgumentAlias(parameter, 1)},
		{"alias-of-alias-type-parameter", amd64ArgumentAlias(parameter, 2)},
		{"generic-named-RHS", amd64ArgumentAlias(generic, 2)},
		{"alias-depth", amd64ArgumentAlias(types.Typ[types.Uint64], 33)},
	} {
		t.Run(test.name, func(t *testing.T) {
			plan, err := planAMD64GoArguments(amd64ArgumentAliasSignature(test.typ))
			if plan != nil || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("unsupported alias RHS acquired a partial plan: %#v, %v", plan, err)
			}
		})
	}
}
