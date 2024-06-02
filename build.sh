cd example/apilambda
go build -o ../../bin/apihandler
cd ../snslambda
go build -o ../../bin/snshandler
cd ../sfnlambda
go build -o ../../bin/lambdahandler
cd ../cronlambda
go build -o ../../bin/cronhandler
cd ../../