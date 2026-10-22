'use strict';

const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const {test} = require('node:test');

const source = fs.readFileSync('amd64_aes_family_fixed_local_test.go', 'utf8');

test('the complete AES bound-consumer root owns one disposable compiler cache', () => {
  assert.match(source, /func TestX86AESFamilyActualBoundFixedLocalConsumers\(t \*testing\.T\) \{\s*t\.Setenv\("GOCACHE", t\.TempDir\(\)\)/);
});

test('all AES targets consume independent original source leases for all three APIs', () => {
  assert.match(source, /fixedLocalFreshAPIs\(t, pkg, request, options, \[\]string\{"Translate", "Module", "GoModule"\}/);
  assert.doesNotMatch(source, /requireGoAMD64FrameCapture\(|runtimeOracleStageContext\(|\.BindFixedLocal\(|fixedLocalAPIs\(/);
});

test('AES retains exactly twelve original names, two source addresses, two targets and one object path', () => {
  const oracle = fs.readFileSync('amd64_aes_family_effects_test.go', 'utf8');
  const names = oracle.match(/var x86AESOracleNames = \[\]string\{([^}]+)\}/);
  assert.ok(names);
  assert.deepEqual([...names[1].matchAll(/"([^"]+)"/g)].map(match => match[1]),
    ['AESENC', 'AESENCLAST', 'AESDEC', 'AESDECLAST', 'AESIMC', 'AESKEYGENASSIST',
      'VAESENC', 'VAESENCLAST', 'VAESDEC', 'VAESDECLAST', 'VAESIMC', 'VAESKEYGENASSIST']);
  const builder = source.slice(source.indexOf('\tvar assembly, declarations'), source.indexOf('\tfor _, target :='));
  assert.equal(crypto.createHash('sha256').update(builder).digest('hex'),
    'ccc5fe309b626943f3f92ef9ecc847e3f009aa66521eb395a4adcc6b6a63ace9');
  assert.match(builder, /\[\]string\{"16\(SP\)", "tmp-144\(SP\)"\}/);
  assert.deepEqual([...source.matchAll(/\{"([^"\n]+\/amd64)", "([^"\n]+)"\}/g)]
    .map(match => match.slice(1)), [['linux/amd64', 'x86_64-unknown-linux-gnu'], ['darwin/amd64', 'x86_64-apple-darwin']]);
  assert.equal((source.match(/compileLLVMToObject\(/g) || []).length, 1);
  assert.match(source, /compileLLVMToObject\(t, llc, target\.triple, "aes-native-local\.ll", "aes-native-local\.o", ir\)/);
  assert.doesNotMatch(source, /x86MulDivRoundOptimize|"-O[0123]"|t\.Skip|t\.Parallel/);
});
