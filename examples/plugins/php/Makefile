buildplugin: # used for docker only, temp til working ARM php bins
	docker build --platform linux/amd64 -t php-plugin-test -f Dockerfile ../../..

testplugin:
	docker compose up