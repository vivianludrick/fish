"""
Doench Rule Set 2 (Azimuth) on-target scoring HTTP server.

Uses the retrained Python 3 model from CrisprScoringHub:
https://github.com/Interventional-Genomics-Unit/CrisprScoringHub

Input:  30-mer sequences (4nt upstream + 20nt guide + 3nt PAM + 3nt downstream)
Output: Scores 0-100 (higher = better on-target activity)

Endpoints:
    POST /score  {"sequences": ["AAAA...30nt..."]}
    GET  /health
"""

import os
import sys
import pickle
import numpy as np
import pandas as pd
from flask import Flask, request, jsonify

# Add CrisprScoringHub models dir to path for featurization import
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
SCORING_HUB = os.path.join(SCRIPT_DIR, "CrisprScoringHub", "models")
sys.path.insert(0, SCORING_HUB)

from featurization import featurize_data

app = Flask(__name__)

# Load model once at startup
MODEL_PATH = os.path.join(SCORING_HUB, "python3_V3_model_no.pos.pickle")
with open(MODEL_PATH, "rb") as f:
    model, learn_options = pickle.load(f)

learn_options["V"] = 2


def score_sequences(sequences: list[str]) -> list[float]:
    """Score 30-mer sequences using Doench Rule Set 2."""
    processed = []
    invalid_indices = set()

    for i, seq in enumerate(sequences):
        seq = seq.upper()
        if "N" in seq:
            invalid_indices.add(i)
            processed.append("A" * 30)  # placeholder
        else:
            # Enforce GG PAM at positions 25-26
            seq_list = list(seq)
            seq_list[25] = "G"
            seq_list[26] = "G"
            processed.append("".join(seq_list))

    seqs = np.array(processed)

    xdf = pd.DataFrame(
        columns=["30mer", "Strand"],
        data=zip(seqs, np.repeat("NA", len(seqs))),
    )

    feature_sets = featurize_data(data=xdf, learn_options=learn_options, Y=pd.DataFrame())

    inputs = np.hstack([feature_sets[k].values for k in feature_sets.keys()])
    raw_scores = model.predict(inputs)

    scores = []
    for i, s in enumerate(raw_scores):
        if i in invalid_indices:
            scores.append(-1.0)
        else:
            scores.append(round(float(s) * 100, 2))

    return scores


@app.route("/health", methods=["GET"])
def health():
    return jsonify({"status": "ok"})


@app.route("/score", methods=["POST"])
def score():
    data = request.get_json()
    if not data:
        return jsonify({"error": "request body required"}), 400

    sequences = data.get("sequences", [])
    if not sequences:
        return jsonify({"error": "no sequences provided"}), 400

    for seq in sequences:
        if len(seq) != 30:
            return jsonify({"error": f"sequence must be 30nt, got {len(seq)}: {seq}"}), 400

    scores = score_sequences(sequences)

    results = {}
    for seq, sc in zip(sequences, scores):
        results[seq] = sc

    return jsonify({"scores": results})


if __name__ == "__main__":
    port = int(os.environ.get("DOENCH_PORT", 8080))
    print(f"Doench scoring server starting on port {port}")
    app.run(host="127.0.0.1", port=port)
