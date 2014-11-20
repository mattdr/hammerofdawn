#
# Simple dockerfile for building curl-loader
#
FROM ubuntu:14.04

MAINTAINER Laramie Leavitt <laramie.leavitt@gmail.com>

RUN apt-get update
RUN apt-get -yq install \
  build-essential \
  libncurses5-dev \
  libssl-dev

ADD http://downloads.sourceforge.net/project/curl-loader/curl-loader-stable/curl-loader-0.56/curl-loader-0.56.tar.bz2 /usr/local/src/curl-loader-0.56.tar.bz2
RUN tar -C /usr/local/src -xjvf /usr/local/src/curl-loader-0.56.tar.bz2
WORKDIR /usr/local/src/curl-loader-0.56
RUN make cleanall
RUN make debug=0 optimize=1
RUN make install
RUN make clean

ENTRYPOINT ["/usr/bin/curl-loader"]
