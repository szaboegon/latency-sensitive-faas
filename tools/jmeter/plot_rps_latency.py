# type: ignore
import pandas as pd
import matplotlib.pyplot as plt
import sys
import os
from typing import Optional

# ---- Config ----
if len(sys.argv) < 2:
    print(
        "Usage: python plot_jtl_metrics.py <path/to/jtl_results.csv> [path/to/reconfig_events.csv] [target_latency_ms]"
    )
    sys.exit(1)

file_path_csv = sys.argv[1]
file_path_events = sys.argv[2] if len(sys.argv) > 2 else None
target_latency_str = sys.argv[3] if len(sys.argv) > 3 else None

granularity = 30  # seconds per bin

# ---- Load Data ----
print(f"Loading {file_path_csv} ...")
df = pd.read_csv(file_path_csv)

# Ensure required columns exist
required_cols = {"timeStamp", "elapsed"}
if not required_cols.issubset(df.columns):
    sys.exit(f"Error: JTL file missing required columns: {required_cols}")

# ---- Convert timestamps ----
df["time"] = pd.to_datetime(df["timeStamp"], unit="ms")
start_time = df["time"].min()
df["rel_time"] = (df["time"] - start_time).dt.total_seconds()

max_test_time = df["rel_time"].max()

# ---- Compute average RPS per 30s ----
df["time_bin"] = (df["rel_time"] // granularity) * granularity
rps = df.groupby("time_bin").size().reset_index(name="requests")
rps["rps"] = rps["requests"] / granularity

# ---- Load Reconfiguration Events Data (Optional) ----
reconfig_events = None
if file_path_events:
    print(f"Loading reconfiguration events from {file_path_events} ...")
    try:
        events_df = pd.read_csv(file_path_events)

        events_df["rel_start_time"] = (
            pd.to_datetime(events_df["event_time"], unit="ms") - start_time
        ).dt.total_seconds()

        events_df["duration_s"] = events_df["duration_ms"] / 1000
        events_df["rel_end_time"] = (
            events_df["rel_start_time"] + events_df["duration_s"]
        )

        reconfig_events = events_df[
            (events_df["rel_start_time"] <= max_test_time)
            & (events_df["rel_end_time"] >= 0)
        ].copy()

        if reconfig_events.empty:
            print("Note: All extracted events occurred outside the test time window.")

        if not reconfig_events.empty:
            print("\n--- Reconfiguration Event Summary ---")
            for index, row in reconfig_events.iterrows():
                print(
                    f"App ID: {row['app_id']} | "
                    f"Start Time (s): {row['rel_start_time']:.3f} | "
                    f"Duration (s): {row['duration_s']:.3f}"
                )
            print("-----------------------------------\n")

    except Exception as e:
        print(
            f"Warning: Error reading or processing events file: {e}. Skipping plotting events."
        )

# ---- Convert Target Latency (Optional) ----
target_latency: Optional[float] = None
if target_latency_str:
    try:
        target_latency = float(target_latency_str)
        print(f"Target latency set to {target_latency} ms.")
    except ValueError:
        print(
            f"Warning: Could not parse target latency '{target_latency_str}'. Skipping plotting target line."
        )

# ---- Plot ----
fig, ax1 = plt.subplots(figsize=(12, 6))
ax2 = ax1.twinx()

ax1.scatter(
    df["rel_time"], df["elapsed"], s=5, color="tab:red", alpha=0.4, label="Latency (ms)"
)

ax2.plot(
    rps["time_bin"], rps["rps"], color="tab:blue", linewidth=2, label="RPS (30s avg)"
)

if target_latency is not None:
    ax1.axhline(
        target_latency,
        color="black",
        linestyle="--",
        linewidth=2,
        alpha=0.8,
        label=f"Target Latency ({target_latency}ms)",
    )

if reconfig_events is not None and not reconfig_events.empty:
    app_ids = reconfig_events["app_id"].unique()
    colors = plt.colormaps["Set1"].resampled(len(app_ids))
    app_labels_seen = set()

    # Determine a suitable y-position for event duration labels
    # This places labels near the top of the latency axis, adjusted dynamically
    y_min, y_max = ax1.get_ylim()
    y_range = y_max - y_min
    label_y_pos = y_min + (y_range * 0.05)  # Start 95% up the y-axis

    for i, app_id in enumerate(app_ids):
        app_events = reconfig_events[reconfig_events["app_id"] == app_id]
        event_color = colors(i)

        for index, row in app_events.iterrows():
            label = "Reconfiguration event" if app_id not in app_labels_seen else ""

            ax1.axvspan(
                row["rel_start_time"],
                row["rel_end_time"],
                alpha=0.15,
                color=event_color,
                label=label,
            )

            ax1.axvline(
                row["rel_start_time"],
                color=event_color,
                linestyle="--",
                linewidth=1,
                alpha=0.7,
            )

            # --- NEW: Add duration label directly on the plot ---
            event_mid_x = (row["rel_start_time"] + row["rel_end_time"]) / 2
            duration_text = f"{row['duration_s']:.1f}s"
            ax1.text(
                event_mid_x,
                label_y_pos,  # Use the dynamically calculated Y position
                duration_text,
                color=event_color,
                ha="center",  # Horizontal alignment: center
                va="bottom",  # Vertical alignment: align bottom of text with y-pos
                fontsize=8,
                weight="bold",
                bbox=dict(
                    facecolor="white",
                    alpha=0.7,
                    edgecolor="none",
                    boxstyle="round,pad=0.2",
                ),  # Background box for readability
            )
            # You might want to adjust label_y_pos slightly downwards for subsequent events if they are close
            # Or implement a more sophisticated collision detection for labels if many events overlap
            # For now, if events are close on X, their labels might overlap.
            # ---------------------------------------------------

            app_labels_seen.add(app_id)

    print("Reconfiguration events visualized as shaded regions with duration labels.")


# ---- Labeling ----
ax1.set_xlabel("Time (seconds from start)")
ax1.set_ylabel("Latency (ms)", color="tab:red")
ax2.set_ylabel("RPS (avg per 30s)", color="tab:blue")
plt.title("Latency vs Average RPS (30s granularity)")

# ---- X-axis range ----
x_limit = ((int(max_test_time) // 60) + 1) * 60
ax1.set_xlim(0, x_limit)

# ---- Grid & legend ----
ax1.grid(True, linestyle="--", alpha=0.5)

lines, labels = ax1.get_legend_handles_labels()
lines2, labels2 = ax2.get_legend_handles_labels()
unique_lines = []
unique_labels = []
seen_labels_full = set()

for line, label in zip(lines, labels):
    if label != "" and label not in seen_labels_full:
        unique_lines.append(line)
        unique_labels.append(label)
        seen_labels_full.add(label)

unique_lines.extend(lines2)
unique_labels.extend(labels2)
ax1.legend(unique_lines, unique_labels, loc="upper right")

fig.tight_layout()


def save_plot_to_file(fig, base_filename="result.png"):
    """Saves the figure to a file, handling name collisions."""
    name, ext = os.path.splitext(base_filename)
    counter = 0
    filename = base_filename

    while os.path.exists(filename):
        counter += 1
        filename = f"{name}-{counter}{ext}"

    try:
        fig.savefig(filename)
        print(f"\n✅ Plot successfully saved to: {filename}")
    except Exception as e:
        print(f"\n❌ Error saving figure to {filename}: {e}")


save_plot_to_file(fig)
plt.show()
