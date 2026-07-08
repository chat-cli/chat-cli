# NFR Design — Unit 4 (Native Document Input)

No new design patterns are introduced for this unit — see `../nfr-requirements/nfr-requirements.md`'s "Design Pattern Reused, Not Redesigned" section. `utils.ValidateLocalPath` (designed and documented in `aidlc-docs/construction/unit-2-tool-use/nfr-design/nfr-design-patterns.md`) is reused as-is for `--document`'s path resolution. The only new logic in this unit (`sanitizeDocumentName`) is a simple string-transformation function, not a design pattern requiring its own architectural treatment.
