class Pik < Formula
  desc "Lightweight Git GUI with line-level staging"
  homepage "https://github.com/t4traw/pik"
  version "0.1.0"
  url "https://github.com/t4traw/pik/releases/download/v#{version}/pik-v#{version}-darwin-universal.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  license "MIT"

  depends_on :macos

  def install
    # The tarball contains `pik.app/`. Keep it as a proper .app bundle under
    # the formula's prefix so macOS code signing stays intact.
    prefix.install "pik.app"

    # Expose a `pik` CLI that exec's the Mach-O inside the installed bundle.
    # Using opt_prefix keeps the symlink stable across version bumps.
    (bin/"pik").write <<~SH
      #!/bin/sh
      exec "#{opt_prefix}/pik.app/Contents/MacOS/pik" "$@"
    SH
    (bin/"pik").chmod 0755
  end

  test do
    assert_predicate prefix/"pik.app/Contents/MacOS/pik", :executable?
  end
end
