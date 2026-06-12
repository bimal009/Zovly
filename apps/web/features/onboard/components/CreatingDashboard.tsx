"use client";

import { useEffect, useState, type CSSProperties } from "react";

interface Props {
  phase: "creating" | "ready";
  onFinish: () => void;
}

const READY_MESSAGES = [
  "Building your workspace...",
  "Setting up your tools...",
  "Making your dashboard ready...",
];

const TOTAL_MS = 4500;

const DRAW: Record<string, [number, number]> = {
  frame:  [0.70, 0.00],
  nav:    [0.35, 0.65],
  side:   [0.30, 0.90],
  logo:   [0.25, 1.05],
  nd1:    [0.18, 1.20], nd2: [0.18, 1.32], nd3: [0.18, 1.44],
  si1:    [0.20, 1.50], si2: [0.20, 1.64], si3: [0.20, 1.78], si4: [0.20, 1.92],
  c1:     [0.38, 1.95], c2:  [0.38, 2.06], c3:  [0.38, 2.17],
  cv1:    [0.26, 2.28], cv2: [0.26, 2.33], cv3: [0.26, 2.38],
  cl1:    [0.20, 2.48], cl2: [0.20, 2.52], cl3: [0.20, 2.56],
  chart:  [0.44, 2.60],
  ctitle: [0.22, 2.90],
  cline:  [0.80, 3.00],
  cd1:    [0.18, 3.20], cd2: [0.18, 3.30], cd3: [0.18, 3.40], cd4: [0.18, 3.45],
  table:  [0.40, 3.48],
  th:     [0.26, 3.76],
  tr1:    [0.24, 3.90],
  tr2:    [0.24, 4.02],
};

// Returns just the animation — stroke/dash attrs go on the SVG element directly
function da(id: string): CSSProperties {
  const [dur, delay] = DRAW[id] ?? [0.4, 0];
  return { animation: `zovly-draw ${dur}s ease-out ${delay}s both` };
}

function rp(x: number, y: number, w: number, h: number, rx = 0): string {
  if (rx === 0) return `M${x},${y} h${w} v${h} h${-w}Z`;
  return (
    `M${x + rx},${y} h${w - 2 * rx}` +
    ` a${rx},${rx} 0 0 1 ${rx},${rx}` +
    ` v${h - 2 * rx}` +
    ` a${rx},${rx} 0 0 1 ${-rx},${rx}` +
    ` h${-(w - 2 * rx)}` +
    ` a${rx},${rx} 0 0 1 ${-rx},${-rx}` +
    ` v${-(h - 2 * rx)}` +
    ` a${rx},${rx} 0 0 1 ${rx},${-rx}Z`
  );
}

function cp(cx: number, cy: number, r: number): string {
  return `M${cx - r},${cy} a${r},${r} 0 1,0 ${2 * r},0 a${r},${r} 0 1,0 ${-2 * r},0`;
}

const CARDS: [number, number][] = [[68, 142], [222, 142], [376, 158]];
const CHART_DOTS: [number, number][] = [[84, 207], [234, 174], [384, 167], [528, 140]];

const PARTICLES = [
  { x: 12, y: 65, s: 3, d: 4.2, delay: 0.3 },
  { x: 80, y: 40, s: 2, d: 3.6, delay: 0.8 },
  { x: 35, y: 75, s: 4, d: 5.0, delay: 1.4 },
  { x: 88, y: 70, s: 2, d: 4.5, delay: 1.9 },
  { x: 22, y: 45, s: 3, d: 3.8, delay: 2.5 },
  { x: 65, y: 55, s: 2, d: 4.8, delay: 0.6 },
  { x: 50, y: 80, s: 3, d: 4.0, delay: 2.0 },
  { x: 92, y: 30, s: 2, d: 3.5, delay: 3.0 },
];

const CSS = `
  .z-orb {
    position: absolute;
    border-radius: 50%;
    pointer-events: none;
  }
  .z-orb-1 {
    top: -25%; left: -15%;
    width: 65vmin; height: 65vmin;
    background: radial-gradient(circle, color-mix(in srgb, var(--primary) 14%, transparent) 0%, transparent 65%);
    animation: z-orb-drift 9s ease-in-out infinite;
  }
  .z-orb-2 {
    bottom: -20%; right: -12%;
    width: 75vmin; height: 75vmin;
    background: radial-gradient(circle, color-mix(in srgb, var(--primary) 9%, transparent) 0%, transparent 65%);
    animation: z-orb-drift 12s ease-in-out infinite 2.5s;
  }
  @keyframes z-orb-drift {
    0%, 100% { transform: scale(1) translate(0, 0); opacity: 1; }
    50%       { transform: scale(1.18) translate(24px, -18px); opacity: 0.75; }
  }

  .z-grid {
    position: absolute;
    inset: 0;
    width: 100%; height: 100%;
    pointer-events: none;
    opacity: 0.025;
    color: var(--foreground);
  }

  .z-particle {
    position: absolute;
    border-radius: 50%;
    background: color-mix(in srgb, var(--primary) 55%, transparent);
    animation: z-particle-rise linear infinite;
    pointer-events: none;
  }
  @keyframes z-particle-rise {
    0%   { transform: translateY(0);      opacity: 0;   }
    15%  { opacity: 0.8; }
    80%  { opacity: 0.25; }
    100% { transform: translateY(-110px); opacity: 0;   }
  }

  .z-ring {
    position: absolute;
    border-radius: 50%;
    border: 1px solid color-mix(in srgb, var(--primary) 50%, transparent);
    animation: z-ring-expand 2.6s ease-out infinite;
  }
  @keyframes z-ring-expand {
    0%   { transform: scale(0.55); opacity: 0.8; }
    100% { transform: scale(1.55); opacity: 0;   }
  }

  @keyframes z-brand-pulse {
    0%, 100% { opacity: 0.6; }
    50%       { opacity: 1; }
  }

  /* ── Core SVG draw animation ──────────────────── */
  @keyframes zovly-draw {
    from { stroke-dashoffset: 100; }
    to   { stroke-dashoffset: 0;   }
  }

  .z-shimmer {
    background: linear-gradient(
      100deg,
      var(--foreground) 0%,
      var(--foreground) 30%,
      var(--primary)    50%,
      var(--foreground) 70%,
      var(--foreground) 100%
    );
    background-size: 220% auto;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    animation: z-text-shimmer 3.2s linear infinite;
  }
  @keyframes z-text-shimmer {
    from { background-position: 220% center; }
    to   { background-position: -220% center; }
  }
  @keyframes z-msg-up {
    from { opacity: 0; transform: translateY(10px); }
    to   { opacity: 1; transform: translateY(0);    }
  }

  .z-shine {
    position: absolute;
    inset-block: 0;
    left: 0;
    width: 45%;
    background: linear-gradient(90deg, transparent 0%, rgba(255,255,255,0.28) 50%, transparent 100%);
    animation: z-bar-sweep 1.9s ease-in-out infinite;
  }
  @keyframes z-bar-sweep {
    from { transform: translateX(-160%); }
    to   { transform: translateX(320%);  }
  }

  @media (prefers-reduced-motion: reduce) {
    .z-orb, .z-ring, .z-particle, .z-shine, .z-shimmer { animation: none !important; }
    [data-draw] { animation-duration: 0.001s !important; animation-delay: 0s !important; }
  }
`;

export function CreatingDashboard({ phase, onFinish }: Props) {
  const [msgIndex, setMsgIndex] = useState(0);
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    if (phase !== "ready") return;
    const finish = setTimeout(onFinish, TOTAL_MS);
    const msg = setInterval(
      () => setMsgIndex((i) => Math.min(i + 1, READY_MESSAGES.length - 1)),
      1500,
    );
    const prog = setInterval(
      () => setProgress((p) => Math.min(p + 100 / (TOTAL_MS / 100), 100)),
      100,
    );
    return () => {
      clearTimeout(finish);
      clearInterval(msg);
      clearInterval(prog);
    };
  }, [phase, onFinish]);

  return (
    <div className="fixed inset-0 z-50 bg-background overflow-hidden flex items-center justify-center">
      <style>{CSS}</style>

      <div className="z-orb z-orb-1" aria-hidden="true" />
      <div className="z-orb z-orb-2" aria-hidden="true" />

      <svg className="z-grid" aria-hidden="true">
        <defs>
          <pattern id="z-gp" x="0" y="0" width="40" height="40" patternUnits="userSpaceOnUse">
            <path d="M 40 0 L 0 0 0 40" fill="none" stroke="currentColor" strokeWidth="0.6" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#z-gp)" />
      </svg>

      {phase === "ready" &&
        PARTICLES.map((p, i) => (
          <div
            key={i}
            className="z-particle"
            style={{
              left: `${p.x}%`,
              top: `${p.y}%`,
              width: p.s,
              height: p.s,
              animationDuration: `${p.d}s`,
              animationDelay: `${p.delay}s`,
            }}
            aria-hidden="true"
          />
        ))}

      <div className="relative z-10 w-full flex flex-col items-center justify-center min-h-full px-6 py-12 gap-10">
        {phase === "creating" ? (
          <div className="flex flex-col items-center gap-10">
            {/* Rings */}
            <div className="relative flex items-center justify-center w-44 h-44">
              <div className="z-ring absolute w-44 h-44" />
              <div className="z-ring absolute w-32 h-32" style={{ animationDelay: "0.65s" }} />
              <div className="z-ring absolute w-20 h-20" style={{ animationDelay: "1.3s" }} />
              <div
                className="w-4 h-4 rounded-full bg-primary/60"
                style={{ animation: "z-brand-pulse 2.8s ease-in-out infinite" }}
              />
            </div>

            <div className="text-center space-y-2.5">
              <h2 className="text-2xl font-semibold tracking-tight text-foreground">
                Creating your business
                <span className="inline-flex gap-0.5 ml-1" aria-hidden="true">
                  {[0, 0.18, 0.36].map((d) => (
                    <span
                      key={d}
                      className="inline-block w-1 h-1 rounded-full bg-primary animate-bounce"
                      style={{ animationDelay: `${d}s` }}
                    />
                  ))}
                </span>
              </h2>
              <p className="text-sm text-muted-foreground">
                Hang tight, this&rsquo;ll be quick
              </p>
            </div>
          </div>
        ) : (
          <>
            {/* SVG drawing card */}
            <div className="w-full max-w-[580px] rounded-2xl border border-border/40 bg-card/35 backdrop-blur-md p-5 shadow-2xl shadow-black/10 ring-1 ring-white/5">
              <svg
                key="draw"
                viewBox="0 0 560 310"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                className="w-full"
                aria-hidden="true"
              >
                <defs />

                {/* ── Outer frame ── */}
                <path d={rp(2,2,556,306,10)} pathLength="100"
                  stroke="var(--border)" strokeWidth="1.5"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("frame")} />

                {/* ── Nav divider ── */}
                <path d="M2,46 L558,46" pathLength="100"
                  stroke="var(--border)" strokeWidth="1"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("nav")} />

                {/* ── Sidebar divider ── */}
                <path d="M54,46 L54,308" pathLength="100"
                  stroke="var(--border)" strokeWidth="1"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("side")} />

                {/* ── Logo mark ── */}
                <path d={rp(12,14,30,18,4)} pathLength="100"
                  stroke="var(--primary)" strokeWidth="1.5"
                  fill="color-mix(in srgb, var(--primary) 12%, transparent)"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("logo")} />

                {/* ── Nav dots ── */}
                {([490, 512, 534] as const).map((cx, i) => (
                  <path key={cx} d={cp(cx,23,5.5)} pathLength="100"
                    stroke="var(--border)" strokeWidth="1.5"
                    strokeDasharray="100" strokeDashoffset="100"
                    data-draw style={da(`nd${i + 1}`)} />
                ))}

                {/* ── Sidebar items ── */}
                {([60, 82, 104, 126] as const).map((y, i) => (
                  <path key={y} d={rp(15,y,24,7,2)} pathLength="100"
                    stroke="var(--border)" strokeWidth="1"
                    fill="color-mix(in srgb, var(--muted) 60%, transparent)"
                    strokeDasharray="100" strokeDashoffset="100"
                    data-draw style={da(`si${i + 1}`)} />
                ))}

                {/* ── Stat cards ── */}
                {CARDS.map(([x, w], i) => (
                  <g key={x}>
                    <path d={rp(x,58,w,56,6)} pathLength="100"
                      stroke="var(--border)" strokeWidth="1.5"
                      fill="color-mix(in srgb, var(--muted) 25%, transparent)"
                      strokeDasharray="100" strokeDashoffset="100"
                      data-draw style={da(`c${i + 1}`)} />
                    <path d={rp(x+10,71,68,10,2)} pathLength="100"
                      stroke="var(--border)" strokeWidth="1"
                      fill="color-mix(in srgb, var(--muted) 60%, transparent)"
                      strokeDasharray="100" strokeDashoffset="100"
                      data-draw style={da(`cv${i + 1}`)} />
                    <path d={rp(x+10,87,42,6,1)} pathLength="100"
                      stroke="var(--border)" strokeWidth="1"
                      fill="color-mix(in srgb, var(--muted) 40%, transparent)"
                      strokeDasharray="100" strokeDashoffset="100"
                      data-draw style={da(`cl${i + 1}`)} />
                  </g>
                ))}

                {/* ── Chart area ── */}
                <path d={rp(68,126,470,96,7)} pathLength="100"
                  stroke="var(--border)" strokeWidth="1.5"
                  fill="color-mix(in srgb, var(--muted) 12%, transparent)"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("chart")} />

                {/* ── Chart title ── */}
                <path d={rp(80,137,70,7,2)} pathLength="100"
                  stroke="var(--border)" strokeWidth="1"
                  fill="color-mix(in srgb, var(--muted) 50%, transparent)"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("ctitle")} />

                {/* ── Chart line (primary + glow) ── */}
                <path
                  d="M84,207 L134,189 L184,196 L234,174 L284,181 L334,161 L384,167 L434,148 L484,154 L528,140"
                  pathLength="100"
                  stroke="var(--primary)" strokeWidth="2.5"
                  strokeLinecap="round" strokeLinejoin="round"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("cline")} />

                {/* ── Chart dots ── */}
                {CHART_DOTS.map(([cx, cy], i) => (
                  <path key={cx} d={cp(cx,cy,4)} pathLength="100"
                    stroke="var(--primary)" strokeWidth="2"
                    fill="var(--background)"
                    strokeDasharray="100" strokeDashoffset="100"
                    data-draw style={da(`cd${i + 1}`)} />
                ))}

                {/* ── Table area ── */}
                <path d={rp(68,234,470,58,7)} pathLength="100"
                  stroke="var(--border)" strokeWidth="1.5"
                  fill="color-mix(in srgb, var(--muted) 12%, transparent)"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("table")} />

                {/* ── Table lines ── */}
                <path d="M80,250 L528,250" pathLength="100"
                  stroke="var(--border)" strokeWidth="1"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("th")} />
                <path d="M80,264 L528,264" pathLength="100"
                  stroke="var(--border)" strokeWidth="1"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("tr1")} />
                <path d="M80,278 L528,278" pathLength="100"
                  stroke="var(--border)" strokeWidth="1"
                  strokeDasharray="100" strokeDashoffset="100"
                  data-draw style={da("tr2")} />
              </svg>
            </div>

            {/* Status + progress */}
            <div className="w-full max-w-xs text-center space-y-4">
              <div
                key={msgIndex}
                style={{ animation: "z-msg-up 0.4s ease-out both" }}
              >
                <p className="z-shimmer text-xl font-bold tracking-tight leading-snug">
                  {READY_MESSAGES[msgIndex]}
                </p>
              </div>

              <div className="space-y-2">
                <div className="relative h-1.5 w-full bg-muted rounded-full overflow-hidden">
                  <div
                    className="absolute inset-y-0 left-0 rounded-full bg-primary transition-all duration-100 ease-linear overflow-hidden"
                    style={{ width: `${progress}%` }}
                  >
                    <div className="z-shine" />
                  </div>
                </div>
                <p className="text-xs text-muted-foreground tabular-nums text-right">
                  {Math.round(progress)}%
                </p>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
