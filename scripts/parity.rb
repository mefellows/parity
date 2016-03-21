require 'formula'

class Parity < Formula
  homepage "https://github.com/mefellows/parity"
  version "0.0.1"

  if Hardware.is_64_bit?
    url "https://github.com/mefellows/parity/releases/download/#{version}/darwin_amd64.zip"
    sha1 'd8cd1e5a3c10ed3f11578b6b8dd80a1d2e34dfb3'
  else
    url "https://github.com/mefellows/parity/releases/download/#{version}/darwin_386.zip"
    sha1 '867d36638b65971e44536ec754b607d22382930e'
  end

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do
    assert_equal "#{version}", `#{bin}/parity version`.strip
  end
end
