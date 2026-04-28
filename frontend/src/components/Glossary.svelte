<script lang="ts">
  // A small in-app dictionary of nmap / networking terms beginners hit first.
  // Keep entries short — readers can click out to nmap.org for depth.
  const entries: Array<{ term: string; def: string; }> = [
    { term: 'TCP SYN scan (-sS)', def: 'Sends a SYN, never finishes the handshake. Fast and quieter than a full connect, but needs raw-socket privileges.' },
    { term: 'TCP Connect scan (-sT)', def: 'Asks the OS to fully open a TCP connection to each port. Works without root; slower and noisier than -sS.' },
    { term: 'UDP scan (-sU)', def: 'Probes UDP ports. Slow because closed UDP ports are silent; pair with -sV to identify services.' },
    { term: 'Ping scan (-sn)', def: 'Discover live hosts only — no port scan. The fastest way to map a subnet.' },
    { term: 'OS detection (-O)', def: 'Compares responses against nmap\'s OS database to guess the remote operating system. Needs at least one open and one closed port.' },
    { term: 'Service / version detection (-sV)', def: 'Probes open ports to identify what software (and version) is listening.' },
    { term: 'NSE script (-sC, --script)', def: 'Nmap Scripting Engine: small Lua programs that extend scans with banner grabbers, vuln checks, brute force, etc.' },
    { term: 'Filtered port', def: 'A firewall is dropping or rewriting packets — nmap can\'t tell if the port is open or closed.' },
    { term: 'Open|filtered port', def: 'Common for UDP / FIN / NULL scans where lack of response is ambiguous.' },
    { term: 'Top ports (--top-ports N)', def: 'Scan the N most commonly-open ports based on nmap\'s frequency data. Far faster than scanning all 65,535.' },
    { term: 'Timing template (-T0..-T5)', def: 'Speed presets. -T3 default, -T4 good for LANs, -T5 trades reliability for speed; -T0/-T1 are stealthy.' },
    { term: 'CIDR (192.168.1.0/24)', def: 'Address+prefix shorthand for a range of IPs. /24 = 256 addresses, /16 = 65,536, etc.' },
    { term: 'Raw sockets', def: 'Direct access to the network layer; required for SYN/UDP/OS scans. Available to root or with cap_net_raw on Linux.' },
  ];

  let query = $state('');
  let visible = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return entries;
    return entries.filter((e) => e.term.toLowerCase().includes(q) || e.def.toLowerCase().includes(q));
  });
</script>

<details class="glossary">
  <summary>Glossary &mdash; nmap terms in plain English</summary>
  <div class="body">
    <input type="text" placeholder="Filter terms…" bind:value={query} />
    <dl>
      {#each visible as e (e.term)}
        <dt>{e.term}</dt>
        <dd>{e.def}</dd>
      {/each}
    </dl>
    <p class="more">
      Need more depth? See the <a href="https://nmap.org/book/" target="_blank" rel="noopener noreferrer">Nmap Reference Guide</a>.
    </p>
  </div>
</details>

<style>
  .glossary {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
  }
  summary {
    list-style: none;
    cursor: pointer;
    padding: 0.85rem 1.1rem;
    font-weight: 600;
    color: var(--text-dim);
  }
  summary::-webkit-details-marker { display: none; }
  summary::before { content: '▸ '; color: var(--text-dim); }
  details[open] > summary::before { content: '▾ '; }
  .body { padding: 0 1.1rem 1rem; }
  input { margin-bottom: 0.85rem; }
  dl { margin: 0; }
  dt {
    font-weight: 600;
    margin-top: 0.6rem;
    color: var(--accent);
  }
  dd {
    margin: 0.2rem 0 0;
    color: var(--text-dim);
    font-size: 0.9rem;
  }
  .more { color: var(--text-dim); font-size: 0.85rem; margin-top: 1rem; }
</style>
