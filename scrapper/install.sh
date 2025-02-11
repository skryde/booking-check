#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status.
set -e

if [ ! -d ".venv" ]; then
    python3 -m venv .venv
fi

source ./.venv/bin/activate
pip install poetry
poetry install
