class Aigg < Formula
  desc "Make packaging and distributing your AI agents a breeze"
  homepage "https://github.com/aupeachmo/aigogo"
  version "0.0.1"
  license "MPL-2.0"

  on_macos do
    on_intel do
      url "https://github.com/aupeachmo/aigogo/releases/download/v0.0.1/aigg-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
    on_arm do
      url "https://github.com/aupeachmo/aigogo/releases/download/v0.0.1/aigg-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/aupeachmo/aigogo/releases/download/v0.0.1/aigg-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
    on_arm do
      url "https://github.com/aupeachmo/aigogo/releases/download/v0.0.1/aigg-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    if OS.mac? && Hardware::CPU.arm?
      bin.install "aigg-darwin-arm64" => "aigg"
    elsif OS.mac? && Hardware::CPU.intel?
      bin.install "aigg-darwin-amd64" => "aigg"
    elsif OS.linux? && Hardware::CPU.arm?
      bin.install "aigg-linux-arm64" => "aigg"
    elsif OS.linux?
      bin.install "aigg-linux-amd64" => "aigg"
    end
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/aigg version")
  end
end
