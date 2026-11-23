# Mistakes ledger

Every entry is a rule earned the hard way. Claude reads this at the start of every session and
treats each rule as a hard constraint. Add an entry whenever a mistake is corrected or a root cause
is found. Newest at the top.

## Entry format

```
### L<number>: <the rule, as an instruction>
- Date: YYYY-MM-DD
- Project: <repo or "all">
- What happened: <1–3 lines, the actual incident>
- Why: <the root cause, not the symptom>
- How to apply: <what to do differently, concretely>
```

## Example (delete once you have real entries)

### L0: "Merged" is not "deployed": check the deploy run before saying a change is live
- Date: 2026-01-01
- Project: all
- What happened: a fix was merged and reported live, but the deploy job had silently been skipped.
- Why: the merge was treated as proof the change reached production.
- How to apply: after merging, open the CI run, confirm the deploy jobs ran (not skipped) and passed, then check the live site.

---

## Entries
