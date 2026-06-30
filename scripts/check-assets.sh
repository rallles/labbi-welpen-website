#!/usr/bin/env sh
set -eu

# check-assets.sh
# Sucht /static/... Referenzen in internal/templates und static/css und prueft,
# ob die referenzierten Dateien im Repo existieren. Es wird nichts geloescht.
#
# Hinweis: ZIP-Arbeitsstaende werden manchmal absichtlich ohne Bilder ausgeliefert.
# Fehlende Bilder bedeuten daher nicht automatisch einen Fehler im Code; in einem
# vollstaendigen Arbeitsstand muessen die referenzierten Assets aber vorhanden sein.
#
# Optionen:
#   SKIP_ASSET_CHECK=1  ueberspringt den gesamten Check (Exit 0).

if [ "${SKIP_ASSET_CHECK:-0}" = "1" ]; then
  echo "==> Asset-Check uebersprungen (SKIP_ASSET_CHECK=1)."
  exit 0
fi

cd "$(dirname "$0")/.."

SEARCH_DIRS="internal/templates static/css"

echo "==> Asset-Check: suche /static/... Referenzen in: ${SEARCH_DIRS}"

existing_dirs=""
for d in $SEARCH_DIRS; do
  if [ -d "$d" ]; then
    existing_dirs="$existing_dirs $d"
  else
    echo "    Hinweis: Verzeichnis fehlt, wird uebersprungen: $d"
  fi
done

if [ -z "$existing_dirs" ]; then
  echo "Keine durchsuchbaren Verzeichnisse gefunden. Nichts zu pruefen."
  exit 0
fi

# Alle /static/... Referenzen extrahieren.
# Die Zeichenklasse stoppt an Leerzeichen, Anfuehrungszeichen, Kommas usw.,
# sodass srcset-Deskriptoren wie "480w" nicht mitgenommen werden.
refs="$(grep -rho -E '/static/[A-Za-z0-9_./-]+' $existing_dirs 2>/dev/null | sort -u || true)"

if [ -z "$refs" ]; then
  echo "Keine /static/... Referenzen gefunden. Nichts zu pruefen."
  exit 0
fi

missing=""
missing_count=0
checked_count=0

for ref in $refs; do
  checked_count=$((checked_count + 1))
  # /static/... -> static/...
  file="static/${ref#/static/}"
  if [ ! -f "$file" ]; then
    missing="${missing}\n  ${ref}  ->  ${file}"
    missing_count=$((missing_count + 1))
  fi
done

echo "    Geprueft: ${checked_count} referenzierte Dateien."

if [ "$missing_count" -eq 0 ]; then
  echo "==> Alle referenzierten Assets sind vorhanden."
  exit 0
fi

echo ""
echo "==> Fehlende Assets (${missing_count}):"
# shellcheck disable=SC2059
printf "${missing}\n"
echo ""
echo "Diese Dateien werden in internal/templates oder static/css referenziert,"
echo "fehlen aber im Repo. Es wurde nichts geloescht."
echo ""
echo "Hinweis: ZIP-Arbeitsstaende sind manchmal absichtlich ohne Bilder. Falls dies"
echo "ein bewusst reduzierter Arbeitsstand ist, kann der Check mit SKIP_ASSET_CHECK=1"
echo "uebersprungen werden. In einem vollstaendigen Arbeitsstand sollten die Dateien"
echo "vorhanden sein."

exit 1
