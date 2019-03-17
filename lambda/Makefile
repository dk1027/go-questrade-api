build:
	GOOS=linux go build -o bin/lambda
	cd bin; zip handler.zip lambda

clean:
	rm -rf bin

deploy:
	aws lambda --region us-west-2 update-function-code \
	--function-name goQuestrade \
	--zip-file fileb://bin/handler.zip
