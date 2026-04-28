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

export interface ScriptArg {
  name: string;
  type: 'string' | 'int' | 'boolean';
  default?: string;
  description?: string;
}

export interface NSEScript {
  id: string;
  categories: string[];
  skill_level: SkillLevel;
  short_description: string;
  long_description?: string;
  args?: ScriptArg[];
  examples?: Example[];
  warnings?: Warning[];
  references?: Reference[];
  tags?: string[];
}

export interface Profile {
  id: string;
  name: string;
  description?: string;
  skill_level: SkillLevel;
  icon?: string;
  needs_root?: boolean;
  flag_ids?: string[];
  flag_values?: Record<string, string>;
  script_ids?: string[];
  script_args?: Record<string, string>;
}

export interface Catalog {
  schema_version: string;
  categories: Category[];
  flags: Flag[];
  scripts?: NSEScript[];
  profiles?: Profile[];
}

export interface Favorite {
  id: string;
  name: string;
  description?: string;
  targets?: string;
  flag_ids?: string[];
  flag_values?: Record<string, string>;
  script_ids?: string[];
  script_args?: Record<string, string>;
  tags?: string[];
  created_at: string;
  updated_at: string;
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
  script_args?: Record<string, string>;
}

export interface ScanStartResponse {
  id: string;
  argv: string[];
  display: string;
  started: string;
}

export type ScanEventKind =
  | 'start'
  | 'stderr'
  | 'scaninfo'
  | 'host'
  | 'taskprogress'
  | 'runstats'
  | 'done'
  | 'error';

export interface ScanInfo {
  type: string;
  protocol: string;
  numservices: number;
  services?: string;
}

export interface Status { state: string; reason?: string; reason_ttl?: number; }
export interface Address { addr: string; addrtype: string; vendor?: string; }
export interface Hostname { name: string; type: string; }
export interface Hostnames { hostname?: Hostname[]; }

export interface PortState { state: string; reason?: string; reason_ttl?: number; }
export interface Service {
  name: string;
  product?: string;
  version?: string;
  extrainfo?: string;
  method?: string;
  conf?: number;
  ostype?: string;
  cpe?: string[];
}
export interface ScriptOutput { id: string; output: string; }
export interface Port {
  protocol: string;
  portid: number;
  state: PortState;
  service?: Service;
  scripts?: ScriptOutput[];
}
export interface ExtraPorts { state: string; count: number; }
export interface Ports { extraports?: ExtraPorts[]; ports?: Port[]; }

export interface OSClass {
  type?: string;
  vendor?: string;
  osfamily?: string;
  osgen?: string;
  accuracy?: number;
}
export interface OSMatch { name: string; accuracy: number; classes?: OSClass[]; }
export interface OS { matches?: OSMatch[]; }

export interface HostTimes { srtt?: number; rttvar?: number; to?: number; }

export interface Host {
  starttime?: number;
  endtime?: number;
  status: Status;
  addresses: Address[];
  hostnames: Hostnames;
  ports?: Ports;
  os?: OS;
  times?: HostTimes;
}

export interface TaskProgress {
  task?: string;
  time?: number;
  percent: number;
  remaining?: number;
  etc?: number;
}

export interface RunStats {
  finished: { time: number; timestr?: string; elapsed: number; summary?: string; exit?: string };
  hosts: { up: number; down: number; total: number };
}

export interface Run {
  scanner: string;
  args: string;
  start: number;
  startstr?: string;
  version: string;
  xmloutputversion?: string;
  scaninfo?: ScanInfo[];
  hosts: Host[];
  runstats?: RunStats;
}

export interface ScanEvent {
  kind: ScanEventKind;
  when: string;
  line?: string;
  code?: number;
  err?: string;
  scaninfo?: ScanInfo;
  host?: Host;
  progress?: TaskProgress;
  runstats?: RunStats;
}

export interface HistorySummary {
  id: string;
  display: string;
  targets: string[];
  started: string;
  ended: string;
  exit_code: number;
  hosts_up: number;
  hosts_total: number;
  open_ports: number;
  error?: string;
}

export interface HistoryRecord {
  id: string;
  display: string;
  argv: string[];
  targets: string[];
  flag_ids?: string[];
  started: string;
  ended: string;
  exit_code: number;
  result?: {
    id: string;
    argv: string[];
    display: string;
    targets?: string[];
    flag_ids?: string[];
    started: string;
    ended: string;
    exit_code: number;
    run?: Run;
    raw_xml?: string;
    error?: string;
  };
  error?: string;
}
