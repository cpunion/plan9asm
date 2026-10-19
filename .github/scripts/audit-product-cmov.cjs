'use strict';

const fs = require('node:fs');
const path = require('node:path');
const {auditNativeOwners} = require('./audit-native-owners.cjs');

// Product and wide AES execution are portable LLVM, not native AVX512.
// CMOV additionally executes original Go opcode bytes. Keep all old programs.
const roots = [
  'TestCrossLinuxRuntimeMatrixX86DwordProductIndependentWholeBytes',
  'TestCrossLinuxRuntimeMatrixX86CMOVIndependentWholeBytes',
  'TestCrossLinuxRuntimeMatrixX86AESFamilyWholeByteAliases',
];
const children = ['-O0', '-O2', 'IR_O2/backend_O2'].map(axis => roots[0] + '/' + axis)
  .concat(['backend_O0', 'backend_O2', 'IR_O2/backend_O2'].map(axis => roots[1] + '/' + axis))
  .concat(['-O0', '-O2', 'IR_O2/backend_O2'].map(axis => roots[2] + '/' + axis));
const auditProductCMOV = (inventory, raw) => auditNativeOwners(inventory, raw, roots, children);

module.exports = {auditProductCMOV, roots, children};
if (require.main === module) {
  const directory = process.argv[2];
  const result = auditProductCMOV(fs.readFileSync(path.join(directory, 'inventory.log'), 'utf8'),
    fs.readFileSync(path.join(directory, 'raw.jsonl'), 'utf8'));
  process.stdout.write(JSON.stringify(result) + '\n');
}
