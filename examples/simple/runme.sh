#!/bin/bash -e
echo "Before running this, make sure you have run 'parity run' first!"
echo ""
echo ""

echo 'This is a new file' > test.txt
docker build -t mirror-test .
docker run -d --name mirror-test -v $PWD:/opt/test-app -p 8080:8080 mirror-test
sleep 2
curl docker:8080/test.txt
echo 'This is an UPDATED file' > test.txt
curl docker:8080/test.txt
docker rm -f mirror-test
