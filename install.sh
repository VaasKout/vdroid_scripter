#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="vdroid-scripter"
PREFIX="${PREFIX:-/usr/local}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVER_DIR="$REPO_ROOT/server"
DEPS_DIR="${VDROID_DEPS:-$REPO_ROOT/build/deps}"
BUILD_OUTPUT=""

OPENCV_VERSION="4.14.0"
OPENCV_MODULES="core,imgproc,imgcodecs,highgui,videoio,features2d,calib3d,objdetect,photo,video,dnn,flann"

main() {
  detect_platform
  install_dependencies
  require_go
  build_opencv
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
  brew install go android-platform-tools ffmpeg tesseract tesseract-lang cmake ninja
  install_macos_legacy_tessdata
}

install_macos_legacy_tessdata() {
  local tessdata
  tessdata="$(brew --prefix)/share/tessdata"
  mkdir -p "$tessdata"
  log "Replacing the tessdata_fast eng model with the full eng.traineddata for better OCR accuracy"
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
    *) die "unsupported Linux distribution: ${DISTRO_ID:-unknown}. Install Go, adb, ffmpeg, tesseract, cmake, ninja and a C++ compiler manually, then run this script's build steps by hand: see the README." ;;
  esac
}

install_arch() {
  sudo pacman -S --needed --noconfirm go android-tools ffmpeg tesseract cmake ninja gcc curl
  sudo pacman -S --needed --noconfirm $(pacman -Sl extra | grep tesseract-data | awk '{print $2}')
}

install_apt() {
  sudo apt-get update
  sudo apt-get install -y \
    golang-go adb ffmpeg pkg-config build-essential cmake ninja-build curl \
    libavcodec-dev libavutil-dev \
    tesseract-ocr libtesseract-dev libleptonica-dev tesseract-ocr-eng tesseract-ocr-rus
}

install_dnf() {
  sudo dnf install -y \
    golang android-tools ffmpeg-free pkgconf-pkg-config gcc-c++ cmake ninja-build curl \
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

build_opencv() {
  local stamp="$DEPS_DIR/stamps/opencv-$OPENCV_VERSION"
  if [ -f "$stamp" ] && [ -f "$DEPS_DIR/lib/pkgconfig/opencv4.pc" ]; then
    log "OpenCV $OPENCV_VERSION is already built in $DEPS_DIR"
    return
  fi
  need_cmd cmake
  need_cmd ninja
  need_cmd curl
  local src="$DEPS_DIR/src/opencv-$OPENCV_VERSION"
  local tarball="$DEPS_DIR/src/opencv-$OPENCV_VERSION.tar.gz"
  mkdir -p "$DEPS_DIR/src" "$DEPS_DIR/stamps"
  if [ ! -d "$src" ]; then
    log "Downloading OpenCV $OPENCV_VERSION"
    curl -fL "https://github.com/opencv/opencv/archive/refs/tags/$OPENCV_VERSION.tar.gz" -o "$tarball"
    tar -xzf "$tarball" -C "$DEPS_DIR/src"
  fi
  log "Building OpenCV $OPENCV_VERSION as static libraries in $DEPS_DIR (this takes several minutes)"
  cmake -S "$src" -B "$src/build" -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX="$DEPS_DIR" \
    -DBUILD_SHARED_LIBS=OFF \
    -DBUILD_LIST="$OPENCV_MODULES" \
    -DOPENCV_GENERATE_PKGCONFIG=ON \
    -DOPENCV_ENABLE_NONFREE=OFF \
    -DBUILD_TESTS=OFF -DBUILD_PERF_TESTS=OFF -DBUILD_EXAMPLES=OFF -DBUILD_DOCS=OFF \
    -DBUILD_opencv_apps=OFF -DBUILD_JAVA=OFF -DBUILD_opencv_python2=OFF -DBUILD_opencv_python3=OFF \
    -DBUILD_ZLIB=ON -DBUILD_PNG=ON -DBUILD_JPEG=ON -DBUILD_PROTOBUF=ON -DPROTOBUF_UPDATE_FILES=OFF \
    -DWITH_PNG=ON -DWITH_JPEG=ON -DWITH_TIFF=OFF -DWITH_WEBP=OFF -DWITH_OPENJPEG=OFF -DWITH_JASPER=OFF \
    -DWITH_OPENEXR=OFF -DWITH_AVIF=OFF -DWITH_SPNG=OFF -DWITH_IMGCODEC_HDR=OFF -DWITH_IMGCODEC_SUNRASTER=OFF \
    -DWITH_IMGCODEC_PXM=OFF -DWITH_IMGCODEC_PFM=OFF -DWITH_IMGCODEC_GIF=OFF \
    -DWITH_FFMPEG=OFF -DWITH_GSTREAMER=OFF -DWITH_V4L=OFF -DWITH_1394=OFF -DWITH_AVFOUNDATION=OFF -DWITH_OBSENSOR=OFF \
    -DWITH_GTK=OFF -DWITH_QT=OFF -DWITH_COCOA=OFF -DWITH_WAYLAND=OFF \
    -DWITH_OPENCL=OFF -DWITH_VA=OFF -DWITH_VA_INTEL=OFF -DWITH_CUDA=OFF -DWITH_VULKAN=OFF \
    -DWITH_IPP=OFF -DWITH_ITT=OFF -DWITH_TBB=OFF -DWITH_OPENMP=OFF -DWITH_EIGEN=OFF -DWITH_LAPACK=OFF \
    -DWITH_OPENVINO=OFF -DWITH_ADE=OFF -DWITH_VTK=OFF -DWITH_CAROTENE=OFF -DWITH_KLEIDICV=OFF -DWITH_FLATBUFFERS=OFF \
    -DWITH_PROTOBUF=ON -DOPENCV_DNN_OPENCL=OFF -DOPENCV_DNN_CUDA=OFF -DCV_TRACE=OFF
  cmake --build "$src/build"
  cmake --install "$src/build"
  merge_private_libs "$DEPS_DIR/lib/pkgconfig/opencv4.pc"
  rm -rf "$src"
  touch "$stamp"
}

merge_private_libs() {
  local pc="$1" tmp
  tmp="$(mktemp)"
  awk '
    /^Libs:/ { libs = $0; next }
    /^Libs.private:/ { sub(/^Libs.private:/, ""); private = $0; next }
    { print }
    END { print libs private }
  ' "$pc" > "$tmp"
  mv "$tmp" "$pc"
}

build_server() {
  log "Building $BIN_NAME"
  local tmpdir
  tmpdir="$(mktemp -d)"
  BUILD_OUTPUT="$tmpdir/$BIN_NAME"
  export PKG_CONFIG_PATH="$DEPS_DIR/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
  if [ "$PLATFORM" = "macos" ]; then
    export PKG_CONFIG_PATH="$PKG_CONFIG_PATH:$(brew --prefix ffmpeg)/lib/pkgconfig"
  fi
  (cd "$SERVER_DIR" && go mod download && CGO_ENABLED=1 go build -a -o "$BUILD_OUTPUT" ./cmd)
}

install_binary() {
  local bindir="$PREFIX/bin"
  local dest="$bindir/$BIN_NAME"
  local sudo_cmd=""
  mkdir -p "$bindir" 2>/dev/null || true
  if [ ! -w "$bindir" ]; then
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
