require 'formula'

class Parity < Formula
  homepage "https://github.com/mefellows/parity"
  version "0.0.1"

  if Hardware.is_64_bit?
    url "https://github.com/mefellows/parity/releases/download/v#{version}/darwin_amd64.zip"
    sha1 '411503eaef44cd14896564495c94c332fda82892'
  else
    url "https://github.com/mefellows/parity/releases/download/v#{version}/darwin_386.zip"
    sha1 'd1a8c956b629ccbaee9b24e44c35747df747f8a8'
  end

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do
    assert_equal "#{version}", `#{bin}/parity version`.strip
  end
end
