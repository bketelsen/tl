#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:?VERSION is required}"
DIST_DIR="${DIST_DIR:-dist}"
TAP_DIR="${TAP_DIR:-homebrew-tap}"
REPO="${REPO:-aholbreich/tl}"

required_archives=(
  "tl-darwin-amd64.tar.gz"
  "tl-darwin-arm64.tar.gz"
  "tl-linux-amd64.tar.gz"
  "tl-linux-arm64.tar.gz"
)

for archive in "${required_archives[@]}"; do
  if [[ ! -f "${DIST_DIR}/${archive}" ]]; then
    echo "missing release archive: ${DIST_DIR}/${archive}" >&2
    exit 1
  fi
done

sha256() {
  sha256sum "$1" | awk '{print $1}'
}

DARWIN_AMD64_SHA="$(sha256 "${DIST_DIR}/tl-darwin-amd64.tar.gz")"
DARWIN_ARM64_SHA="$(sha256 "${DIST_DIR}/tl-darwin-arm64.tar.gz")"
LINUX_AMD64_SHA="$(sha256 "${DIST_DIR}/tl-linux-amd64.tar.gz")"
LINUX_ARM64_SHA="$(sha256 "${DIST_DIR}/tl-linux-arm64.tar.gz")"

mkdir -p "${TAP_DIR}/Formula"

cat > "${TAP_DIR}/Formula/tl.rb" <<RUBY
# Homebrew formula for tl.
class Tl < Formula
  desc "Git-native task ledger for human and AI agent coordination"
  homepage "https://github.com/${REPO}"
  version "${VERSION}"
  license "MIT"

  head "https://github.com/${REPO}.git", branch: "main"

  livecheck do
    url :stable
    regex(/^v?(\\d+(?:\\.\\d+)+)$/i)
  end

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/${REPO}/releases/download/#{version}/tl-darwin-amd64.tar.gz"
      sha256 "${DARWIN_AMD64_SHA}"
    else
      url "https://github.com/${REPO}/releases/download/#{version}/tl-darwin-arm64.tar.gz"
      sha256 "${DARWIN_ARM64_SHA}"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/${REPO}/releases/download/#{version}/tl-linux-amd64.tar.gz"
      sha256 "${LINUX_AMD64_SHA}"
    else
      url "https://github.com/${REPO}/releases/download/#{version}/tl-linux-arm64.tar.gz"
      sha256 "${LINUX_ARM64_SHA}"
    end
  end

  depends_on "go" => :build if build.head?

  def install
    if build.head?
      system "go", "build", "-o", bin/"tl", "-ldflags", "-s -w -X main.version=HEAD", "."
    else
      bin.install "tl"
    end
  end

  test do
    assert_match "tl version", shell_output("#{bin}/tl --version")
  end
end
RUBY

echo "Updated ${TAP_DIR}/Formula/tl.rb for tl ${VERSION}"
