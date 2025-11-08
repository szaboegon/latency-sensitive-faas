import pandas as pd
import matplotlib.pyplot as plt
import sys

# ---- Config ----
if len(sys.argv) < 2:
    print(
        "Usage: python plot_jtl_metrics.py <path/to/jtl_results.csv> [path/to/reconfig_events.csv]"
    )
    sys.exit(1)

file_path_csv = sys.argv[1]
file_path_events = sys.argv[2] if len(sys.argv) > 2 else None
granularity = 30  # seconds per bin

# ---- Load Data ----
print(f"Loading {file_path_csv} ...")
df = pd.read_csv(file_path_csv)

# Ensure required columns exist
required_cols = {"timeStamp", "elapsed"}
if not required_cols.issubset(df.columns):
    sys.exit(f"Error: JTL file missing required columns: {required_cols}")

# ---- Convert timestamps ----
# Convert to datetime, then relative seconds from start
df["time"] = pd.to_datetime(df["timeStamp"], unit="ms")
start_time = df["time"].min()
df["rel_time"] = (df["time"] - start_time).dt.total_seconds()

# Calculate max test time for filtering events and setting x-axis limit
max_test_time = df["rel_time"].max()

# ---- Compute average RPS per 30s ----
df["time_bin"] = (df["rel_time"] // granularity) * granularity
rps = df.groupby("time_bin").size().reset_index(name="requests")
rps["rps"] = rps["requests"] / granularity  # convert to RPS

# ---- Load Reconfiguration Events Data (Optional) ----
reconfig_events = None
if file_path_events:
    print(f"Loading reconfiguration events from {file_path_events} ...")
    try:
        events_df = pd.read_csv(file_path_events)

        # Calculate relative start time (seconds) and end time (seconds)
        events_df["rel_start_time"] = (
            pd.to_datetime(events_df["event_time"], unit="ms") - start_time
        ).dt.total_seconds()
        events_df["rel_end_time"] = events_df["rel_start_time"] + (
            events_df["duration_ms"] / 1000
        )

        # Filter events to those overlapping with the JTL test window.
        reconfig_events = events_df[
            (events_df["rel_start_time"] <= max_test_time)
            & (events_df["rel_end_time"] >= 0)
        ].copy()

        if reconfig_events.empty:
            print("Note: All extracted events occurred outside the test time window.")
    except Exception as e:
        print(
            f"Warning: Error reading or processing events file: {e}. Skipping plotting events."
        )

# ---- Plot ----
fig, ax1 = plt.subplots(figsize=(12, 6))
ax2 = ax1.twinx()

# Plot all latencies as scatter points
ax1.scatter(
    df["rel_time"], df["elapsed"], s=5, color="tab:red", alpha=0.4, label="Latency (ms)"
)

# Plot average RPS as a line
ax2.plot(
    rps["time_bin"], rps["rps"], color="tab:blue", linewidth=2, label="RPS (30s avg)"
)

# ---- Plot Reconfiguration Events (Conditional) ----
if reconfig_events is not None and not reconfig_events.empty:
    app_ids = reconfig_events["app_id"].unique()
    colors = plt.cm.get_cmap("Set1", len(app_ids))
    app_labels_seen = set()

    for i, app_id in enumerate(app_ids):
        app_events = reconfig_events[reconfig_events["app_id"] == app_id]
        event_color = colors(i)

        for index, row in app_events.iterrows():
            # Add label only once per app_id for the legend
            label = f"Reconfig: {app_id}" if app_id not in app_labels_seen else ""

            # Plot shaded region
            ax1.axvspan(
                row["rel_start_time"],
                row["rel_end_time"],
                alpha=0.15,
                color=event_color,
                label=label,
            )

            # Mark start time
            ax1.axvline(
                row["rel_start_time"],
                color=event_color,
                linestyle="--",
                linewidth=1,
                alpha=0.7,
            )

            app_labels_seen.add(app_id)

# ---- Labeling ----
ax1.set_xlabel("Time (seconds from start)")
ax1.set_ylabel("Latency (ms)", color="tab:red")
ax2.set_ylabel("RPS (avg per 30s)", color="tab:blue")
plt.title("Latency vs Average RPS (30s granularity)")

# ---- X-axis range ----
# Use dynamic x_limit based on max_test_time
x_limit = ((int(max_test_time) // 60) + 1) * 60
ax1.set_xlim(0, x_limit)

# ---- Grid & legend ----
ax1.grid(True, linestyle="--", alpha=0.5)

# Complex legend handling to de-duplicate axvspan labels
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
plt.show()
