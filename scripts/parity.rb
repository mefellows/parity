require 'formula'

class Parity < Formula
  homepage "https://github.com/mefellows/parity"
  version "0.0.1"

  if Hardware.is_64_bit?
    url "https://github.com/mefellows/parity/releases/download/#{version}/darwin_amd64.zip"
    sha1 'ab7c3e649d8ad45f504e2bc5c9d25190a018e383'
  else
    url "https://github.com/mefellows/parity/releases/download/#{version}/darwin_386.zip"
    sha1 '57ff641c30bc2f2bec900abf8f5ab3f804ab8202'
  end

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do
    assert_equal "#{version}", `#{bin}/parity version`.strip
  end
end
