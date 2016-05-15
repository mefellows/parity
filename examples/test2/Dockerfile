FROM rails:4.2.5

#
# Base Docker Image
#
# Base Docker image that all other containers (Agent, Distribution) inherit from.
# Any changes that effect layers in this will result in re-builds of those further
# down the chain, most notably, changes to Gemfiles.

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY Gemfile /usr/src/app/
COPY Gemfile.lock /usr/src/app/
RUN bundle install
COPY . /usr/src/app
RUN echo "something elseaoeu and more"

CMD ["rails", "server", "-p", "80", "-b", "0.0.0.0"]
