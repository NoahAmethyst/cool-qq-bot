# Log Level Color Design

## Goal

Make console logs visually distinguishable in Kubernetes Pod log output:

- Color only the `INFO` level label green.
- Color the complete `WARN` log line yellow.
- Color the complete `ERROR` log line red.

The change must preserve the existing log text layout and must not add ANSI
color sequences to file logs or to console logs when colorful output is
disabled.

## Current behavior

`global.LogFormat` formats both console and file output. When `EnableColor` is
true, it currently writes one level-specific ANSI sequence before the
timestamp and resets the color after the complete line. As a result, every
colored level affects the complete line. `INFO` currently uses white, while
`WARN` and `ERROR` already use yellow and red.

`cmd/gocq.Main` creates two formatter instances:

- Console formatter: color controlled by `output.log-colorful`.
- File formatter: color always disabled.

The Kubernetes manifest and runtime configuration do not need to change.

## Selected design

Keep the existing `global.LogFormat` and make its color placement
level-sensitive:

1. For `INFO`, write the timestamp and opening level bracket without a color
   sequence, write the green sequence immediately before `INFO`, and reset it
   immediately after `INFO`.
2. For all other levels, retain the existing whole-line placement: write the
   level color before the timestamp and reset it after the line.
3. Change the `INFO` color code from white to green.
4. When `EnableColor` is false, write no ANSI sequences.

This preserves the existing whole-line behavior for `DEBUG`, `TRACE`,
`WARN`, `ERROR`, `FATAL`, and `PANIC`. In particular, `WARN` remains yellow
and `ERROR` remains red across the complete line.

## Alternatives considered

### Color-scope policy table

Define explicit per-level policies such as `labelOnly` and `wholeLine`. This
would be convenient if more level-specific rules were expected, but adds
types and indirection for a single exception.

### Replace the formatter

Adopt or wrap Logrus's text formatter. This makes mixed label-only and
whole-line coloring harder to express and risks changing timestamps,
punctuation, and file output. It is not appropriate for this focused change.

## Testing

Add focused tests for `global.LogFormat`:

- Color-enabled `INFO` output has green ANSI codes directly around only the
  `INFO` label; the timestamp and message are outside the green scope.
- Color-enabled `WARN` output starts with yellow and resets only after the
  complete line.
- Color-enabled `ERROR` output starts with red and resets only after the
  complete line.
- Color-disabled output contains no ANSI escape sequences.

Tests will use a fixed timestamp so the exact formatted bytes are
deterministic. Implementation will follow red-green-refactor: add the failing
tests first, verify the expected failures, then make the smallest formatter
change that passes them.

## Scope and compatibility

Only the formatter and its focused tests are in scope. No configuration,
Kubernetes manifest, deployment workflow, logging level, file path, or log
message content changes are included.

Kubernetes log collectors will continue receiving ANSI bytes when
`output.log-colorful` is enabled, matching the project's current behavior.
Consumers that do not render ANSI colors may display or store those control
sequences exactly as they do today.
