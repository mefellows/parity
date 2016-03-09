require 'formula'

class Muxy < Formula
  homepage "https://github.com/mefellows/parity"
  version "0.0.1"

  if Hardware.is_64_bit?
    url "https://github.com/mefellows/parity/releases/download/v#{version}/darwin_amd64.zip"
    sha1 '68437c4bcdd0b5c5c0937758c80b518ff80a2cea'
  else
    url "https://github.com/mefellows/parity/releases/download/v#{version}/darwin_386.zip"
    sha1 '721d8590108d381555f1dd576d42232637971a4a'
  end

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do

  end
end
