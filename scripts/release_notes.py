#!/usr/bin/env python3
"""Extract the exact release's curated changelog section; fail if missing."""
from pathlib import Path
import re

root=Path(__file__).resolve().parents[1]
version=(root/'VERSION').read_text().strip()
text=(root/'CHANGELOG.md').read_text()
match=re.search(r'^## \['+re.escape(version)+r'\][^\n]*\n(.*?)(?=^## |\Z)',text,re.M|re.S)
if not match or not match[1].strip():
    raise SystemExit('Missing CHANGELOG entry for '+version)
print(match[1].strip())
