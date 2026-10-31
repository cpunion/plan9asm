package plan9asm

import (
 "encoding/hex"
 "fmt"
 "runtime"
 "strings"
 "testing"

 "github.com/xgo-dev/plan9asm/internal/goasmrules"
)

func TestX86LegacyScalarConversionOriginalGoBytesAndModes(t *testing.T) {
 catalog,err:=goasmrules.LoadX86LegacyRules(runtime.GOROOT());if err!=nil {t.Fatal(err)}
 type sample struct {name,table,prefix string;wide,fromFloat bool}
 samples:=[]sample{
  {"CVTSD2SL","yxcvfl","f20f2d",false,true},
  {"CVTSD2SQ","yxcvfq","f2480f2d",true,true},
  {"CVTSS2SL","yxcvfl","f30f2d",false,true},
  {"CVTSS2SQ","yxcvfq","f3480f2d",true,true},
  {"CVTTSD2SL","yxcvfl","f20f2c",false,true},
  {"CVTTSD2SQ","yxcvfq","f2480f2c",true,true},
  {"CVTTSS2SL","yxcvfl","f30f2c",false,true},
  {"CVTTSS2SQ","yxcvfq","f3480f2c",true,true},
  {"CVTSL2SD","yxcvlf","f20f2a",false,false},
  {"CVTSQ2SD","yxcvqf","f2480f2a",true,false},
  {"CVTSL2SS","yxcvlf","f30f2a",false,false},
  {"CVTSQ2SS","yxcvqf","f3480f2a",true,false},
 }
 entries:=map[string]goasmrules.X86LegacyOpcode{}
 for _,entry:=range catalog.Optab {for _,sample:=range samples {if entry.Opcode==sample.name {if _,dup:=entries[entry.Opcode];dup {t.Fatal("duplicate original table")};entries[entry.Opcode]=entry}}}
 if len(entries)!=12 {t.Fatal("missing complete original conversion family")}
 for _,arch:=range []string{"amd64","386"} {
  t.Run(arch,func(t *testing.T){
   var source strings.Builder
   var cases []x86OfficialPackedCase
   for _,sample:=range samples {
    entry:=entries[sample.name]
    if entry.Ytab!=sample.table||len(entry.Rows)!=1 {t.Fatal("original row ownership changed",entry)}
    operands:=[]string{"CX,X1","0(SI),X1"};modrm:=[]string{"c9","0e"}
    if sample.fromFloat {operands=[]string{"X0,BX","0(SI),BX"};modrm=[]string{"d8","1e"}}
    for index,args:=range operands {
     instruction:=sample.name+" "+args
     if arch=="386"&&sample.wide {
      t.Run("Q-refusal/"+instruction,func(t *testing.T){x86ScalarLiteralExpectedGoListing(t,arch,"TEXT ·NoQ(SB),4,$0-0\n"+instruction+"\nRET\n",false)})
      continue
     }
     symbol:=fmt.Sprintf("ScalarConversion%d",len(cases))
     fmt.Fprintf(&source,"TEXT ·%s(SB),4,$0-0\n",symbol)
     line:=strings.Count(source.String(),"\n")+1
     fmt.Fprintln(&source,instruction);source.WriteString("RET\n")
     encoding,err:=hex.DecodeString(sample.prefix+modrm[index]);if err!=nil {t.Fatal(err)}
     cases=append(cases,x86OfficialPackedCase{symbol:symbol,line:line,row:fmt.Sprintf("optab/%d/%s/%s/0",entry.TableIndex,entry.Opcode,entry.Ytab),witness:goasmrules.X86PackedBinaryWitness{Instruction:instruction,Encoding:encoding,Destination:strings.Split(args,",")[1]}})
    }
   }
   want:=24;if arch=="386" {want=12};if len(cases)!=want||len(cases)>32 {t.Fatal("whole-mode byte witness inventory changed",len(cases))}
   listing,err:=x86ScalarLiteralGoListing(t,arch,source.String());if err!=nil {t.Fatal(err)}
   if err:=x86OfficialPackedVerifyListing(listing,cases);err!=nil {t.Fatal("independent original conversion bytes",err)}
  })
 }
}
