// Builder state — what flags the user has selected and what target they typed.
// Kept in module scope as a plain reactive object that components import.

import type { Flag } from './types';

export interface BuilderState {
  selectedFlagIds: Set<string>;
  flagValues: Record<string, string>;
  targets: string;
}

// Plain singleton; components mutate fields directly. We deliberately avoid
// Svelte 5 runes in this module so the store can be shared across components
// without prop-drilling. Components should subscribe via $state mirrors when
// they need to react to changes.
export const builder: BuilderState = {
  selectedFlagIds: new Set<string>(),
  flagValues: {},
  targets: '',
};

// Build the canonical command preview from the current builder state.
export function previewCommand(allFlags: Flag[], targets: string, selected: Set<string>, values: Record<string, string>): string {
  const parts: string[] = ['nmap'];
  const sorted = [...selected].sort();
  for (const id of sorted) {
    const flag = allFlags.find((f) => f.id === id);
    if (!flag) continue;
    const tok = flag.short || flag.long;
    if (!tok) continue;
    parts.push(tok);
    if (flag.value_type && flag.value_type !== 'boolean') {
      const v = values[id];
      if (v) parts.push(v);
      else parts.push('<value>');
    }
  }
  parts.push('-oX', '-');
  parts.push('--stats-every', '2s');
  const tlist = targets
    .split(/[\s,]+/)
    .map((t) => t.trim())
    .filter(Boolean);
  if (tlist.length > 0) {
    parts.push('--', ...tlist);
  } else {
    parts.push('--', '<target>');
  }
  return parts.join(' ');
}

// Plain-English summary from selected flag short_descriptions.
export function previewSummary(allFlags: Flag[], selected: Set<string>): string {
  if (selected.size === 0) return 'Pick one or more options below to build a scan.';
  const lines: string[] = [];
  for (const id of selected) {
    const f = allFlags.find((x) => x.id === id);
    if (f) lines.push('• ' + f.short_description);
  }
  return lines.join('\n');
}

// ParseResult is what the reverse parser returns from a free-form nmap
// command: the matched flag IDs (with their values), plus the leftover
// tokens that look like targets, plus any tokens we couldn't match.
export interface ParseResult {
  flagIDs: string[];
  flagValues: Record<string, string>;
  targets: string[];
  unrecognized: string[];
}

const SKIP_FLAGS = new Set([
  '-oX', '-oG', '-oA', '-oN', '-oS', // output format flags we always inject
  '--stats-every',                    // injected automatically
  '--', '-',                          // separators / stdout marker
]);

// Tokenize a shell-style command line. Single and double quotes preserve
// whitespace; backslashes are passed through. Good enough for a paste-and-
// match flow — not a full POSIX shell parser.
function tokenize(input: string): string[] {
  const out: string[] = [];
  let buf = '';
  let quote: '' | "'" | '"' = '';
  for (let i = 0; i < input.length; i++) {
    const c = input[i];
    if (quote) {
      if (c === quote) { quote = ''; continue; }
      buf += c;
      continue;
    }
    if (c === '"' || c === "'") { quote = c; continue; }
    if (/\s/.test(c)) {
      if (buf) { out.push(buf); buf = ''; }
      continue;
    }
    buf += c;
  }
  if (buf) out.push(buf);
  return out;
}

// Match a token against the catalog. Tries exact short/long, then the
// "-T<n>" pattern where -T0..-T5 collapse to the timing-template flag, and
// finally the long=value style (--top-ports=20).
function matchFlag(token: string, allFlags: Flag[]): { flag: Flag; inlineValue?: string } | null {
  for (const f of allFlags) {
    if (f.short && token === f.short) return { flag: f };
    if (f.long && token === f.long) return { flag: f };
  }
  // Inline value: --foo=bar
  if (token.startsWith('--') && token.includes('=')) {
    const eq = token.indexOf('=');
    const head = token.slice(0, eq);
    const val = token.slice(eq + 1);
    for (const f of allFlags) {
      if (f.long === head) return { flag: f, inlineValue: val };
    }
  }
  // Glued value: -p22, -T4
  for (const f of allFlags) {
    if (f.short && f.value_type && f.value_type !== 'boolean' && token.startsWith(f.short) && token.length > f.short.length) {
      return { flag: f, inlineValue: token.slice(f.short.length) };
    }
  }
  return null;
}

export function parseCommand(input: string, allFlags: Flag[]): ParseResult {
  const tokens = tokenize(input.trim());
  const result: ParseResult = { flagIDs: [], flagValues: {}, targets: [], unrecognized: [] };
  const seen = new Set<string>();

  // Drop the leading "nmap" if present.
  let i = 0;
  if (tokens[0] && /^nmap(\.exe)?$/i.test(tokens[0].split('/').pop() ?? '')) i = 1;
  if (tokens[0] && tokens[0].endsWith('/nmap')) i = 1;

  let afterDoubleDash = false;

  for (; i < tokens.length; i++) {
    const tok = tokens[i];
    if (tok === '--') { afterDoubleDash = true; continue; }
    if (afterDoubleDash) {
      result.targets.push(tok);
      continue;
    }
    if (SKIP_FLAGS.has(tok)) {
      // Some skip flags consume the next token as a value (e.g. --stats-every 2s, -oX -).
      if (tok === '--stats-every' || tok === '-oX' || tok === '-oG' || tok === '-oA' || tok === '-oN' || tok === '-oS') {
        i++;
      }
      continue;
    }
    if (!tok.startsWith('-')) {
      // Bare token — almost always a target.
      result.targets.push(tok);
      continue;
    }
    const m = matchFlag(tok, allFlags);
    if (!m) { result.unrecognized.push(tok); continue; }
    const { flag, inlineValue } = m;
    if (seen.has(flag.id)) continue;
    seen.add(flag.id);
    result.flagIDs.push(flag.id);
    if (flag.value_type && flag.value_type !== 'boolean') {
      if (inlineValue !== undefined) {
        result.flagValues[flag.id] = inlineValue;
      } else if (i + 1 < tokens.length && !tokens[i + 1].startsWith('-')) {
        result.flagValues[flag.id] = tokens[++i];
      }
    }
  }
  return result;
}
