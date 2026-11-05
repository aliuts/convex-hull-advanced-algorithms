# /// script
# dependencies = ["matplotlib"]
# ///
"""Draws the report's figures from a CSV saved by the experiments command.

    go run ./cmd/experiments -csv report/results.csv
    uv run report/plot.py [report/results.csv]

Figures go to report/figures/ (and are copied to ~/Downloads/convex-hull-figures).
"""
import csv
import os
import shutil
import statistics
import sys
from collections import defaultdict

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
from matplotlib.ticker import FuncFormatter

HERE = os.path.dirname(os.path.abspath(__file__))
CSV_PATH = sys.argv[1] if len(sys.argv) > 1 else os.path.join(HERE, "results.csv")
OUT = os.path.join(HERE, "figures")
os.makedirs(OUT, exist_ok=True)

INK, MUTED, GRID = "#1f1f1e", "#6b6a64", "#e4e3de"
ALGS = {  # fixed colour + marker per algorithm, used in every figure
    "Graham Scan": ("#2a78d6", "o"),
    "Jarvis March": ("#eb6834", "s"),
    "Chan's": ("#1baf7a", "^"),
}
plt.rcParams.update({
    "font.family": "DejaVu Sans", "font.size": 10, "text.color": INK,
    "axes.edgecolor": MUTED, "axes.labelcolor": INK, "axes.linewidth": 0.8,
    "xtick.color": MUTED, "ytick.color": MUTED, "axes.grid": True,
    "grid.color": GRID, "grid.linewidth": 0.8, "axes.axisbelow": True,
    "axes.spines.top": False, "axes.spines.right": False,
    "legend.frameon": False, "savefig.dpi": 300, "savefig.bbox": "tight",
})
thousands = FuncFormatter(lambda v, _: f"{v:,.0f}" if v >= 1 else f"{v:g}")

ALG_KEYS = {"graham_scan": "Graham Scan", "jarvis_march": "Jarvis March", "chan": "Chan's"}
KIND_TITLES = {"uniform": "Uniform random", "circular": "Circular ring",
               "clustered": "Clustered", "collinear": "Collinear"}
# Exact hull sizes of the collinear dataset (seed 42), from TestCollinearDatasetAgainstExactHull.
EXACT_COLLINEAR = {100: 8, 500: 10, 1000: 12, 5000: 15, 10000: 16}


def load(path):
    """Median time (ns) and hull size per (experiment, kind, n, target_h, algorithm)."""
    trials = defaultdict(list)
    hull = {}
    with open(path, newline="") as f:
        for r in csv.DictReader(f):
            # Older CSVs have no target_h column; a controlled polygon's hull is its target.
            target = r.get("target_h") or (r["hull_vertices"] if r["experiment"] == "controlled" else 0)
            key = (r["experiment"], r["kind"], int(r["n"]), int(target), ALG_KEYS[r["algorithm"]])
            trials[key].append(int(r["duration_ns"]))
            hull[key] = int(r["hull_vertices"])
    return {k: statistics.median(v) for k, v in trials.items()}, hull


TIME, HULL = load(CSV_PATH)
N = sorted({k[2] for k in TIME if k[0] == "distribution"})


def dist(kind, alg, unit=1e3):
    return [TIME[("distribution", kind, n, 0, alg)] / unit for n in N]


def dist_h(kind, alg):
    return [HULL[("distribution", kind, n, 0, alg)] for n in N]


CN = next(k[2] for k in TIME if k[0] == "controlled")
H = sorted({k[3] for k in TIME if k[0] == "controlled"})


def ctrl(alg):
    hs = [h for h in H if ("controlled", "polygon", CN, h, alg) in TIME]
    return hs, [TIME[("controlled", "polygon", CN, h, alg)] / 1e6 for h in hs]


def line(ax, x, y, name, **kw):
    c, m = ALGS[name]
    ax.plot(x, y, color=c, marker=m, markersize=6, linewidth=2,
            markeredgecolor="white", markeredgewidth=1, label=name, **kw)


# Figure 1: runtime vs n, one panel per distribution
fig, axes = plt.subplots(2, 2, figsize=(8, 6.2), sharex=True, sharey=True)
for ax, (kind, title) in zip(axes.flat, KIND_TITLES.items()):
    for name in ALGS:
        line(ax, N, dist(kind, name), name)
    ax.set_xscale("log"); ax.set_yscale("log")
    ax.set_title(title, loc="left", fontsize=10.5, fontweight="bold")
    ax.xaxis.set_major_formatter(thousands); ax.yaxis.set_major_formatter(thousands)
    ax.set_xticks(N); ax.minorticks_off()
for ax in axes[1]: ax.set_xlabel("Input size n (points)")
for ax in axes[:, 0]: ax.set_ylabel("Median time (µs, log scale)")
for ax in axes[1]: ax.tick_params(axis="x", labelrotation=30)
fig.legend(*axes[0, 0].get_legend_handles_labels(), loc="upper center",
           ncol=3, bbox_to_anchor=(0.5, 1.03))
fig.tight_layout(rect=(0, 0, 1, 0.97))
fig.savefig(f"{OUT}/fig1_runtime_by_distribution.png"); plt.close(fig)

# Figure 2: controlled hull size, n = 100,000
fig, ax = plt.subplots(figsize=(7, 4.3))
for name in ALGS:
    line(ax, *ctrl(name), name)
jh, jt = ctrl("Jarvis March")
if jh[-1] < H[-1]:  # Jarvis skipped at the largest h: extrapolate its O(nh) line
    est = jt[-1] / jh[-1] * H[-1]
    ax.plot([jh[-1], H[-1]], [jt[-1], est], color=ALGS["Jarvis March"][0], linewidth=1.5,
            linestyle=(0, (3, 3)))
    ax.annotate(f"Jarvis at h = n skipped\n(≈ {est / 1000:.0f} s estimated)", (H[-1] / 5, est * 0.9),
                ha="right", va="center", color=MUTED, fontsize=9)
ax.axvline(64, color=MUTED, linewidth=1, linestyle=":")
ax.annotate("Jarvis ≈ Chan's\nat h = 64", (64, 300), xytext=(6, 0),
            textcoords="offset points", color=MUTED, fontsize=9)
_, gt = ctrl("Graham Scan")
ax.annotate("Graham Scan fastest from h = 512", (900, gt[H.index(512)]), xytext=(0, -16),
            textcoords="offset points", ha="center", va="top", color=MUTED, fontsize=9)
ax.set_xscale("log"); ax.set_yscale("log")
ax.set_xticks(H); ax.minorticks_off()
ax.xaxis.set_major_formatter(thousands); ax.yaxis.set_major_formatter(thousands)
ax.set_xlabel(f"Hull size h (vertices), n = {CN:,}")
ax.set_ylabel("Time (ms, log scale)")
ax.legend(loc="lower right")
fig.tight_layout()
fig.savefig(f"{OUT}/fig2_controlled_hull_size.png"); plt.close(fig)

# Figure 3: Jarvis time vs n·h (one point per configuration with a single h)
fig, ax = plt.subplots(figsize=(6, 4.3))
shapes = {"Uniform random": "o", "Circular ring": "s", "Clustered": "D", "Collinear": "v"}
xs, ys = [], []
for kind, m in zip(KIND_TITLES, shapes.values()):
    title = KIND_TITLES[kind]
    j = dist(kind, "Jarvis March")
    nh = [n * hh for n, hh in zip(N, dist_h(kind, "Jarvis March"))]  # Jarvis's own h
    xs += nh; ys += j
    ax.scatter(nh, j, marker=m, s=42, color=ALGS["Jarvis March"][0],
               edgecolor="white", linewidth=1, label=title, zorder=3)
k = sum(x * y for x, y in zip(xs, ys)) / sum(x * x for x in xs)  # fit y = k·x
ax.plot([min(xs), max(xs)], [k * min(xs), k * max(xs)], color=MUTED, linewidth=1.2,
        linestyle="--", label=f"time ∝ n·h  ({k * 1000:.2f} ns per n·h)")
ax.set_xscale("log"); ax.set_yscale("log"); ax.minorticks_off()
ax.xaxis.set_major_formatter(thousands); ax.yaxis.set_major_formatter(thousands)
ax.set_xlabel("n · h (log scale)")
ax.set_ylabel("Jarvis March median time (µs, log scale)")
ax.legend(loc="upper left")
fig.tight_layout()
fig.savefig(f"{OUT}/fig3_jarvis_vs_nh.png"); plt.close(fig)

# Figure 4: collinear hull sizes vs the exact hull
exact = [EXACT_COLLINEAR[n] for n in N]
got = {"Exact hull": exact} | {name: dist_h("collinear", name) for name in ALGS}
colours = {"Exact hull": "#9a9890"} | {name: c for name, (c, _) in ALGS.items()}
fig, ax = plt.subplots(figsize=(7, 4))
w = 0.2
for i, (name, v) in enumerate(got.items()):
    xs = [p + (i - 1.5) * w for p in range(len(N))]
    bars = ax.bar(xs, v, width=w - 0.03, color=colours[name], label=name, zorder=3)
    ax.bar_label(bars, padding=2, fontsize=8, color=MUTED)
ax.set_xticks(range(len(N)), [f"{n:,}" for n in N])
ax.set_xlabel("Input size n (collinear dataset)")
ax.set_ylabel("Hull vertices returned")
ax.grid(axis="x", visible=False)
ax.legend(loc="upper left", ncol=2)
ax.set_ylim(0, max(max(v) for v in got.values()) * 1.12)
fig.tight_layout()
fig.savefig(f"{OUT}/fig4_collinear_hull_sizes.png"); plt.close(fig)
DOWNLOADS = os.path.expanduser("~/Downloads/convex-hull-figures")
shutil.copytree(OUT, DOWNLOADS, dirs_exist_ok=True)
print(f"{OUT}\n{DOWNLOADS}")
