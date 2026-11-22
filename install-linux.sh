#!/bin/sh
set -eu

if [ "$(uname -s)" != Linux ]; then
    echo "This installer only supports Linux." >&2
    exit 1
fi

if ! command -v go >/dev/null 2>&1; then
    echo "Go is required to build Pac-Man; see the README for prerequisites." >&2
    exit 1
fi

case ${HOME:-} in
    /*) ;;
    *) echo "HOME must be an absolute path." >&2; exit 1 ;;
esac

case ${XDG_DATA_HOME:-} in
    /*) data_home=$XDG_DATA_HOME ;;
    *) data_home=$HOME/.local/share ;;
esac

if printf '%s%s' "$HOME" "$data_home" | LC_ALL=C grep -q '[[:cntrl:]]'; then
    echo "The installation path cannot contain control characters." >&2
    exit 1
fi

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
temp_dir=$(mktemp -d)
trap 'rm -rf -- "$temp_dir"' EXIT

host_arch=$(go env GOHOSTARCH)
echo "Building Pac-Man..."
(cd "$project_dir" && GOOS=linux GOARCH="$host_arch" go build -o "$temp_dir/pacman" .)

bin_dir=$HOME/.local/bin
desktop_dir=$data_home/applications
executable=$bin_dir/pacman

# Use a shell with a positional argument so desktop launchers can handle '%' in
# the executable path; the path itself is never interpreted as shell code.
# Exec values have both desktop-entry string escaping and command-line quoting.
exec_path=$(printf '%s' "$executable" | sed 's/\\/\\\\\\\\/g; s/"/\\\\"/g; s/\$/\\\\$/g; s/`/\\\\`/g; s/%/%%/g')
cat > "$temp_dir/geckoblu-pacman.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=Pac-Man
Comment=Play Pac-Man
EOF
printf '%s%s%s\n' 'Exec=sh -c "exec \\"\\$1\\"" sh "' "$exec_path" '"' >> "$temp_dir/geckoblu-pacman.desktop"
cat >> "$temp_dir/geckoblu-pacman.desktop" <<EOF
Icon=geckoblu-pacman
Terminal=false
Categories=Game;ArcadeGame;
EOF

install -d "$bin_dir" "$desktop_dir"
install -m 755 "$temp_dir/pacman" "$executable"
for size in 16 32 64 128; do
    icon_dir=$data_home/icons/hicolor/${size}x${size}/apps
    install -d "$icon_dir"
    install -m 644 "$project_dir/assets/icon_$size.png" "$icon_dir/geckoblu-pacman.png"
done
install -m 644 "$temp_dir/geckoblu-pacman.desktop" "$desktop_dir/geckoblu-pacman.desktop"

echo "Pac-Man installed. Launch it from your desktop application menu."
