require 'formula'

class Muxy < Formula
  homepage "https://github.com/mefellows/parity"
  version "0.0.1"

  if Hardware.is_64_bit?
    url "https://github.com/mefellows/parity/releases/download/v#{version}/darwin_amd64.zip"
    sha1 '9c435e627e649ef68b97dc9206'
  else
    url "https://github.com/mefellows/parity/releases/download/v#{version}/darwin_386.zip"
    sha1 '805d2aee7363a413'
  end

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do

  end
end
