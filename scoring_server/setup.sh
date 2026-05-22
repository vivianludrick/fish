#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "Setting up Doench scoring server..."

# Create venv if it doesn't exist
if [ ! -d "venv" ]; then
    python3 -m venv venv
fi
source venv/bin/activate

pip install -r requirements.txt

# Clone CrisprScoringHub for model + featurization
if [ ! -d "CrisprScoringHub" ]; then
    git clone --depth 1 https://github.com/Interventional-Genomics-Unit/CrisprScoringHub.git
fi

echo ""
echo "Setup complete. Start the server with:"
echo "  cd $SCRIPT_DIR && source venv/bin/activate && python server.py"
