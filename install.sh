#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="vdroid-scripter"
PREFIX="${PREFIX:-/usr/local}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVER_DIR="$REPO_ROOT/server"
BUILD_OUTPUT=""

main() {
  detect_platform
  install_dependencies
  require_go
  build_server
  install_binary
  print_summary
}

log() { printf '==> %s\n' "$*"; }
warn() { printf 'WARNING: %s\n' "$*" >&2; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

version_ge() {
  [ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n1)" = "$2" ]
}

detect_platform() {
  local os
  os="$(uname -s)"
  case "$os" in
    Darwin) PLATFORM="macos" ;;
    Linux) PLATFORM="linux"; detect_linux_distro ;;
    *) die "unsupported operating system: $os" ;;
  esac
}

detect_linux_distro() {
  [ -r /etc/os-release ] || die "cannot read /etc/os-release to detect the distribution"
  . /etc/os-release
  DISTRO_ID="${ID:-}"
  DISTRO_LIKE="${ID_LIKE:-}"
}

install_dependencies() {
  log "Installing system dependencies"
  if [ "$PLATFORM" = "macos" ]; then
    install_macos
  else
    install_linux
  fi
}

install_macos() {
  need_cmd brew
  brew install go android-platform-tools ffmpeg opencv@4 tesseract tesseract-lang
  install_macos_legacy_tessdata
}

install_macos_legacy_tessdata() {
  local tessdata
  tessdata="$(brew --prefix)/share/tessdata"
  mkdir -p "$tessdata"
  log "Fetching legacy eng.traineddata for keyboard detection"
  curl -fL https://github.com/tesseract-ocr/tessdata/raw/main/eng.traineddata \
    -o "$tessdata/eng.traineddata"
}

install_linux() {
  case "$DISTRO_ID" in
    arch | manjaro | endeavouros) install_arch ;;
    debian | ubuntu | linuxmint | pop) install_apt ;;
    fedora) install_dnf ;;
    *) install_linux_by_family ;;
  esac
}

install_linux_by_family() {
  case "$DISTRO_LIKE" in
    *arch*) install_arch ;;
    *debian*) install_apt ;;
    *fedora* | *rhel*) install_dnf ;;
    *) die "unsupported Linux distribution: ${DISTRO_ID:-unknown}. Install Go, adb, ffmpeg, opencv and tesseract manually, then run: cd server && go build -o $BIN_NAME ./cmd" ;;
  esac
}

install_arch() {
  sudo pacman -S --needed --noconfirm go android-tools ffmpeg tesseract
  sudo pacman -S --needed --noconfirm $(pacman -Sl extra | grep tesseract-data | awk '{print $2}')
  install_arch_opencv4
}

install_arch_opencv4() {
  if pacman -Qi opencv4 >/dev/null 2>&1; then
    return
  fi
  if pacman -Si opencv4 >/dev/null 2>&1; then
    sudo pacman -S --needed --noconfirm opencv4
    return
  fi
  if command -v paru >/dev/null 2>&1; then
    paru -S --needed --noconfirm opencv4
    return
  fi
  if command -v yay >/dev/null 2>&1; then
    yay -S --needed --noconfirm opencv4
    return
  fi
  die "Arch's official 'opencv' package is now OpenCV 5, but gocv needs OpenCV 4 (pkg-config opencv4). Install 'opencv4' from the AUR (with paru or yay; the chaotic-aur repo carries it prebuilt), then re-run this script."
}

install_apt() {
  sudo apt-get update
  sudo apt-get install -y \
    golang-go adb ffmpeg pkg-config \
    libopencv-dev \
    libavcodec-dev libavutil-dev \
    tesseract-ocr libtesseract-dev libleptonica-dev tesseract-ocr-eng tesseract-ocr-rus
}

install_dnf() {
  sudo dnf install -y \
    golang android-tools ffmpeg-free pkgconf-pkg-config \
    opencv-devel \
    libavcodec-free-devel libavutil-free-devel \
    tesseract tesseract-devel leptonica-devel tesseract-langpack-eng tesseract-langpack-rus
}

require_go() {
  need_cmd go
  local required current
  required="$(awk '/^go /{print $2; exit}' "$SERVER_DIR/go.mod")"
  current="$(go env GOVERSION 2>/dev/null | sed 's/^go//')"
  [ -n "$current" ] || current="$(go version | awk '{print $3}' | sed 's/^go//')"
  if ! version_ge "$current" "$required"; then
    die "Go $required or newer is required, but found $current. Install a newer Go from https://go.dev/dl/ and re-run."
  fi
}

build_server() {
  log "Building $BIN_NAME"
  local tmpdir
  tmpdir="$(mktemp -d)"
  BUILD_OUTPUT="$tmpdir/$BIN_NAME"
  if [ "$PLATFORM" = "macos" ]; then
    export PKG_CONFIG_PATH="$(brew --prefix opencv@4)/lib/pkgconfig:$(brew --prefix ffmpeg)/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
  fi
  (cd "$SERVER_DIR" && go mod download && CGO_ENABLED=1 go build -o "$BUILD_OUTPUT" ./cmd)
}

install_binary() {
  local bindir="$PREFIX/bin"
  local dest="$bindir/$BIN_NAME"
  local sudo_cmd=""
  if [ ! -w "$bindir" ] && [ ! -w "$PREFIX" ]; then
    sudo_cmd="sudo"
  fi
  log "Installing to $dest"
  $sudo_cmd mkdir -p "$bindir"
  $sudo_cmd install -m 0755 "$BUILD_OUTPUT" "$dest"
  rm -rf "$(dirname "$BUILD_OUTPUT")"
}

print_summary() {
  printf '\n'
  log "Done. Installed $BIN_NAME to $PREFIX/bin"
  case ":$PATH:" in
    *":$PREFIX/bin:"*)
      printf 'Connect a device (adb devices), then run: %s\n' "$BIN_NAME"
      ;;
    *)
      warn "$PREFIX/bin is not on your PATH."
      printf 'Add it by appending this line to your shell profile:\n'
      printf '  export PATH="%s/bin:$PATH"\n' "$PREFIX"
      printf 'Then run: %s\n' "$BIN_NAME"
      ;;
  esac
}

main "$@"
