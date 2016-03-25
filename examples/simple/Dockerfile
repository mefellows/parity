FROM debian

RUN apt-get update && apt-get install -y python
WORKDIR /opt/test-app
EXPOSE 8080

CMD ["python", "-m", "SimpleHTTPServer", "8080"]
