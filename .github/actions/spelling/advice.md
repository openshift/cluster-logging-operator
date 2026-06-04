### Spell check advice for contributors

If the spell checker flags a word that is correctly spelled:

1. **Project-specific term** (API type name, product name): Add it to `.github/actions/spelling/allow.txt`
2. **Technical jargon** that may come and go: Add it to `.github/actions/spelling/expect.txt`
3. **A pattern** (URLs, code, encoded data): Add a regex to `.github/actions/spelling/patterns.txt`

If the spell checker flags a word in a **generated file** under `docs/reference/`:
- Fix the typo in the Go source code comment, then regenerate with `make docs`
- Do NOT add the misspelling to `allow.txt` or `expect.txt`

To re-run the spell check after updating configuration files, push a new commit or re-run the workflow from the Actions tab.
