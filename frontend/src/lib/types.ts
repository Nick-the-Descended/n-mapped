// Mirrors the JSON shapes returned by the Go backend.
// Keep these in sync with internal/catalog and internal/nmap.

export type SkillLevel = 'beginner' | 'intermediate' | 'advanced';
export type ValueType = '' | 'boolean' | 'string' | 'int' | 'port-list' | 'ip';

export interface Example {
  command: string;
  explanation?: string;
}

export interface Warning {
  level: 'info' | 'warn' | 'danger';
  text: string;
}

export interface Reference {
  title: string;
  url: string;
}

export interface Flag {
  id: string;
  short?: string;
  long?: string;
  category: string;
  value_type?: ValueType;
  skill_level: SkillLevel;
  requires_root?: boolean;
  min_version?: string;
  max_version?: string;
  mutually_exclusive_with?: string[];
  implies?: string[];
  short_description: string;
  long_description?: string;
  examples?: Example[];
  warnings?: Warning[];
  references?: Reference[];
  tags?: string[];
}

export interface Category {
  id: string;
  name: string;
  description?: string;
  order: number;
}

export interface Catalog {
  schema_version: string;
  categories: Category[];
  flags: Flag[];
}

export type PrivilegeMode = 'user' | 'capability' | 'root';

export interface PrivilegeState {
  mode: PrivilegeMode;
  euid: number;
  os: string;
  capability?: string;
  reason?: string;
}

export interface NmapInfo {
  path: string;
  version: string;
  major: number;
  minor: number;
  raw: string;
  ok: boolean;
  error?: string;
}

export interface ScanRequest {
  targets: string[];
  flag_ids: string[];
  flag_values?: Record<string, string>;
  script_ids?: string[];
}

export interface ScanStartResponse {
  id: string;
  argv: string[];
  display: string;
  started: string;
}

export type ScanEventKind = 'start' | 'stdout' | 'stderr' | 'done' | 'error';

export interface ScanEvent {
  kind: ScanEventKind;
  line?: string;
  when: string;
  code?: number;
  err?: string;
}
