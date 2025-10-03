#!/bin/bash

# Sicherstellen, dass im richtigen Verzeichnis
echo "Aktuelles Verzeichnis: $(pwd)"

echo
echo "1. Zeige Remotes:"
git remote -v

echo
echo "2. Zeige Branches:"
git branch

echo
echo "3. Hole neueste Infos von origin:"
git fetch origin

echo
echo "4. Pulle und rebase main:"
git pull origin main

echo
echo "Fertig!"
