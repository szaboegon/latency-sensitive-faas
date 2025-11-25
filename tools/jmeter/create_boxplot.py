# type: ignore
import pandas as pd
import matplotlib.pyplot as plt
import os
import sys
from pathlib import Path
from typing import List, Dict, Tuple

RESULTS_FOLDER = "results"
OUTPUT_FOLDER = "box-plot"


def find_result_folders(results_folder: str) -> List[Tuple[str, str, str]]:
    """
    Finds all test run folders that contain both results.csv (JMeter) and redis_results.csv.
    Returns a list of tuples (folder_name, jtl_path, redis_path).
    """
    run_folders = []
    results_path = Path(results_folder)

    if not results_path.exists():
        print(f"Error: Results folder '{results_folder}' does not exist.")
        return run_folders

    # Iterate through subdirectories
    for subfolder in results_path.iterdir():
        if subfolder.is_dir():
            jtl_path = subfolder / "results.csv"
            redis_path = subfolder / "redis_results.csv"

            if jtl_path.exists() and redis_path.exists():
                run_folders.append((subfolder.name, str(jtl_path), str(redis_path)))
                print(f"Found: {subfolder.name}")
            elif jtl_path.exists():
                print(
                    f"Warning: {subfolder.name} has results.csv but no redis_results.csv. Skipping."
                )
            elif redis_path.exists():
                print(
                    f"Warning: {subfolder.name} has redis_results.csv but no results.csv. Skipping."
                )

    return run_folders


def load_latency_data(run_folders: List[Tuple[str, str, str]]) -> Dict[str, pd.Series]:
    """
    Loads latency data by merging JMeter and Redis results.
    Calculates true end-to-end latency as: redis_end_time - jmeter_start_time
    Returns a dictionary mapping run names to their latency data.
    """
    latency_data = {}

    for run_name, jtl_path, redis_path in run_folders:
        try:
            # Load JMeter results
            df_jtl = pd.read_csv(jtl_path)

            # Load Redis results
            df_redis = pd.read_csv(redis_path)

            # Check required columns in JTL
            if (
                "timeStamp" not in df_jtl.columns
                or "correlation_id" not in df_jtl.columns
            ):
                print(
                    f"Warning: {run_name} - results.csv missing 'timeStamp' or 'correlation_id'. Skipping."
                )
                continue

            # Check required columns in Redis results
            if (
                "correlation_id" not in df_redis.columns
                or "end_time_ms" not in df_redis.columns
            ):
                print(
                    f"Warning: {run_name} - redis_results.csv missing 'correlation_id' or 'end_time_ms'. Skipping."
                )
                continue

            # Ensure correlation_id is string type for both
            df_jtl["correlation_id"] = df_jtl["correlation_id"].astype(str)
            df_redis["correlation_id"] = df_redis["correlation_id"].astype(str)

            # Keep only necessary columns
            df_jtl_subset = df_jtl[["correlation_id", "timeStamp"]].copy()
            df_redis_subset = df_redis[["correlation_id", "end_time_ms"]].copy()

            # Merge on correlation_id
            df_merged = pd.merge(
                df_jtl_subset, df_redis_subset, on="correlation_id", how="inner"
            )

            if df_merged.empty:
                print(
                    f"Warning: {run_name} - no matching correlation_ids between JMeter and Redis. Skipping."
                )
                continue

            # Calculate true end-to-end latency: redis_end_time - jmeter_start_time
            df_merged["true_latency_ms"] = (
                df_merged["end_time_ms"] - df_merged["timeStamp"]
            )

            # Store latency data
            latency_data[run_name] = df_merged["true_latency_ms"]
            print(f"Loaded {len(df_merged)} samples from {run_name}")

        except Exception as e:
            print(f"Error loading {run_name}: {e}")

    return latency_data


def calculate_statistics(
    latency_data: Dict[str, pd.Series],
) -> Dict[str, Dict[str, float]]:
    """
    Calculates statistical measures for each run and combined data.
    Returns a dictionary with min, q1, median, q3, max for each run.
    """
    stats = {}

    for run_name, data in latency_data.items():
        stats[run_name] = {
            "min": data.min(),
            "q1": data.quantile(0.25),
            "median": data.median(),
            "q3": data.quantile(0.75),
            "max": data.max(),
            "mean": data.mean(),
            "std": data.std(),
            "count": len(data),
        }

    # Add combined statistics
    all_data = pd.concat(latency_data.values(), ignore_index=True)
    stats["_COMBINED"] = {
        "min": all_data.min(),
        "q1": all_data.quantile(0.25),
        "median": all_data.median(),
        "q3": all_data.quantile(0.75),
        "max": all_data.max(),
        "mean": all_data.mean(),
        "std": all_data.std(),
        "count": len(all_data),
    }

    return stats


def create_boxplot(latency_data: Dict[str, pd.Series], output_folder: str):
    """
    Creates a single box plot visualization combining latency data from all runs.
    """
    if not latency_data:
        print("No data to plot.")
        return None

    # Combine all latency data from all runs into a single series
    all_latencies = pd.concat(latency_data.values(), ignore_index=True)

    print(f"Combined {len(all_latencies)} samples from {len(latency_data)} test run(s)")

    # Set font sizes to match plot_rps_latency
    plt.rcParams.update(
        {
            "font.size": 14,
            "axes.labelsize": 16,
            "axes.titlesize": 18,
            "xtick.labelsize": 14,
            "ytick.labelsize": 14,
            "legend.fontsize": 14,
        }
    )

    # Create figure - wider for horizontal boxplot
    fig, ax = plt.subplots(figsize=(12, 6))

    # Create horizontal box plot with single box
    bp = ax.boxplot(
        [all_latencies],
        labels=[""],  # Empty label
        patch_artist=True,
        showmeans=True,
        meanprops=dict(
            marker="D",
            markerfacecolor="#FF6B6B",
            markeredgecolor="#C92A2A",
            markersize=10,
            linewidth=2,
        ),
        medianprops=dict(color="#2F9E44", linewidth=2.5),
        whiskerprops=dict(color="#495057", linewidth=1.5),
        capprops=dict(color="#495057", linewidth=1.5),
        flierprops=dict(
            marker="o",
            markerfacecolor="#868E96",
            markersize=6,
            linestyle="none",
            markeredgecolor="#495057",
            alpha=0.5,
        ),
        vert=False,  # Make horizontal
    )

    # Customize box plot colors with a nice gradient-like appearance
    for patch in bp["boxes"]:
        patch.set_facecolor("#4DABF7")
        patch.set_edgecolor("#1971C2")
        patch.set_linewidth(2)
        patch.set_alpha(0.7)

    # Labels and title - swap x and y for horizontal orientation
    ax.set_ylabel("")  # Remove y-axis label
    ax.set_yticks([])  # Remove y-axis ticks
    ax.set_xlabel("Latency (ms)", fontsize=16)

    # Set x-axis ticks at 500, 1500, 2500, etc.
    import numpy as np

    max_latency = all_latencies.max()
    x_ticks = np.arange(500, max_latency + 1000, 1000)
    ax.set_xticks(x_ticks)
    ax.tick_params(axis="x", labelsize=14)

    ax.set_title("Latency Distribution (All Test Runs Combined)", fontsize=18)
    ax.grid(True, linestyle="--", alpha=0.3, axis="x")

    # Add legend explaining the box plot elements
    from matplotlib.patches import Patch
    from matplotlib.lines import Line2D

    legend_elements = [
        Patch(
            facecolor="#4DABF7",
            edgecolor="#1971C2",
            alpha=0.7,
            label="Box (Q1 to Q3)",
        ),
        Line2D([0], [0], color="#2F9E44", linewidth=2.5, label="Median"),
        Line2D(
            [0],
            [0],
            marker="D",
            color="w",
            markerfacecolor="#FF6B6B",
            markeredgecolor="#C92A2A",
            markersize=10,
            linewidth=0,
            label="Average",
        ),
        Line2D([0], [0], color="#495057", linewidth=1.5, label="Whiskers (1.5×IQR)"),
        Line2D(
            [0],
            [0],
            marker="o",
            color="w",
            markerfacecolor="#868E96",
            markeredgecolor="#495057",
            markersize=6,
            alpha=0.5,
            linewidth=0,
            label="Outliers",
        ),
    ]

    ax.legend(
        handles=legend_elements,
        loc="upper right",
        fontsize=14,
        framealpha=0.95,
        edgecolor="#DEE2E6",
    )

    plt.tight_layout()

    # Save plot
    os.makedirs(output_folder, exist_ok=True)
    plot_path = os.path.join(output_folder, "boxplot.png")

    try:
        fig.savefig(plot_path, dpi=300, bbox_inches="tight")
        print(f"\n✅ Box plot saved to: {plot_path}")
        return plot_path
    except Exception as e:
        print(f"\n❌ Error saving box plot: {e}")
        return None


def save_statistics(stats: Dict[str, Dict[str, float]], output_folder: str):
    """
    Saves statistical measures to a text file.
    """
    os.makedirs(output_folder, exist_ok=True)
    stats_path = os.path.join(output_folder, "boxplot_statistics.txt")

    try:
        with open(stats_path, "w") as f:
            f.write("=" * 80 + "\n")
            f.write(" " * 20 + "LATENCY STATISTICS SUMMARY\n")
            f.write("=" * 80 + "\n\n")

            # Show combined statistics first
            if "_COMBINED" in stats:
                s = stats["_COMBINED"]
                f.write("COMBINED - All Test Runs\n")
                f.write("=" * 80 + "\n")
                f.write(f"  Sample Count:      {s['count']}\n")
                f.write(f"  Minimum:           {s['min']:.2f} ms\n")
                f.write(f"  Q1 (25th %ile):    {s['q1']:.2f} ms\n")
                f.write(f"  Median (50th %ile): {s['median']:.2f} ms\n")
                f.write(f"  Q3 (75th %ile):    {s['q3']:.2f} ms\n")
                f.write(f"  Maximum:           {s['max']:.2f} ms\n")
                f.write(f"  Mean:              {s['mean']:.2f} ms\n")
                f.write(f"  Std Deviation:     {s['std']:.2f} ms\n")
                f.write(f"  IQR (Q3-Q1):       {s['q3'] - s['q1']:.2f} ms\n")
                f.write("\n\n")

            # Then show individual run statistics
            f.write("INDIVIDUAL TEST RUNS\n")
            f.write("=" * 80 + "\n\n")

            # Sort by run name, skip _COMBINED
            for run_name in sorted(stats.keys()):
                if run_name == "_COMBINED":
                    continue
                s = stats[run_name]
                f.write(f"Test Run: {run_name}\n")
                f.write("-" * 80 + "\n")
                f.write(f"  Sample Count:      {s['count']}\n")
                f.write(f"  Minimum:           {s['min']:.2f} ms\n")
                f.write(f"  Q1 (25th %ile):    {s['q1']:.2f} ms\n")
                f.write(f"  Median (50th %ile): {s['median']:.2f} ms\n")
                f.write(f"  Q3 (75th %ile):    {s['q3']:.2f} ms\n")
                f.write(f"  Maximum:           {s['max']:.2f} ms\n")
                f.write(f"  Mean:              {s['mean']:.2f} ms\n")
                f.write(f"  Std Deviation:     {s['std']:.2f} ms\n")
                f.write(f"  IQR (Q3-Q1):       {s['q3'] - s['q1']:.2f} ms\n")
                f.write("\n")

            f.write("=" * 80 + "\n")

        print(f"✅ Statistics saved to: {stats_path}")
        return stats_path
    except Exception as e:
        print(f"❌ Error saving statistics: {e}")
        return None


def main():
    print("=" * 80)
    print(" " * 25 + "Box Plot Generator")
    print("=" * 80)
    print()

    # Find all result folders with both JMeter and Redis data
    print(f"Scanning '{RESULTS_FOLDER}' folder for test results...\n")
    run_folders = find_result_folders(RESULTS_FOLDER)

    if not run_folders:
        print(f"\nNo valid test runs found in '{RESULTS_FOLDER}' subfolders.")
        print(
            "Each run folder must contain both 'results.csv' and 'redis_results.csv'."
        )
        sys.exit(1)

    print(f"\nFound {len(run_folders)} test run(s).\n")

    # Load latency data
    print("Loading and calculating latency data...\n")
    latency_data = load_latency_data(run_folders)

    if not latency_data:
        print("\nNo valid latency data could be loaded.")
        sys.exit(1)

    # Calculate statistics
    print("\nCalculating statistics...\n")
    stats = calculate_statistics(latency_data)

    # Display statistics
    print("\nCombined Statistics:")
    if "_COMBINED" in stats:
        s = stats["_COMBINED"]
        print(
            f"All Runs: Min={s['min']:.2f}ms, Q1={s['q1']:.2f}ms, "
            f"Median={s['median']:.2f}ms, Q3={s['q3']:.2f}ms, Max={s['max']:.2f}ms, "
            f"Mean={s['mean']:.2f}ms, Samples={s['count']}"
        )

    print("\nIndividual Run Statistics:")
    for run_name in sorted(stats.keys()):
        if run_name == "_COMBINED":
            continue
        s = stats[run_name]
        print(
            f"{run_name}: Min={s['min']:.2f}ms, Q1={s['q1']:.2f}ms, "
            f"Median={s['median']:.2f}ms, Q3={s['q3']:.2f}ms, Max={s['max']:.2f}ms"
        )

    # Create box plot
    print("\nGenerating box plot...")
    plot_path = create_boxplot(latency_data, OUTPUT_FOLDER)

    # Save statistics to file
    print("\nSaving statistics...")
    stats_path = save_statistics(stats, OUTPUT_FOLDER)

    print("\n" + "=" * 80)
    print(" " * 30 + "Complete!")
    print("=" * 80)

    if plot_path and stats_path:
        print(f"\nOutputs saved to '{OUTPUT_FOLDER}/' folder:")
        print(f"  - {os.path.basename(plot_path)}")
        print(f"  - {os.path.basename(stats_path)}")

    # Show plot
    plt.show()


if __name__ == "__main__":
    main()
