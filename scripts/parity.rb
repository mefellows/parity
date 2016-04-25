require 'formula'

class Parity < Formula
  homepage "https://github.com/mefellows/parity"
  version "0.0.1"

  if Hardware.is_64_bit?
    url "https://github.com/mefellows/parity/releases/download/#{version}/darwin_amd64.zip"
    sha1 'a63cbb1431cf9a1b48e01834361875b919280dcf'
  else
    url "https://github.com/mefellows/parity/releases/download/#{version}/darwin_386.zip"
    sha1 'f9802545b0be43990d99ffabd19f74e12c923310'
  end

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do
    assert_equal "#{version}", `#{bin}/parity version`.strip
  end
end
