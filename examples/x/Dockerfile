FROM debian:latest

# See: https://github.com/jfrazelle/dockerfiles/blob/master/chrome/stable/Dockerfile

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y install x11-common Xorg wget

RUN mkdir /src/

RUN wget https://dl.google.com/linux/direct/google-talkplugin_current_amd64.deb -O /src/google-talkplugin_current_amd64.deb

RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb -O /src/google-chrome-stable_current_amd64.deb

# Install Chromium
RUN apt-get update --fix-missing

RUN echo 'deb http://httpredir.debian.org/debian testing main' >> /etc/apt/sources.list && \
	apt-get update && apt-get install -y \
	ca-certificates \
	fonts-liberation \
	fonts-roboto \
	gconf-service \
	hicolor-icon-theme \
	libappindicator1 \
	libasound2 \
	libcanberra-gtk-module \
	libcurl3 \
	libexif-dev \
	libfontconfig1 \
	libfreetype6 \
	libgconf-2-4 \
	libgl1-mesa-dri \
	libgl1-mesa-glx \
	libnspr4 \
	libnss3 \
	libpango1.0-0 \
	libv4l-0 \
	libxss1 \
	libxtst6 \
	lsb-base \
	strace \
	wget \
	xdg-utils \
	-t testing \
	fonts-symbola \
	--no-install-recommends && \
	dpkg -i '/src/google-chrome-stable_current_amd64.deb' && \
	dpkg -i '/src/google-talkplugin_current_amd64.deb' \
	&& rm -rf /var/lib/apt/lists/* \
	&& rm -rf /src/*.deb

# Google Chrome hack. See https://github.com/travis-ci/travis-ci/issues/938#issuecomment-77785455
RUN chmod 1777 /dev/shm

CMD ["bash"]
