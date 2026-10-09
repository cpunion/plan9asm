"""Script delegation checks; historical loop RED remains in Git history.

The actual process/source/object regressions live in internal/goasmoracle and
cmd/plan9asmgooracle, so the script and root tests consume one classifier.
"""

import pathlib
import unittest


class ARM64Plan9OracleScriptTests(unittest.TestCase):
    def test_shared_owner_and_retained_reports(self):
        script = (pathlib.Path(__file__).resolve().parent / "check-arm64-plan9-corpus.sh").read_text()
        for expected in ('/cmd/plan9asmgooracle"', '-go="$go_path" -goroot="$goroot"',
                         'accepted_cases="$oracle_parent/go-oracle/accepted.txt"',
                         'full_report="$oracle_parent/full.json"',
                         'accepted_report="$oracle_parent/accepted.json"'):
            self.assertIn(expected, script)
        for forbidden in ("subprocess.DEVNULL", "subprocess.run", "else rejected"):
            self.assertNotIn(forbidden, script)


if __name__ == "__main__":
    unittest.main(verbosity=2)
