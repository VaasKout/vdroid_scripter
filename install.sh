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
FFMPEG_VERSION="9.0.2"
LEPTONICA_VERSION="1.87.0"
TESSERACT_VERSION="5.5.3"
TESSDATA_LANGS="${VDROID_TESSDATA_LANGS:-eng rus}"
TESSDATA_URL="https://github.com/tesseract-ocr/tessdata/raw/main"

main() {
  detect_platform
  check_prefix
  install_dependencies
  require_go
  build_opencv
  build_ffmpeg
  build_leptonica
  build_tesseract
  build_server
  install_binary
  install_tessdata
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

check_prefix() {
  case "$PREFIX" in
    *" "*) die "PREFIX must not contain spaces: $PREFIX" ;;
  esac
}

tessdata_prefix() {
  printf '%s/share/%s' "$PREFIX" "vdroid_scripter"
}

job_count() {
  nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4
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
  if [ -n "${VDROID_SKIP_PACKAGES:-}" ]; then
    log "Skipping package installation (VDROID_SKIP_PACKAGES is set)"
    return
  fi
  log "Installing the build toolchain and ADB"
  if [ "$PLATFORM" = "macos" ]; then
    install_macos
  else
    install_linux
  fi
}

install_macos() {
  need_cmd brew
  brew install go android-platform-tools pkgconf cmake ninja nasm
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
    *) die "unsupported Linux distribution: ${DISTRO_ID:-unknown}. Install Go, adb, pkg-config, cmake, ninja, make, nasm, curl and a C++ compiler manually, then run: VDROID_SKIP_PACKAGES=1 ./install.sh" ;;
  esac
}

install_arch() {
  sudo pacman -S --needed --noconfirm go android-tools pkgconf gcc make cmake ninja nasm curl
}

install_apt() {
  sudo apt-get update
  sudo apt-get install -y golang-go adb pkg-config build-essential cmake ninja-build nasm curl
}

install_dnf() {
  sudo dnf install -y golang android-tools pkgconf-pkg-config gcc-c++ make cmake ninja-build nasm curl
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

dependency_built() {
  local name="$1" version="$2" pc="$3" detail="${4:-}"
  local stamp="$DEPS_DIR/stamps/$name-$version"
  [ -f "$stamp" ] && [ -f "$DEPS_DIR/lib/pkgconfig/$pc" ] && [ "$(cat "$stamp")" = "$detail" ]
}

mark_built() {
  local name="$1" version="$2" detail="${3:-}"
  mkdir -p "$DEPS_DIR/stamps"
  printf '%s' "$detail" > "$DEPS_DIR/stamps/$name-$version"
}

fetch_source() {
  local url="$1" tarball="$2" dir="$3"
  if [ -d "$dir" ]; then
    return
  fi
  need_cmd curl
  mkdir -p "$DEPS_DIR/src"
  log "Downloading $(basename "$tarball")"
  curl -fL "$url" -o "$tarball"
  tar -xzf "$tarball" -C "$DEPS_DIR/src"
}

merge_private_libs() {
  local pc="$1" tmp
  tmp="$(mktemp)"
  awk '
    function trim(text) { sub(/^[ \t]+/, "", text); sub(/[ \t,]+$/, "", text); return text }
    /^Libs:/ { libs = trim(substr($0, 6)); has_libs = 1; next }
    /^Libs.private:/ { part = trim(substr($0, 14)); if (part != "") libs = libs " " part; has_libs = 1; next }
    /^Requires:/ { requires = trim(substr($0, 10)); has_requires = 1; next }
    /^Requires.private:/ {
      part = trim(substr($0, 18))
      if (part != "" && requires != "") requires = requires ", "
      requires = requires part
      has_requires = 1
      next
    }
    { print }
    END {
      if (has_requires && requires != "") print "Requires: " requires
      if (has_libs) print "Libs: " libs
    }
  ' "$pc" > "$tmp"
  mv "$tmp" "$pc"
}

build_opencv() {
  if dependency_built opencv "$OPENCV_VERSION" opencv4.pc; then
    log "OpenCV $OPENCV_VERSION is already built in $DEPS_DIR"
    return
  fi
  need_cmd cmake
  need_cmd ninja
  local src="$DEPS_DIR/src/opencv-$OPENCV_VERSION"
  fetch_source "https://github.com/opencv/opencv/archive/refs/tags/$OPENCV_VERSION.tar.gz" "$DEPS_DIR/src/opencv-$OPENCV_VERSION.tar.gz" "$src"
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
  mark_built opencv "$OPENCV_VERSION"
}

build_ffmpeg() {
  if dependency_built ffmpeg "$FFMPEG_VERSION" libavcodec.pc; then
    log "FFmpeg $FFMPEG_VERSION is already built in $DEPS_DIR"
    return
  fi
  need_cmd make
  local src="$DEPS_DIR/src/ffmpeg-$FFMPEG_VERSION"
  fetch_source "https://ffmpeg.org/releases/ffmpeg-$FFMPEG_VERSION.tar.gz" "$DEPS_DIR/src/ffmpeg-$FFMPEG_VERSION.tar.gz" "$src"
  log "Building FFmpeg $FFMPEG_VERSION (libavcodec with only the H.264 decoder) as static libraries in $DEPS_DIR"
  (
    cd "$src"
    ./configure --prefix="$DEPS_DIR" --disable-shared --enable-static --enable-pic \
      --disable-everything --enable-decoder=h264 --enable-parser=h264 \
      --disable-programs --disable-doc --disable-network --disable-autodetect --disable-debug \
      --disable-avdevice --disable-avformat --disable-avfilter --disable-swscale
    make -j"$(job_count)"
    make install
  )
  merge_private_libs "$DEPS_DIR/lib/pkgconfig/libavcodec.pc"
  merge_private_libs "$DEPS_DIR/lib/pkgconfig/libavutil.pc"
  merge_private_libs "$DEPS_DIR/lib/pkgconfig/libswresample.pc"
  rm -rf "$src"
  mark_built ffmpeg "$FFMPEG_VERSION"
}

build_leptonica() {
  if dependency_built leptonica "$LEPTONICA_VERSION" lept.pc; then
    log "Leptonica $LEPTONICA_VERSION is already built in $DEPS_DIR"
    return
  fi
  need_cmd cmake
  need_cmd ninja
  local src="$DEPS_DIR/src/leptonica-$LEPTONICA_VERSION"
  fetch_source "https://github.com/DanBloomberg/leptonica/releases/download/$LEPTONICA_VERSION/leptonica-$LEPTONICA_VERSION.tar.gz" "$DEPS_DIR/src/leptonica-$LEPTONICA_VERSION.tar.gz" "$src"
  log "Building Leptonica $LEPTONICA_VERSION as a static library in $DEPS_DIR"
  cmake -S "$src" -B "$src/build" -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX="$DEPS_DIR" \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
    -DBUILD_SHARED_LIBS=OFF -DBUILD_PROG=OFF -DSW_BUILD=OFF \
    -DENABLE_ZLIB=OFF -DENABLE_PNG=OFF -DENABLE_GIF=OFF -DENABLE_JPEG=OFF -DENABLE_TIFF=OFF -DENABLE_WEBP=OFF -DENABLE_OPENJPEG=OFF
  cmake --build "$src/build"
  cmake --install "$src/build"
  merge_private_libs "$DEPS_DIR/lib/pkgconfig/lept.pc"
  rm -rf "$src"
  mark_built leptonica "$LEPTONICA_VERSION"
}

build_tesseract() {
  if dependency_built tesseract "$TESSERACT_VERSION" tesseract.pc "$(tessdata_prefix)"; then
    log "Tesseract $TESSERACT_VERSION is already built in $DEPS_DIR"
    return
  fi
  need_cmd cmake
  need_cmd ninja
  local src="$DEPS_DIR/src/tesseract-$TESSERACT_VERSION"
  fetch_source "https://github.com/tesseract-ocr/tesseract/archive/refs/tags/$TESSERACT_VERSION.tar.gz" "$DEPS_DIR/src/tesseract-$TESSERACT_VERSION.tar.gz" "$src"
  log "Building Tesseract $TESSERACT_VERSION as a static library in $DEPS_DIR (language files under $(tessdata_prefix)/tessdata)"
  PKG_CONFIG_PATH="$DEPS_DIR/lib/pkgconfig" cmake -S "$src" -B "$src/build" -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX="$DEPS_DIR" \
    -DCMAKE_CXX_FLAGS="-DTESSDATA_PREFIX=\\\"$(tessdata_prefix)\\\"" \
    -DCMAKE_PREFIX_PATH="$DEPS_DIR" \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
    -DBUILD_SHARED_LIBS=OFF -DBUILD_TRAINING_TOOLS=OFF -DBUILD_TESTS=OFF -DSW_BUILD=OFF \
    -DDISABLE_ARCHIVE=ON -DDISABLE_CURL=ON -DDISABLE_TIFF=ON -DGRAPHICS_DISABLED=ON \
    -DOPENMP_BUILD=OFF -DENABLE_LTO=OFF -DENABLE_OPENCL=OFF
  cmake --build "$src/build"
  cmake --install "$src/build"
  merge_private_libs "$DEPS_DIR/lib/pkgconfig/tesseract.pc"
  rm -rf "$src"
  mark_built tesseract "$TESSERACT_VERSION" "$(tessdata_prefix)"
}

build_server() {
  log "Building $BIN_NAME"
  local tmpdir
  tmpdir="$(mktemp -d)"
  BUILD_OUTPUT="$tmpdir/$BIN_NAME"
  export PKG_CONFIG_PATH="$DEPS_DIR/lib/pkgconfig"
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

install_tessdata() {
  local dir="$(tessdata_prefix)/tessdata"
  local sudo_cmd="" lang tmp
  mkdir -p "$dir" 2>/dev/null || true
  if [ ! -w "$dir" ]; then
    sudo_cmd="sudo"
  fi
  $sudo_cmd mkdir -p "$dir"
  for lang in $TESSDATA_LANGS; do
    if [ -f "$dir/$lang.traineddata" ]; then
      continue
    fi
    log "Downloading Tesseract language data: $lang"
    tmp="$(mktemp)"
    curl -fL "$TESSDATA_URL/$lang.traineddata" -o "$tmp"
    $sudo_cmd install -m 0644 "$tmp" "$dir/$lang.traineddata"
    rm -f "$tmp"
  done
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
