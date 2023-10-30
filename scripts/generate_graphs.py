import matplotlib.pyplot as plt
import numpy as np
import pandas as pd

algorithms = ["phash", "sift", "orb", "brisk"]

scenarios = \
    ["identical", "scaled", "moved", "background", "rotated", "mirrored", "part", "mixed"]

scenarios_desc = \
    ["Identisch", "Skaliert", "Verschoben", "Hintergrund", "Rotiert", "Gespiegelt", "Teilmotiv", "Gemischt"]

recall_list_phash = []
specificity_list_phash = []
accuracy_list_phash = []

recall_list_hybrid = []
specificity_list_hybrid = []
accuracy_list_hybrid = []


def get_phash_and_hybrid_data():
    for scenario in scenarios:
        csv_phash = pd.read_csv(f"../test-output/csv-files/phash/{scenario}-overall-evaluation.csv", index_col=0)
        csv_phash = csv_phash.sort_values(by="threshold")

        standard_phash_evaluation = csv_phash.loc[4]
        recall_list_phash.append(standard_phash_evaluation["recall"])
        specificity_list_phash.append(standard_phash_evaluation["specificity"])
        accuracy_list_phash.append(standard_phash_evaluation["balanced accuracy"])

        csv_hybrid = (
            pd.read_csv(f"../test-output/csv-files/hybrid/{scenario}-hybrid-overall-evaluation.csv", index_col=0))
        csv_hybrid = csv_hybrid.sort_values(by="threshold")

        standard_hybrid_evaluation = csv_hybrid.loc[0]
        recall_list_hybrid.append(standard_hybrid_evaluation["recall"])
        specificity_list_hybrid.append(standard_hybrid_evaluation["specificity"])
        accuracy_list_hybrid.append(standard_hybrid_evaluation["balanced accuracy"])


def generate_detail_graphs():
    for algorithm in algorithms:
        matcher = ""
        if algorithm in ["sift", "orb", "brisk"]:
            matcher += "-bfm"

        scenario_index = 0
        for scenario in scenarios:

            csv = pd.read_csv(f"../test-output/csv-files/{algorithm}/{scenario}{matcher}-overall-evaluation.csv")
            csv = csv.sort_values(by="threshold")

            threshold_values = csv["threshold"]
            recall_values = csv["recall"]
            specificity_values = csv["specificity"]
            accuracy_values = csv["balanced accuracy"]

            plt.title(f"{scenarios_desc[scenario_index]}")
            scenario_index += 1

            plt.plot(threshold_values, recall_values, label="Recall")
            plt.plot(threshold_values, specificity_values, label="Spezifität")
            plt.plot(threshold_values, accuracy_values, label="Balancierte-Genauigkeit", linestyle="--")

            plt.ylim(-0.05, 1.12)

            if algorithm == "phash":
                plt.xticks(threshold_values)
                plt.xlabel("Grenzwerte Hamming Distanz")
            else:
                plt.xlabel("Grenzwerte Bildähnlichkeits-Score")

            plt.ylabel("Zuverlässigkeitsbewertung")
            plt.legend(loc='upper left', ncol=3, frameon=False)

            plt.savefig(f'../test-output/graphs/{algorithm}/{scenario}.png')

            plt.clf()


def generate_general_phash_graph():
    bar_width = 0.25
    bar_positions = np.arange(len(scenarios))

    plt.bar(bar_positions - bar_width, recall_list_phash, bar_width, label='Recall')
    plt.bar(bar_positions, specificity_list_phash, bar_width, label='Spezifität')
    plt.bar(bar_positions + bar_width, accuracy_list_phash, bar_width, label='Balancierte-Genauigkeit')

    plt.ylim(-0.05, 1.12)
    plt.ylabel("Zuverlässigkeitsbewertung")
    plt.xticks(bar_positions, scenarios_desc, rotation=45)
    plt.legend(loc='upper left', ncol=3, frameon=False)

    plt.subplots_adjust(bottom=0.175, top=0.8)
    plt.savefig(f'../test-output/graphs/phash/standard.png')
    plt.clf()


def generate_comparison_chart():
    bar_width = 0.125
    bar_offset = 0.3
    half_width = bar_width / 2
    bar_positions = np.arange(len(scenarios))
    alpha = 0.5

    plt.bar(bar_positions - bar_offset + half_width, recall_list_hybrid, bar_width, label='Recall', color='blue')
    plt.bar(bar_positions - bar_offset - half_width, recall_list_phash, bar_width, color='blue', alpha=alpha)

    plt.bar([0], label="pHash", height=0, color="silver")

    plt.bar(bar_positions + half_width, specificity_list_hybrid, bar_width, label='Spezifität', color='orange')
    plt.bar(bar_positions - half_width, specificity_list_phash, bar_width, color='orange', alpha=alpha)

    plt.bar([0], label="Neues System", height=0, color="grey")

    plt.bar(bar_positions + bar_offset + half_width, accuracy_list_hybrid, bar_width, label='Balancierte-Genauigkeit',
            color='green')
    plt.bar(bar_positions + bar_offset - half_width, accuracy_list_phash, bar_width, color='green', alpha=alpha)

    plt.ylim(-0.05, 1.19)
    plt.ylabel("Zuverlässigkeitsbewertung")
    plt.xticks(bar_positions, scenarios_desc, rotation=45)
    plt.legend(loc='upper left', ncol=3, frameon=False)

    plt.subplots_adjust(bottom=0.175, top=0.95)
    plt.savefig(f'../test-output/graphs/phash/comparison.png')
    plt.show()
    plt.clf()


def main():
    # generate_detail_graphs()
    get_phash_and_hybrid_data()
    # generate_general_phash_graph()
    generate_comparison_chart()


if __name__ == "__main__":
    main()
