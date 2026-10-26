#!/usr/bin/env python3
import glob
import sys
import xml.etree.ElementTree as ET

def main():
    files = glob.glob("assets/*.svg")
    if len(files) < 5:
        print(f"Error: expected at least 5 SVGs in assets/, found {len(files)}", file=sys.stderr)
        sys.exit(1)

    for path in sorted(files):
        # 1. Check well-formed XML
        try:
            tree = ET.parse(path)
            root = tree.getroot()
        except Exception as e:
            print(f"Error parsing XML for {path}: {e}", file=sys.stderr)
            sys.exit(1)

        # 2. Check no scripts
        content = open(path, "r", encoding="utf-8").read()
        if "<script" in content.lower():
            print(f"Error: forbidden <script> tag found in {path}", file=sys.stderr)
            sys.exit(1)

        # 3. Check no external references (http / https)
        if "http" in content.lower():
            print(f"Error: forbidden external http reference in {path}", file=sys.stderr)
            sys.exit(1)

        print(f"✓ {path}: valid XML, 0 scripts, 0 external references (viewBox={root.attrib.get('viewBox')})")

    print(f"\nAll {len(files)} SVG assets successfully validated!")

if __name__ == "__main__":
    main()
