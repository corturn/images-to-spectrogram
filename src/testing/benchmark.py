import subprocess
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
from typing import List, Dict

THREADS = [2,4,6,8,12]
SIZES = ["small.json", "mix.json", "large.json"]
NUM_RUNS = 3

def main():
    print("Starting benchmark")
    print("- - - - - -")

    seq: Dict[str, float] = run_seq()
    print(seq)
    map_graph_data = run_par(seq, "map")
    print(map_graph_data)
    generate_graph(map_graph_data, "map_reduce_speedups")
    steal_graph_data = run_par(seq, "steal")
    print(steal_graph_data)
    generate_graph(steal_graph_data, "work_steal_speedups")

    print("- - - - - -")
    print("Done :)")

def run_seq():
    print("...Running sequential...")
    seq = {}
    for size in SIZES:
        print(f"    file: {size}")

        cmd = ["go", "run", "testingio.go", size, "seq"]
        average_runtime = run_process(cmd, NUM_RUNS)
        seq[size] = average_runtime
    return seq


def run_par(seq: Dict[str, float], mapsteal: str) -> Dict[str, List[float]]:
    print(f"...Running {mapsteal}...")

    graph_data: Dict[str, List[float]] = {}
    for size in SIZES:
        print(f"    file: {size}")
        speedups = []
        for thread_count in THREADS:
            print(f"        threads: {thread_count}")
            cmd = ["go", "run", "testingio.go", size, mapsteal, f"{thread_count}"]
            average_runtime = run_process(cmd, NUM_RUNS)
            # Speedup calculation for that size and thread count
            speedup = seq[size] / average_runtime
            speedups.append(speedup)
        graph_data[size] = speedups
    return graph_data

def generate_graph(graph_data, name):
    print(f"...Generating graph {name}...")
    plt.clf()
    for size in SIZES:
        plt.plot(THREADS, graph_data[size], label = size)
    plt.legend()
    plt.xlabel("Number of Threads")
    plt.ylabel("Speedup")
    plt.title(f"{name}")
    plt.savefig(f"./{name}.png")


def run_process(cmd, num_runs):
    # Runs command specified num times and returns average runtime
    outputs = []
    for _ in range(num_runs):
        output = subprocess.run(cmd, capture_output=True)
        if output.stderr:
            raise RuntimeError("running testingio.go failed")
        outputs.append(float(output.stdout.decode("utf-8")))
    return sum(outputs) / len(outputs)

if __name__ == "__main__":
    main()