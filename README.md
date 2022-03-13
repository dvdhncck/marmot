deployment
==============================================
cd ~/projects/go/src/davidhancock.com/marmot/
go install main/marmot.go

ends up in  /home/dave/projects/go/bin/marmot

sudo ln -s /home/dave/projects/go/bin/marmot /usr/local/bin/marmot to make it system wide


start during dev:
==============================================

 go run main/marmot.go -server

 curl localhost:8088 playlist?albumid=1023




Interesting test cases:
=============================================
q=487 - very limited mp3 tags
