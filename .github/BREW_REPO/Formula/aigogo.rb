class Aigogo < Formula
  desc "Make packaging and distributing your AI agents a breeze"
  homepage "https://github.com/aupeachmo/aigogo"
  version "0.0.1"
  license "MPL-2.0"

  on_macos do
    on_intel do
      url "https://github.com/aupeachmo/aigogo/releases/download/v0.0.1/aigogo-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
    on_arm do
      url "https://github.com/aupeachmo/aigogo/releases/download/v0.0.1/aigogo-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/aupeachmo/aigogo/releases/download/v0.0.1/aigogo-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
    on_arm do
      url "https://github.com/aupeachmo/aigogo/releases/download/v0.0.1/aigogo-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    if OS.mac? && Hardware::CPU.arm?
      bin.install "aigogo-darwin-arm64" => "aigogo"
    elsif OS.mac? && Hardware::CPU.intel?
      bin.install "aigogo-darwin-amd64" => "aigogo"
    elsif OS.linux? && Hardware::CPU.arm?
      bin.install "aigogo-linux-arm64" => "aigogo"
    elsif OS.linux?
      bin.install "aigogo-linux-amd64" => "aigogo"
    end
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/aigogo version")
  end
end
