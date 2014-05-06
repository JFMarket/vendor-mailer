build:
	( test -d dist || mkdir dist ) && cd dist && gox ../

Readme.txt: Readme.md
	pandoc Readme.md -o Readme.txt

vendor-mailer.zip: build Readme.txt
	zip -r vendor-mailer.zip dist/vendor-mailer_windows_amd64.exe vendor-emails-example.csv Readme.txt