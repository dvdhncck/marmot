go install main/marmot.go

ends up in  /home/dave/projects/go/bin/marmot

sudo ln -s /home/dave/projects/go/bin/marmot /usr/local/bin/marmot to make it system wide



GRAMMA

marmot meta init   - create a sample metadata.json (with best guess of album name?)
       meta check  - validate the metajson.json

marmot ingest PATH [PATH]...

marmot db_to_xls FILE

marmot xls_to_db FILE


 