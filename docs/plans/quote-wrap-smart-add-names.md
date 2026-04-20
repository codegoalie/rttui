# Plan: Quote-wrap list/tag names in Smart Add transform

## Problem

RTM's Smart Add parser is observed to consume text greedily after `#` — the tag/list name continues through whitespace until it hits another Smart Add symbol (`!`, `^`, `*`, `~`, `=`, `+`, `@`, `//`) or end of string.

With `add_preset = "#PrizePicks"` placed at the start of the input (pre-bc215a4 behavior), inputs like:

    #PrizePicks Review oddsfather 233 !1 ^today

were parsed by RTM as tag name `PrizePicks Review oddsfather 233` (stops at `!1`), producing junk tags like `prizepicksreviewoddsfather233` and leaving the literal Smart Add tokens in the title.

Commit `bc215a4` moved the preset to the end, which mitigates the issue for this specific flow since `#PrizePicks` is now terminated by end-of-string. But the input is still fragile: any free-typed `#tag` followed by additional task text re-introduces the same greedy-consumption bug.

## Proposed fix

In `internal/ui/smartinput.go`, harden `transformForRTM` so every list and tag token is emitted as `#"Name"` (double-quoted) before being sent to RTM. This pins the name boundary unambiguously, regardless of what follows.

Transformations:
- `#PrizePicks` → `#"PrizePicks"`
- `%errands`   → `#"errands"`  (existing `%` → `#` conversion preserved)

Other token kinds (`!`, `^`, `*`) are unchanged.

## Prerequisite: confirm RTM supports quoted names

Before landing, empirically verify that RTM's Smart Add accepts `#"Name"` and treats it as a single name. No public docs confirm or deny this syntax. Two verification paths:

1. Build a tiny throwaway script that calls `rtm.tasks.add` with `parse=1` and `name=#"PrizePicks"` — inspect the resulting task via `rtm.tasks.getList` or the web UI.
2. Add a task manually in the RTM web UI's Smart Add field with a quoted name and observe.

If quotes are NOT supported, fall back to one of:
- Send `list_id` explicitly via `rtm.tasks.add`'s `list_id` parameter instead of encoding the list in the Smart Add string. Requires resolving list name → id (one `rtm.lists.getList` call, cached).
- Strip Smart Add parsing entirely for the list token and use `list_id` for that case only.

## Implementation

### 1. `internal/ui/smartinput.go`

Update `transformForRTM` (currently lines 100–112):

```go
func transformForRTM(input string) string {
    toks := tokenize([]rune(input))
    var sb strings.Builder
    for _, tok := range toks {
        switch tok.kind {
        case tokenList:
            if len(tok.raw) > 1 {
                sb.WriteString(`#"`)
                sb.WriteString(tok.raw[1:])
                sb.WriteByte('"')
                continue
            }
        case tokenTag:
            if len(tok.raw) > 1 {
                sb.WriteString(`#"`)
                sb.WriteString(tok.raw[1:])
                sb.WriteByte('"')
                continue
            }
        }
        sb.WriteString(tok.raw)
    }
    return sb.String()
}
```

Edge cases:
- Bare `#` or `%` (no name): pass through unchanged so the user still sees what they typed when the submit fails validation on RTM's side.
- Names containing a literal `"`: not supported today; not a concern for the immediate use case. Document as a known limitation.

### 2. Tests

Add `internal/ui/smartinput_test.go` (new file — there are no existing UI tests). Table-driven cases:

| input | expected RTM string |
|---|---|
| `Review oddsfather 233 !1 ^today #PrizePicks` | `Review oddsfather 233 !1 ^today #"PrizePicks"` |
| `Buy milk %errands` | `Buy milk #"errands"` |
| `#Work !2 %urgent ^friday *weekly` | `#"Work" !2 #"urgent" ^friday *weekly` |
| `Plain task no markers` | `Plain task no markers` |
| `Edge case with # alone` | `Edge case with # alone` |
| `Multi-word #list followed by text` | `Multi-word #"list" followed by text` |

Keep the test package as `ui` (internal) so `transformForRTM` stays unexported.

### 3. Docs

Update `README.md`:
- Remove or soften the line explaining the `%` → `#` rewrite; the mechanics are now "we wrap names in quotes too, so multi-word lists would work if the user ever typed one."
- Mention that list/tag names are sent quoted so RTM doesn't greedily consume surrounding text.

Update `CLAUDE.md`'s "Smart Add uses a custom prefix syntax" bullet to reflect the quote-wrapping.

## Out of scope

- Changing the input grammar to support multi-word list/tag names in the rttui input itself (e.g. accepting `#"my list"` as user input). The smart input still tokenizes on whitespace. The quote wrapping here is purely an output-side escape for single-word names so they don't bleed into adjacent text on the RTM server.
- Switching to `list_id` for list assignment. Worth considering long-term (it's unambiguous and avoids Smart Add quirks entirely), but out of scope for this fix.

## Rollout / verification

1. Implement the transform change + tests; `go test ./...` green.
2. Manual test: with `add_preset = "#PrizePicks"`, add a task like `Review oddsfather 999 !1 ^today`. Expected in RTM web UI:
   - Title: `Review oddsfather 999`
   - List: `PrizePicks`
   - Priority: 1
   - Due: today
   - No stray auto-generated tags.
3. Clean up any junk tags RTM auto-created from previous broken adds (manual, via web UI).

## Risks

- **Quote syntax unsupported.** If RTM rejects or ignores `#"Name"`, every add will fail or misfile until reverted. Mitigation: verify on a throwaway task before merging.
- **Existing muscle memory.** Users who type `#foo bar` expecting RTM's greedy behavior to capture `foo bar` as a single tag will lose that. Not a concern here — rttui's tokenizer already splits on whitespace, so that case was never supported by rttui anyway.
