.DEFAULT_GOAL := publish
.PHONY: publish build tag upload git-push prepare
.SILENT:

# publish set mp3 tags and deploy podcast file to the server and push changes to github
# usage: make publish /foo/bar/podcast/ump_podcast123.mp3
publish: tag upload git-push

prepare:
		@docker rm -f uwp_publisher_tmp
		docker run -d --rm --name=uwp_publisher_tmp uwp-publisher /srv/publisher sleepy --duration=1h
		docker cp uwp_publisher_tmp:/srv/publisher.osx /tmp/uwp-publisher
		/tmp/uwp-publisher prep --location=/Users/umputun/dev.umputun/podcast-uwp/hugo/content/posts
		@docker rm -f uwp_publisher_tmp

build:
		cd build && docker build -t uwp-publisher .

tag:
		# upload the file to the server
		src=$(word 2,$(MAKECMDGOALS)); \
  		fname=$$(basename $$src); \
		echo "src=$$src, fname=$$fname"; \
		docker run -it --rm -e APP_UID=502 -v $$src:/srv/$$fname -v /Users/umputun/.ssh/id_rsa:/home/app/.ssh/id_rsa uwp-publisher \
		/srv/publisher mp3 --file=/srv/$$fname

upload:
		# deploy the file to the podcast server
		src=$(word 2,$(MAKECMDGOALS)); \
		fname=$$(basename $$src); \
		echo "src=$$src, fname=$$fname"; \
		docker run -it --rm -e APP_UID=502 -v $$src:/srv/$$fname -v /Users/umputun/.ssh/id_rsa:/home/app/.ssh/id_rsa uwp-publisher \
			/srv/publisher deploy --file=/srv/$$fname --days-keep=750

git-push:
		# push changes to github
		git pull && git commit -am "new podcast $$fname" && git push
