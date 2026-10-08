//go:build go1.27

package gostaticstorage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// This reader uses the selected SDK's binary readers and PCToLine. It does not
// call the production parser or derive a source read/native-entry permission.
const sdkTextReaderSource = `package main
import (
 "cmd/internal/archive"
 "cmd/internal/goobj"
 "cmd/internal/objfile"
 "encoding/json"
 "fmt"
 "os"
)
type Ref struct {
 PackageIndex, SymbolIndex uint32
 Package, Name string
 ABI uint16
}
type Reloc struct { Offset, Size int64; Type uint16; Add int64; Target Ref }
type Aux struct { Type uint8; Target Ref }
type Position struct { Start, End int64; File string; Line int32 }
type Text struct {
 Index uint32; Name string; ABI uint16; Flags, Flags2 uint8
 Code []byte; Relocations []Reloc; Auxiliary []Aux; Positions []Position
}
type Member struct { Name string; Offset, Size int64; Flags uint32; Texts []Text }
func main() {
 f,err:=os.Open(os.Args[1]); if err!=nil { panic(err) }; defer f.Close()
 a,err:=archive.Parse(f,false); if err!=nil { panic(err) }
 o,err:=objfile.Open(os.Args[1]); if err!=nil { panic(err) }; defer o.Close()
 entries:=map[string]*objfile.Entry{}
 for _,e:=range o.Entries() { if entries[e.Name()]!=nil { panic("duplicate entry") }; entries[e.Name()]=e }
 result:=[]Member{}
 for _,e:=range a.Entries {
  m:=Member{Name:e.Name,Offset:e.Offset,Size:e.Size}
  if e.Type!=archive.EntryPkgDef && e.Type!=archive.EntryGoObj { panic("unknown member") }
  if e.Type==archive.EntryGoObj {
   data:=make([]byte,e.Obj.Size)
   n,err:=f.ReadAt(data,e.Obj.Offset); if err!=nil||n!=len(data) { panic(fmt.Sprint(n,err)) }
   r:=goobj.NewReaderFromBytes(data,false); if r==nil { panic("goobj") }; m.Flags=r.Flags()
   refs:=map[goobj.SymRef]string{}
   for i:=0;i<r.NRefName();i++ { q:=r.RefName(i); refs[q.Sym()]=q.Name(r) }
   resolve:=func(q goobj.SymRef) Ref {
    v:=Ref{PackageIndex:q.PkgIdx,SymbolIndex:q.SymIdx}; var index uint32
    switch q.PkgIdx {
    case goobj.PkgIdxInvalid: return v
    case goobj.PkgIdxSelf: index=q.SymIdx
    case goobj.PkgIdxHashed64: index=q.SymIdx+uint32(r.NSym())
    case goobj.PkgIdxHashed: index=q.SymIdx+uint32(r.NSym()+r.NHashed64def())
    case goobj.PkgIdxNone: index=q.SymIdx+uint32(r.NSym()+r.NHashed64def()+r.NHasheddef())
    case goobj.PkgIdxBuiltin:
     name,abi:=goobj.BuiltinName(int(q.SymIdx)); v.Name=name; v.ABI=uint16(abi); return v
    default: v.Package=r.Pkg(int(q.PkgIdx)); v.Name=refs[q]; return v
    }
    s:=r.Sym(index); v.Name=s.Name(r); v.ABI=s.ABI(); return v
   }
   entry:=entries[e.Name]; if entry==nil { panic("missing objfile entry") }
   liner,err:=entry.PCLineTable(); if err!=nil { panic(err) }
   count:=r.NSym()+r.NHashed64def()+r.NHasheddef()+r.NNonpkgdef()
   for i:=0;i<count;i++ {
    s:=r.Sym(uint32(i)); if s.Type()!=1 { continue }
    v:=Text{Index:uint32(i),Name:s.Name(r),ABI:s.ABI(),Flags:s.Flag(),Flags2:s.Flag2(),Code:r.Data(uint32(i))}
    for j:=0;j<r.NReloc(uint32(i));j++ {
     q:=r.Reloc(uint32(i),j)
     v.Relocations=append(v.Relocations,Reloc{int64(q.Off()),int64(q.Siz()),q.Type(),q.Add(),resolve(q.Sym())})
    }
    for j:=0;j<r.NAux(uint32(i));j++ { q:=r.Aux(uint32(i),j); v.Auxiliary=append(v.Auxiliary,Aux{q.Type(),resolve(q.Sym())}) }
    for pc:=int64(0);pc<int64(s.Siz());pc++ {
     file,line,_:=liner.PCToLine(uint64(r.DataOff(uint32(i)))+uint64(pc))
     k:=len(v.Positions)-1
     if k>=0 && v.Positions[k].File==file && v.Positions[k].Line==int32(line) { v.Positions[k].End=pc+1
     } else { v.Positions=append(v.Positions,Position{pc,pc+1,file,int32(line)}) }
    }
    m.Texts=append(m.Texts,v)
   }
  }
  result=append(result,m)
 }
 if err:=json.NewEncoder(os.Stdout).Encode(result);err!=nil { panic(err) }
}
`

func TestGoStaticStorageTEXTActualSourceReport(t *testing.T) {
	dir, binary, _ := fixture(t)
	oracle := officialReaderForSource(t, binary, sdkTextReaderSource)
	for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
		t.Run(target, func(t *testing.T) {
			_, report := captured(t, request(t, dir, binary, target))
			encoded, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			var envelope map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &envelope); err != nil {
				t.Fatal(err)
			}
			var recorded map[string]json.RawMessage
			if err := json.Unmarshal(envelope["TextArchives"], &recorded); err != nil {
				// Keep observing the real original object before reporting RED.
				recorded = nil
			}
			packages, err := reportPackages(report, dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{fixtureModule, fixtureModule + "/dep"} {
				raw := run(t, oracle, dir, fixtureEnv(t, target), packages[path].Export)
				var original []map[string]json.RawMessage
				if err := json.Unmarshal(raw, &original); err != nil || len(original) != 2 {
					t.Fatalf("actual original archive oracle failed: %v", err)
				}
				if artifact := os.Getenv("PLAN9ASM_STATIC_STORAGE_ARTIFACT_DIR"); artifact != "" {
					folder := filepath.Join(artifact, strings.ReplaceAll(target, "/", "-"), strings.TrimPrefix(path, fixtureModule))
					if err := os.MkdirAll(folder, 0700); err != nil {
						t.Fatal(err)
					}
					for name, data := range map[string][]byte{"text-sdk.json": raw, "text-report.json": encoded, "text-original.a": report.Archives[path]} {
						if err := os.WriteFile(filepath.Join(folder, name), data, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
				t.Logf("ORIGINAL_TEXT %s %s archiveSHA256=%s members=%d", target, path, report.ArchiveSHA256[path], len(original))
				var actual any
				if err := json.Unmarshal(raw, &actual); err != nil {
					t.Fatal(err)
				}
				var claimed any
				if err := json.Unmarshal(recorded[path], &claimed); err != nil || !reflect.DeepEqual(actual, claimed) {
					t.Errorf("actual selected source/archive lacks complete TEXT relocation Add/Sym/Aux/source-PC readback: %s", path)
				}
			}
		})
	}
}
