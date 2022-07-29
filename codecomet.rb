class Codecomet < Formula
  desc "CodeComet (mark I)"
  homepage "https://github.com/codecomet-io/installers"
  license "All rights reserved"
  head "https://github.com/codecomet-io/installers.git", branch: "master"

  depends_on "qemu"

  def install
    arch = "arm64"
    platform = "darwin"

    on_intel do
      arch = "amd64"
    end
    bin.install Dir["release/mark-I/darwin/#{arch}/bin/*"]
    share.install Dir["release/mark-I/darwin/#{arch}/share/*"]

  end

  def post_install
    # XXX homebrew does sandbox installation, meaning we cannot touch plist directly...
    # system "touch", "/Users/dmp/io.codecomet.runner.plist"
    # system bin/"codecomet-machine", "install"
  end

  def uninstall
    # system bin/"codecomet-machine", "uninstall"
  end
end