require 'formula'

class Parity < Formula
  homepage "https://github.com/mefellows/parity"
  version "0.0.1"

  if Hardware.is_64_bit?
    url "https://github.com/mefellows/parity/releases/download/#{version}/darwin_amd64.zip"
    sha1 ''
  else
    url "https://github.com/mefellows/parity/releases/download/#{version}/darwin_386.zip"
    sha1 ''
  end

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do
    assert_equal "#{version}", `#{bin}/parity version`.strip
  end
end
