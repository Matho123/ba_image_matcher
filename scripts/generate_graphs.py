import matplotlib.pyplot as plt
import numpy as np
import pandas as pd

algorithms = ["phash", "sift", "orb", "brisk"]

scenarios = ["identical", "scaled", "moved", "background", "rotated", "mirrored", "part", "mixed"]
scenarios_desc = ["Identisch", "Skaliert", "Verschoben", "Hintergrund", "Rotiert", "Gespiegelt", "Teilmotiv",
                  "Gemischt"]

for algorithm in algorithms:
    algorithm_file_name = algorithm
    if algorithm in ["sift", "orb", "brisk"]:
        algorithm_file_name += "-bfm"

    scenario_index = 0
    for scenario in scenarios:

        csv = pd.read_csv(f"../test-output/csv-files/{scenario}/{algorithm_file_name}-overall-evaluation.csv")
        csv = csv.sort_values(by="threshold")

        threshold_values = csv["threshold"]
        recall_values = csv["recall"]
        specificity_values = csv["specificity"]
        accuracy_values = csv["balanced accuracy"]

        plt.title(f"{scenarios_desc[scenario_index]}")
        scenario_index += 1

        plt.plot(threshold_values, recall_values, label="Recall")
        plt.plot(threshold_values, specificity_values, label="Specificity")
        plt.plot(threshold_values, accuracy_values, label="Balanced Accuracy", linestyle="--")

        plt.ylim(-0.05, 1.12)

        if algorithm == "phash":
            plt.xticks(threshold_values)
            plt.xlabel("Grenzwerte Hamming Distanz")
        else:
            plt.xlabel("Grenzwerte Bildähnlichkeits-Score")

        plt.ylabel("Zuverlässigkeits Bewertung")
        plt.legend(loc='upper left', ncol=3, frameon=False)

        plt.savefig(f'../test-output/graphs/{algorithm}/{scenario}.png')

        plt.clf()

# recall_list = []
# specificity_list = []
# accuracy_list = []
#
# for scenario in scenarios:
#     csv = pd.read_csv(f"../test-output/csv-files/{scenario}/phash-overall-evaluation.csv")
#     csv = csv.sort_values(by="threshold")
#
#     standard_phash_evaluation = csv[csv["threshold"] == 4]
#     recall_list.append(standard_phash_evaluation["recall"][0])
#     specificity_list.append(standard_phash_evaluation["specificity"][0])
#     accuracy_list.append(standard_phash_evaluation["balanced accuracy"][0])
#
# bar_width = 0.25
# bar_positions = np.arange(len(scenarios))
#
# plt.bar(bar_positions, recall_list, bar_width, label='Recall')
# plt.bar(bar_positions, specificity_list, bar_width, label='Specificity')
# plt.bar(bar_positions, accuracy_list, bar_width, label='Balanced Accuracy')
#
#
# plt.xticks(bar_positions, scenarios_desc, rotation=45)
# plt.legend(loc='upper right', bbox_to_anchor=(1, 1.3))
# plt.subplots_adjust(bottom=0.175, top=0.8)
#
# plt.savefig(f'../test-output/graphs/phash/standard.png', )
