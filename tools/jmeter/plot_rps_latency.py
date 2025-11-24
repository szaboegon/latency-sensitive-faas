# type: ignore
import pandas as pd
import matplotlib.pyplot as plt
import sys
import os
import redis
import json
import re
import shutil
from typing import Optional, List, Dict, Any
from datetime import datetime

# ---- Redis Config ----
REDIS_HOST = "localhost"
REDIS_PORT = 6379
# No fallback APP_NAME defined. The script will exit if detection fails.
# ----------------------


# ---- Function to Determine APP_NAME ----
def determine_app_name(file_path_csv: str) -> str:
    """
    Reads the JTL file to find and extract the application ID from the URL format.
    Exits the script if the APP_NAME cannot be determined.
    """
    # Pattern to extract app name from the expected URL structure
    pattern = re.compile(r"http://app-(.*?)\.application\.127\.0\.0\.1\.sslip\.io")

    try:
        # Read only the header and first 10 rows to quickly identify the URL
        df_head = pd.read_csv(file_path_csv, nrows=10)

        # Check for a 'URL' column specifically, then look in all string columns as a fallback
        search_cols = []
        if "URL" in df_head.columns:
            search_cols.append("URL")

        # Add other potential string columns
        search_cols.extend(
            col
            for col in df_head.select_dtypes(include=["object"]).columns
            if col not in search_cols
        )

        for col in search_cols:
            for value in df_head[col].dropna().astype(str):
                match = pattern.search(value)
                if match:
                    app_name = match.group(1)
                    print(f"✅ Auto-detected APP_NAME: {app_name} from column '{col}'.")
                    return app_name

    except FileNotFoundError:
        sys.stderr.write(f"Error: JTL file not found at {file_path_csv}\n")
        sys.exit(1)
    except Exception as e:
        sys.stderr.write(
            f"Error during APP_NAME auto-detection: {e}. Cannot proceed.\n"
        )
        sys.exit(1)

    # If the loop finishes without finding a match, exit
    sys.stderr.write(
        "Error: Failed to auto-detect APP_NAME from the JTL file. "
        "Could not find the pattern 'http://app-{app_name}.application.127.0.0.1.sslip.io' in the inspected rows. Exiting.\n"
    )
    sys.exit(1)


# ----------------------------------------


# ---- Command Line Setup ----
if len(sys.argv) < 2:
    print(
        "Usage: python plot_jtl_metrics.py <path/to/jtl_results.csv> [path/to/reconfig_events.csv] [target_latency_ms]"
    )
    sys.exit(1)

file_path_csv = sys.argv[1]
file_path_events = sys.argv[2] if len(sys.argv) > 2 else None
target_latency_str = sys.argv[3] if len(sys.argv) > 3 else None

# ---- Auto-Detect APP_NAME and Set Redis Key ----
APP_NAME = determine_app_name(file_path_csv)
PERF_KEY = f"perf:{APP_NAME}"
print(f"Using Redis Key: {PERF_KEY}")

# ---- Create Results Subfolder with Timestamp ----
RESULTS_BASE_FOLDER = "results"
if not os.path.exists(RESULTS_BASE_FOLDER):
    os.makedirs(RESULTS_BASE_FOLDER)

# Create timestamped subfolder for this run
timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
RUN_FOLDER = os.path.join(RESULTS_BASE_FOLDER, timestamp)
os.makedirs(RUN_FOLDER)
print(f"Created run folder: {RUN_FOLDER}")
# ------------------------------------------------

granularity = 30  # seconds per bin


# ---- Redis Helper Function ----
def fetch_redis_perf_data() -> pd.DataFrame:
    """Connects to Redis, retrieves all performance metrics, and returns a DataFrame."""
    print(f"Connecting to Redis at {REDIS_HOST}:{REDIS_PORT}...")
    try:
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)
        r.ping()
        print("Redis connection successful.")

        # Get all elements from the performance list
        raw_data: List[str] = r.lrange(PERF_KEY, 0, -1)

        if not raw_data:
            print(f"Warning: No data found in Redis list '{PERF_KEY}'.")
            return pd.DataFrame()

        # Parse JSON data
        parsed_data: List[Dict[str, Any]] = [json.loads(item) for item in raw_data]
        perf_df = pd.DataFrame(parsed_data)

        # --- Keep ALL correlation_id records as requested by the user ---

        required_perf_cols = {"correlation_id", "timestamp"}
        if not required_perf_cols.issubset(perf_df.columns):
            sys.exit(
                f"Error: Redis data missing required fields ({required_perf_cols}). Ensure your FaaS code uses 'timestamp'."
            )

        # NO de-duplication: Keep all events for a correlation ID to see all timespans
        print(f"Processing all {len(perf_df)} recorded end-time events from Redis.")

        # 1. Convert ISO string timestamp to milliseconds Unix epoch time (for latency calculation)
        # pandas.to_datetime returns nanoseconds, so we divide by 10^6 to get milliseconds
        perf_df["end_time_ms"] = (
            pd.to_datetime(perf_df["timestamp"], utc=True).astype("int64") // 10**6
        )

        # 2. Keep only the columns needed for merging: correlation_id and the newly created end_time_ms
        perf_df = perf_df[["correlation_id", "end_time_ms"]]
        perf_df["correlation_id"] = perf_df["correlation_id"].astype(str)

        return perf_df

    except redis.exceptions.ConnectionError as e:
        sys.exit(
            f"Error: Could not connect to Redis at {REDIS_HOST}:{REDIS_PORT}. Details: {e}"
        )
    except Exception as e:
        sys.exit(f"An unexpected error occurred during Redis operation: {e}")


# ---- Load Data ----
print(f"Loading {file_path_csv} ...")
df_jtl = pd.read_csv(file_path_csv)
df_jtl["correlation_id"] = df_jtl["correlation_id"].astype(str)

# Ensure required columns exist
required_cols = {"timeStamp", "elapsed", "correlation_id"}
if not required_cols.issubset(df_jtl.columns):
    sys.exit(
        f"Error: JTL file missing required columns: {required_cols}. Did you configure sample_variables=correlation_id?"
    )

# ---- Fetch and Merge Data ----
df_perf = fetch_redis_perf_data()

if df_perf.empty:
    sys.exit("Cannot proceed without performance data from Redis.")

# Merge JTL start times with Redis end times on the unique correlation ID
# Using an 'inner' merge will match every JTL start time against every corresponding Redis end time.
df_merged = pd.merge(df_jtl, df_perf, on="correlation_id", how="inner")

if df_merged.empty:
    sys.exit(
        "Error: Merged DataFrame is empty. Check correlation_id values in JTL and Redis."
    )

print(
    f"Successfully merged {len(df_merged)} records (out of {len(df_jtl)} JTL records)."
)
if len(df_merged) > len(df_jtl):
    print(
        f"Note: The merged DataFrame has more rows ({len(df_merged)}) than the JTL file ({len(df_jtl)}), as requested, because multiple end times were recorded per request."
    )

# ---- Recalculate End-to-End Latency ----

# 1. Start Time (from JTL, already in milliseconds)
df_merged["start_time_ms"] = df_merged["timeStamp"]

# 2. End Time (from Redis, already merged as 'end_time_ms')

# 3. Calculate True End-to-End Latency (in milliseconds)
df_merged["true_latency_ms"] = df_merged["end_time_ms"] - df_merged["start_time_ms"]

# Set the primary DataFrame for plotting to the merged, calculated data
df = df_merged.copy()

# ---- Convert timestamps ----
# Time is now based on the JTL Start Time (timeStamp)
df["time"] = pd.to_datetime(df["timeStamp"], unit="ms")
start_time = df["time"].min()
df["rel_time"] = (df["time"] - start_time).dt.total_seconds()

max_test_time = df["rel_time"].max()

# ---- Compute average RPS per 30s ----
# RPS is based on the start time (timeStamp) as this is when the request originated.
df["time_bin"] = (df["rel_time"] // granularity) * granularity
rps = df.groupby("time_bin").size().reset_index(name="requests")
rps["rps"] = rps["requests"] / granularity

# ---- Load Reconfiguration Events Data (Optional) ----
reconfig_events = None
if file_path_events:
    print(f"Loading reconfiguration events from {file_path_events} ...")
    try:
        events_df = pd.read_csv(file_path_events)

        # Assuming event_time in events_df is also in milliseconds since epoch
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

# Plot the new, calculated true end-to-end latency
# This scatter will now include multiple points per initial request if multiple end events were logged.
ax1.scatter(
    df["rel_time"],
    df["true_latency_ms"],
    s=5,
    color="tab:red",
    alpha=0.4,
    label="Latency (ms)",
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
    y_min, y_max = ax1.get_ylim()
    y_range = y_max - y_min
    label_y_pos = y_min + (y_range * 0.05)

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

            # --- Add duration label directly on the plot ---
            event_mid_x = (row["rel_start_time"] + row["rel_end_time"]) / 2
            duration_text = f"{row['duration_s']:.1f}s"
            ax1.text(
                event_mid_x,
                label_y_pos,
                duration_text,
                color=event_color,
                ha="center",
                va="bottom",
                fontsize=8,
                weight="bold",
                bbox=dict(
                    facecolor="white",
                    alpha=0.7,
                    edgecolor="none",
                    boxstyle="round,pad=0.2",
                ),
            )
            app_labels_seen.add(app_id)

    print("Reconfiguration events visualized as shaded regions with duration labels.")


# ---- Labeling ----
ax1.set_xlabel("Time (seconds from start)")
ax1.set_ylabel("Latency (ms)", color="tab:red")
ax2.set_ylabel(f"RPS (avg per {granularity}s)", color="tab:blue")
plt.title("End-to-End Latency vs Average RPS")

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
ax1.legend(
    unique_lines,
    unique_labels,
    loc="upper left",
)

fig.tight_layout()

# ---- Statistical Analysis ----
latency_data = df["true_latency_ms"]

# 1. Total Average Latency (Mean)
mean_latency = latency_data.mean()

# 2. Spread (Standard Deviation and IQR)
std_latency = latency_data.std()
q25 = latency_data.quantile(0.25)
q75 = latency_data.quantile(0.75)
iqr_latency = q75 - q25

# 3. Percentage under Target Latency
percentage_under_target = None
if target_latency is not None:
    count_under_target = (latency_data < target_latency).sum()
    total_count = len(latency_data)
    percentage_under_target = (count_under_target / total_count) * 100


def generate_stats_output(
    df,
    mean_latency,
    std_latency,
    iqr_latency,
    q25,
    q75,
    target_latency,
    percentage_under_target,
) -> str:
    """Generates the formatted string output for performance statistics."""

    output = []
    output.append("=" * 50)
    output.append("           PERFORMANCE STATISTICS")
    output.append("=" * 50)
    output.append(f"Total Requests Analyzed: {len(df)}")
    output.append(f"Overall Mean Latency:    {mean_latency:.2f} ms")
    output.append("-" * 50)
    output.append("Latency Spread Measures:")
    output.append(f"  Standard Deviation:    {std_latency:.2f} ms")
    output.append(f"  Interquartile Range:   {iqr_latency:.2f} ms")
    output.append(f"  25th Percentile (Q1):  {q25:.2f} ms")
    output.append(f"  75th Percentile (Q3):  {q75:.2f} ms")

    if percentage_under_target is not None:
        output.append("-" * 50)
        output.append(f"Target Latency ({target_latency} ms):")
        output.append(f"  % Under Target:        {percentage_under_target:.2f} %")
    output.append("=" * 50)

    return "\n".join(output)


stats_output = generate_stats_output(
    df,
    mean_latency,
    std_latency,
    iqr_latency,
    q25,
    q75,
    target_latency,
    percentage_under_target,
)
print(stats_output)


def save_plot_to_file(fig, run_folder: str, filename="result.png"):
    """Saves the figure to the run-specific folder."""
    filepath = os.path.join(run_folder, filename)

    try:
        fig.savefig(filepath)
        print(f"\n✅ Plot successfully saved to: {filepath}")
        return filepath
    except Exception as e:
        print(f"\n❌ Error saving figure to {filepath}: {e}")
        return None


def save_stats_to_file(run_folder: str, stats_data: str, filename="result.txt"):
    """Saves the statistics to the run-specific folder."""
    filepath = os.path.join(run_folder, filename)

    try:
        with open(filepath, "w") as f:
            f.write(stats_data)
        print(f"✅ Statistics successfully saved to: {filepath}")
        return filepath
    except Exception as e:
        print(f"❌ Error saving statistics to {filepath}: {e}")
        return None


def copy_csv_files(run_folder: str, csv_path: str, events_path: Optional[str]):
    """Copies the input CSV files to the run-specific folder."""
    # Copy results CSV
    if os.path.exists(csv_path):
        results_csv_dest = os.path.join(run_folder, "results.csv")
        try:
            shutil.copy2(csv_path, results_csv_dest)
            print(f"✅ Results CSV copied to: {results_csv_dest}")
        except Exception as e:
            print(f"❌ Error copying results CSV: {e}")

    # Copy reconfig events CSV if provided
    if events_path and os.path.exists(events_path):
        events_csv_dest = os.path.join(run_folder, "reconfig_events.csv")
        try:
            shutil.copy2(events_path, events_csv_dest)
            print(f"✅ Reconfig events CSV copied to: {events_csv_dest}")
        except Exception as e:
            print(f"❌ Error copying reconfig events CSV: {e}")


plot_filename = save_plot_to_file(fig, RUN_FOLDER)
stats_filename = save_stats_to_file(RUN_FOLDER, stats_output)
copy_csv_files(RUN_FOLDER, file_path_csv, file_path_events)

plt.show()
